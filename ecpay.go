package ecpay_shipping

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

const stagingURL = "https://logistics-stage.ecpay.com.tw/Express"
const productionURL = "https://logistics.ecpay.com.tw/Express"

type EcpayConfig struct {
	MerchantID   string
	HashKey      string
	HashIV       string
	IsProduction bool

	SenderName     string
	SenderPhone    string
	ServerReplyURL string
}

type ChooseShipStoreConfig struct {
	MerchantTradeNo   string
	ShippingStoreType string
	NeedPayment       bool
	ServerReplyURL    string
	IsMobile          bool
	Extra             interface{}
}

type ChooseShipStoreResponse struct {
	MerchantTradeNo        string `json:"MerchantID"`
	ShippingStoreType      string `json:"LogisticsSubType"`
	ShippingStoreID        string `json:"CVSStoreID"`
	ShippingStoreName      string `json:"CVSStoreName"`
	ShippingStoreAddress   string `json:"CVSAddress"`
	ShippingStoreTel       string `json:"CVSTelephone"`
	IsShippingStoreOutside bool   `json:"CVSOutSide"`
}

type CreateShippingOrderConfig struct {
	MerchantTradeNo   string
	TradeDate         time.Time
	ShippingStoreType string
	Amount            float32
	NeedPayment       bool
	EntreeName        string
	ReceiverName      string
	ReceiverPhone     string
	ReceiverEmail     string
	ReceiverStoreID   string
	ClientReplyURL    string
}

type ShipOrderResponse struct {
	MerchantID        string
	MerchantTradeNo   string
	RtnCode           string
	RtnMsg            string
	LogisticsID       string
	ShippingStoreType string
	GoodsAmount       string
	UpdateStatusDate  string
	CSVNo             string
	Status            string
}

type EcpayShipping interface {
	ChooseShipStore(config ChooseShipStoreConfig) (string, error)
	CreateShipOrder(config CreateShippingOrderConfig) (string, error)
	ParseShipOrderResponse(resp string) (ShipOrderResponse, error)
}

type EcpayShippingImpl struct {
	EcpayConfig
	client *req.Client
}

func NewEcpayShipping(config EcpayConfig) EcpayShipping {
	client := req.C().SetTimeout(30 * time.Second)

	return &EcpayShippingImpl{
		EcpayConfig: config,
		client:      client,
	}
}

func (e *EcpayShippingImpl) getURL() string {
	if e.IsProduction {
		return productionURL
	}
	return stagingURL
}

func (e *EcpayShippingImpl) ChooseShipStore(config ChooseShipStoreConfig) (string, error) {
	postData := map[string]string{
		"MerchantID":       e.MerchantID,
		"MerchantTradeNo":  config.MerchantTradeNo,
		"LogisticsType":    "CVS",
		"LogisticsSubType": FormatStoreType(config.ShippingStoreType),
		"IsCollection":     FormatNeedPayment(config.NeedPayment),
		"ServerReplyURL":   config.ServerReplyURL,
		"ExtraData":        fmt.Sprintf("%v", config.Extra),
		"Device":           FormatIsMobile(config.IsMobile),
	}

	postDataHtml := ""
	for key, value := range postData {
		postDataHtml += fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, key, value)
	}
	url := fmt.Sprintf("%s/map", e.getURL())

	html := fmt.Sprintf(`
		<!DOCTYPE html>
					<html>
					<head>
						<title></title>
					</head>
					<body>
						<form id="myForm" method="POST" action="%s">
							%s
						</form>

						<script>
							// Automatically submit the form when the page loads
							document.addEventListener("DOMContentLoaded", function() {
								document.getElementById("myForm").submit();
							});
						</script>
					</body>
					</html>
	`, url, postDataHtml)

	return html, nil
}

func (e *EcpayShippingImpl) CreateShipOrder(config CreateShippingOrderConfig) (string, error) {
	params := map[string]string{
		"MerchantID":        e.MerchantID,
		"MerchantTradeNo":   config.MerchantTradeNo,
		"MerchantTradeDate": config.TradeDate.Format("2006/01/02 15:04:05"),
		"LogisticsType":     "CVS",
		"LogisticsSubType":  FormatStoreType(config.ShippingStoreType),
		"GoodsAmount":       fmt.Sprintf("%.0f", config.Amount),
		"CollectionAmount":  fmt.Sprintf("%.0f", config.Amount),
		"IsCollection":      FormatNeedPayment(config.NeedPayment),
		"GoodsName":         config.EntreeName,
		"SenderName":        e.SenderName,
		"SenderPhone":       e.SenderPhone,
		"SenderCellPhone":   e.SenderPhone,
		"ReceiverName":      config.ReceiverName,
		"ReceiverPhone":     config.ReceiverPhone,
		"ReceiverCellPhone": config.ReceiverPhone,
		"ReceiverEmail":     config.ReceiverEmail,
		"ReceiverStoreID":   config.ReceiverStoreID,
		"ClientReplyURL":    config.ClientReplyURL,
		"ServerReplyURL":    e.ServerReplyURL,
		"PlatformID":        "",
	}
	checkMac := NewCheckMacValueService(e.EcpayConfig).GenerateCheckMacValue(params)
	params["CheckMacValue"] = checkMac

	resp, err := e.client.R().SetFormData(params).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Cache-Control", "no-cache").
		Post(fmt.Sprintf("%v/Create", e.getURL()))
	if err != nil {
		return "", err
	}
	respString := resp.String()
	param := strings.Split(respString, "|")
	if param[0] != "1" {
		return "", fmt.Errorf("response error status: %v, err: %v", param[0], param[1])
	}
	return param[1], nil
}

func (e *EcpayShippingImpl) ParseShipOrderResponse(resp string) (ShipOrderResponse, error) {
	var response ShipOrderResponse
	values, err := url.ParseQuery(resp)
	if err != nil {
		return response, err
	}
	response.MerchantID = values.Get("MerchantID")
	response.MerchantTradeNo = values.Get("MerchantTradeNo")
	response.RtnCode = values.Get("RtnCode")
	response.RtnMsg = values.Get("RtnMsg")
	response.LogisticsID = values.Get("AllPayLogisticsID")
	response.ShippingStoreType = TransferStoreType(values.Get("LogisticsSubType"))
	response.GoodsAmount = values.Get("GoodsAmount")
	response.UpdateStatusDate = values.Get("UpdateStatusDate")

	if response.ShippingStoreType == "711" {
		response.CSVNo = values.Get("CVSPaymentNo") + values.Get("CVSValidationNo")
	} else {
		response.CSVNo = values.Get("CVSPaymentNo")
	}
	response.Status = TransferStatus(response.ShippingStoreType, response.RtnCode)
	return response, err
}

func TransferStoreType(storeType string) string {
	switch storeType {
	case "FAMIC2C":
		return "FAMI"
	case "UNIMARTC2C":
		return "711"
	case "HILIFEC2C":
		return "HI-LIFI"
	case "OKMARTC2C":
		return "OK"
	default:
		return ""
	}
}

const CREATED = "CREATED"
const SELLER_SEND_TO_STORE = "SELLER_SEND_TO_STORE"
const DELIVERED = "DELIVERED"
const BUYER_PICK_UP = "BUYER_PICK_UP"
const BUYER_DIDNT_PICK_UP = "BUYER_DIDNT_PICK_UP"
const RETURN_TO_STORE = "RETURN_TO_STORE"
const RETURNED = "RETURNED"
const UNDEFINE = "UNDEFINE"

var SevenStatus = map[string]string{
	"300":  CREATED,
	"2068": SELLER_SEND_TO_STORE,
	"2030": SELLER_SEND_TO_STORE,
	"2073": DELIVERED,
	"2067": BUYER_PICK_UP,
	"2074": BUYER_DIDNT_PICK_UP,
	"2072": RETURN_TO_STORE,
	"2070": RETURNED,
}

var FamiStatus = map[string]string{
	"300":  CREATED,
	"3024": SELLER_SEND_TO_STORE,
	"3018": DELIVERED,
	"3022": BUYER_PICK_UP,
	"3020": BUYER_DIDNT_PICK_UP,
	"3019": RETURN_TO_STORE,
	"3023": RETURNED,
}

var HiLifeStatus = map[string]string{
	"300":  CREATED,
	"2030": SELLER_SEND_TO_STORE,
	"3024": SELLER_SEND_TO_STORE,
	"3032": SELLER_SEND_TO_STORE,
	"2063": DELIVERED,
	"3018": DELIVERED,
	"2067": BUYER_PICK_UP,
	"3022": BUYER_PICK_UP,
	"2074": BUYER_DIDNT_PICK_UP,
	"3020": BUYER_DIDNT_PICK_UP,
	"2072": RETURN_TO_STORE,
	"3019": RETURN_TO_STORE,
	"2070": RETURNED,
	"3023": RETURNED,
	"9001": RETURNED,
	"9002": RETURNED,
}

var OkStatus = map[string]string{
	"300":  CREATED,
	"2030": SELLER_SEND_TO_STORE,
	"2073": DELIVERED,
	"3022": BUYER_PICK_UP,
	"2074": BUYER_DIDNT_PICK_UP,
	"2072": RETURN_TO_STORE,
	"3023": RETURNED,
}

func TransferStatus(storeType string, rtnCode string) string {
	switch storeType {
	case "FAMI":
		if val, ok := FamiStatus[rtnCode]; ok {
			return val
		}
		return UNDEFINE
	case "711":
		if val, ok := SevenStatus[rtnCode]; ok {
			return val
		}
		return UNDEFINE
	case "HI-LIFI":
		if val, ok := HiLifeStatus[rtnCode]; ok {
			return val
		}
		return UNDEFINE
	case "OK":
		if val, ok := OkStatus[rtnCode]; ok {
			return val
		}
		return UNDEFINE
	default:
		return UNDEFINE
	}
}

func FormatStoreType(storeType string) string {
	switch storeType {
	case "FAMI":
		return "FAMIC2C"
	case "711":
		return "UNIMARTC2C"
	case "HI-LIFI":
		return "HILIFEC2C"
	case "OK":
		return "OKMARTC2C"
	default:
		return ""
	}
}

func FormatNeedPayment(needPayment bool) string {
	if needPayment {
		return "Y"
	}
	return "N"
}

func FormatIsMobile(isMobile bool) string {
	if isMobile {
		return "1"
	}
	return "0"
}
