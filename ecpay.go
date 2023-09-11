package ecpay

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

const shipStagingURL = "https://logistics-stage.ecpay.com.tw/Express"
const shipProductionURL = "https://logistics.ecpay.com.tw/Express"

const paymentStagingURL = "https://payment-stage.ecpay.com.tw/Cashier/AioCheckOut/V5"
const paymentProductionURL = "https://payment.ecpay.com.tw/Cashier/AioCheckOut/V5"

const creditPaymentStatusStagingURL = ""
const creditPaymentStatusProductionURL = "https://payment.ecpay.com.tw/CreditDetail/QueryTrade/V2"

type EcpayConfig struct {
	MerchantID     string
	HashKey        string
	HashIV         string
	CreditCheckKey string
	IsProduction   bool

	SenderName         string
	SenderPhone        string
	ShipServerReplyURL string

	PaymentServerReplyURL string
}

type ChooseShipStoreConfig struct {
	MerchantTradeNo   string
	ShippingStoreType string
	NeedPayment       bool
	ServerReplyURL    string
	IsMobile          bool
	Extra             string
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

type PaymentConfig struct {
	MerchantTradeNo string
	TradeDate       time.Time
	Amount          float32
	EntreeName      string
	SupportPayments []string
	StoreID         string
	ClientReplyURL  string
}

type CreditPaymentStatusConfig struct {
	MerchantTradeNo string
	RefundID        string
	TradeNo         string
	Amount          float32
}

type Ecpay interface {
	ChooseShipStore(config ChooseShipStoreConfig) (string, error)
	CreateShipOrder(config CreateShippingOrderConfig) (string, error)
	ParseShipOrderResponse(resp string) (ShipOrderResponse, error)

	CreatePaymentOrder(config PaymentConfig) (string, error)
	ParsePaymentResult(resp string) (*PaymentResponse, error)

	GetCreditPaymentStatus(config CreditPaymentStatusConfig) (*CreditPaymentStatusResponse, error)
}

type EcpayImpl struct {
	EcpayConfig
	client *req.Client
}

func NewEcpay(config EcpayConfig) Ecpay {
	client := req.C().SetTimeout(30 * time.Second)

	return &EcpayImpl{
		EcpayConfig: config,
		client:      client,
	}
}

func (e *EcpayImpl) getShipURL() string {
	if e.IsProduction {
		return shipProductionURL
	}
	return shipStagingURL
}

func (e *EcpayImpl) ChooseShipStore(config ChooseShipStoreConfig) (string, error) {
	postData := map[string]string{
		"MerchantID":       e.MerchantID,
		"MerchantTradeNo":  config.MerchantTradeNo,
		"LogisticsType":    "CVS",
		"LogisticsSubType": FormatStoreType(config.ShippingStoreType),
		"IsCollection":     FormatNeedPayment(config.NeedPayment),
		"ServerReplyURL":   config.ServerReplyURL,
		"ExtraData":        config.Extra,
		"Device":           FormatIsMobile(config.IsMobile),
	}

	postDataHtml := ""
	for key, value := range postData {
		postDataHtml += fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, key, value)
	}
	url := fmt.Sprintf("%s/map", e.getShipURL())

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

func (e *EcpayImpl) CreateShipOrder(config CreateShippingOrderConfig) (string, error) {
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
		"ServerReplyURL":    e.ShipServerReplyURL,
		"PlatformID":        "",
	}
	checkMac := NewShipMacValue(e.EcpayConfig).GenerateCheckMacValue(params)
	params["CheckMacValue"] = checkMac

	resp, err := e.client.R().SetFormData(params).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Cache-Control", "no-cache").
		Post(fmt.Sprintf("%v/Create", e.getShipURL()))
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

func (e *EcpayImpl) ParseShipOrderResponse(resp string) (ShipOrderResponse, error) {
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

func (e *EcpayImpl) getPaymentURL() string {
	if e.IsProduction {
		return paymentProductionURL
	}
	return paymentStagingURL
}

var supportPayments = []string{
	"Credit",
	"WebATM",
	"ATM",
	"CVS",
	"BARCODE",
	"TWQR",
}

func (e *EcpayImpl) CreatePaymentOrder(config PaymentConfig) (string, error) {
	params := map[string]string{
		"MerchantID":        e.MerchantID,
		"MerchantTradeNo":   config.MerchantTradeNo,
		"MerchantTradeDate": config.TradeDate.Format("2006/01/02 15:04:05"),
		"PaymentType":       "aio",
		"ChoosePayment":     "ALL",
		"TotalAmount":       fmt.Sprintf("%.0f", config.Amount),
		"TradeDesc":         config.EntreeName,
		"ItemName":          config.EntreeName,
		"ReturnURL":         e.PaymentServerReplyURL,
		"StoreID":           config.StoreID,
		"EncryptType":       "1",
		"ClientBackURL":     config.ClientReplyURL,
		"NeedExtraPaidInfo": "Y",
	}
	var ignorePayments []string
	for _, supportPayment := range supportPayments {
		var find = false
		for _, choosePayment := range config.SupportPayments {
			if choosePayment == supportPayment {
				find = true
				break
			}
		}
		if !find {
			ignorePayments = append(ignorePayments, supportPayment)
		}
	}
	params["IgnorePayment"] = strings.Join(ignorePayments, "#")

	checkMac := NewPaymentMacValue(e.EcpayConfig).GenerateCheckMacValue(params)
	params["CheckMacValue"] = checkMac

	postDataHtml := ""
	for key, value := range params {
		postDataHtml += fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, key, value)
	}
	url := e.getPaymentURL()
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

type PaymentResponse struct {
	MerchantID        string
	TradeNo           string
	StoreID           string
	RtnCode           string
	RtnMsg            string
	BankTransactionID string
	Amount            float64
	TradeDate         string
	PaymentType       string
	PaymentFee        float64
	PaymentDate       time.Time
	Simulation        bool
}

func (p *PaymentResponse) HasPaid() bool {
	return p.RtnCode == "1"
}

func (e *EcpayImpl) ParsePaymentResult(resp string) (*PaymentResponse, error) {
	values, err := url.ParseQuery(resp)
	if err != nil {
		return nil, err
	}
	var respMap = make(map[string]string)
	for key, value := range values {
		if key == "CheckMacValue" {
			continue
		}
		respMap[key] = value[0]
	}

	amount, _ := strconv.ParseFloat(respMap["TradeAmt"], 64)
	paymentFee, _ := strconv.ParseFloat(respMap["PaymentTypeChargeFee"], 64)
	simulation := respMap["SimulatePaid"] == "1"
	paymentDate, _ := time.Parse("2006/01/02 15:04:05", respMap["PaymentDate"])

	var response *PaymentResponse = &PaymentResponse{}
	response.MerchantID = respMap["MerchantID"]
	response.TradeNo = respMap["MerchantTradeNo"]
	response.StoreID = respMap["StoreID"]
	response.RtnCode = respMap["RtnCode"]
	response.RtnMsg = respMap["RtnMsg"]
	response.BankTransactionID = respMap["TradeNo"]
	response.Amount = amount
	response.TradeDate = respMap["TradeDate"]
	response.PaymentType = respMap["PaymentType"]
	response.PaymentFee = paymentFee
	response.PaymentDate = paymentDate
	response.Simulation = simulation
	return response, nil
}

type CreditPaymentStatusResponse struct {
	TradeID int     `json:"TradeID"`
	Amount  float32 `json:"Amount"`
	Status  string  `json:"Status"`
}

func (r *CreditPaymentStatusResponse) IsAuthorized() bool {
	return r.Status == "已授權"
}

func (r *CreditPaymentStatusResponse) IsPaid() bool {
	return r.Status == "已關帳"
}

func (r *CreditPaymentStatusResponse) IsRefunded() bool {
	return r.Status == "已取消"
}

func (e *EcpayImpl) GetCreditPaymentStatus(config CreditPaymentStatusConfig) (*CreditPaymentStatusResponse, error) {
	params := map[string]string{
		"MerchantID":      e.MerchantID,
		"CreditRefundId":  config.RefundID,
		"CreditAmount":    strconv.FormatFloat(float64(config.Amount), 'f', 0, 64),
		"CreditCheckCode": e.CreditCheckKey,
	}
	checkMac := NewShipMacValue(e.EcpayConfig).GenerateCheckMacValue(params)
	params["CheckMacValue"] = checkMac
	resp, err := e.client.R().SetFormData(params).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Cache-Control", "no-cache").
		Post(creditPaymentStatusProductionURL)
	if err != nil {
		return nil, err
	}

	var response *CreditPaymentStatusResponse
	values, _ := url.ParseQuery(resp.String())
	rtnMsg := values.Get("RtnMsg")
	rtnValue := values.Get("RtnValue")

	if rtnMsg != "" {
		return nil, fmt.Errorf("get credit payment status error: %v", rtnMsg)
	}
	err = json.Unmarshal([]byte(rtnValue), &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
