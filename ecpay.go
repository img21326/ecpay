package ecpay_shipping

import (
	"fmt"
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

type EcpayShipping interface {
	ChooseShipStore(config ChooseShipStoreConfig) (string, error)
	CreateShipOrder(config CreateShippingOrderConfig) (string, error)
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
	resp, err := e.client.R().SetFormData(map[string]string{
		"MerchantID":       e.MerchantID,
		"MerchantTradeNo":  config.MerchantTradeNo,
		"LogisticsType":    "CVS",
		"LogisticsSubType": FormatStoreType(config.ShippingStoreType),
		"IsCollection":     FormatNeedPayment(config.NeedPayment),
		"ServerReplyURL":   config.ServerReplyURL,
		"ExtraData":        "",
		"Device":           FormatIsMobile(config.IsMobile),
	}).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Cache-Control", "no-cache").
		Post(fmt.Sprintf("%v/map", e.getURL()))

	if err != nil {
		return "", err
	}

	if resp.IsErrorState() {
		return "", fmt.Errorf("http status code %v", resp.StatusCode)
	}
	return resp.String(), nil
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
	return resp.String(), err
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
