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

const shipStagingURL = "https://logistics-stage.ecpay.com.tw/"
const shipProductionURL = "https://logistics.ecpay.com.tw/"

const paymentStagingURL = "https://payment-stage.ecpay.com.tw"
const paymentProductionURL = "https://payment.ecpay.com.tw"

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

type QueryConfig struct {
	MerchantTradeNo string
}

type Ecpay interface {
	ChooseShipStore(config ChooseShipStoreConfig) (string, error)
	CreateShipOrder(config CreateShippingOrderConfig) (string, error)
	QueryShip(config QueryShipConfig) (ShipOrderResponse, error)
	ParseShipOrderResponse(resp string) (ShipOrderResponse, error)

	CreatePaymentOrder(config PaymentConfig) (string, error)
	ParsePaymentResult(resp string) (*PaymentResponse, error)

	QueryPayment(config QueryConfig) (*PaymentResponse, error)
	RefundPayment(config RefundConfig) (*RefundResponse, error)
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
	url := fmt.Sprintf("%s/Express/map", e.getShipURL())

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
		Post(fmt.Sprintf("%v/Express/Create", e.getShipURL()))
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

type QueryShipConfig struct {
	MerchantTradeNo string
}

func (e *EcpayImpl) QueryShip(config QueryShipConfig) (ShipOrderResponse, error) {
	params := map[string]string{
		"MerchantID":      e.MerchantID,
		"MerchantTradeNo": config.MerchantTradeNo,
		"TimeStamp":       fmt.Sprintf("%d", time.Now().Unix()),
	}
	checkMac := NewShipMacValue(e.EcpayConfig).GenerateCheckMacValue(params)
	params["CheckMacValue"] = checkMac
	resp, err := e.client.R().SetFormData(params).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Cache-Control", "no-cache").
		Post(fmt.Sprintf("%v/Helper/QueryLogisticsTradeInfo/V4", e.getShipURL()))
	if err != nil {
		return ShipOrderResponse{}, err
	}
	respString := resp.String()
	fmt.Printf("resp: %v\n", respString)

	var response ShipOrderResponse
	values, err := url.ParseQuery(respString)
	if err != nil {
		return response, err
	}
	response.MerchantID = values.Get("MerchantID")
	response.MerchantTradeNo = values.Get("MerchantTradeNo")
	response.RtnCode = values.Get("LogisticsStatus")
	response.RtnMsg = values.Get("RtnMsg")
	response.LogisticsID = values.Get("AllPayLogisticsID")
	response.ShippingStoreType = TransferStoreType(values.Get("LogisticsType"))
	response.GoodsAmount = values.Get("GoodsAmount")
	response.UpdateStatusDate = values.Get("TradeDate")

	if response.ShippingStoreType == "711" {
		response.CSVNo = values.Get("CVSPaymentNo") + values.Get("CVSValidationNo")
	} else {
		response.CSVNo = values.Get("CVSPaymentNo")
	}
	response.Status = TransferStatus(response.ShippingStoreType, response.RtnCode)
	return response, err
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
	`, fmt.Sprintf("%s/Cashier/AioCheckOut/V5", url), postDataHtml)

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

	// more info ..
	WebATMAccBank  string
	WebATMAccNo    string
	WebATMBankName string
	PaymentNo      string
	ProcessDate    string
	Card4No        string
	AuthCode       string
	RefundID       string

	// from where create
	By string
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
	response.By = "notify"

	if val, ok := respMap["WebATMAccBank"]; ok {
		response.WebATMAccBank = val
	}
	if val, ok := respMap["WebATMAccNo"]; ok {
		response.WebATMAccNo = val
	}
	if val, ok := respMap["WebATMBankName"]; ok {
		response.WebATMBankName = val
	}
	if val, ok := respMap["PaymentNo"]; ok {
		response.PaymentNo = val
	}
	if val, ok := respMap["process_date"]; ok {
		response.ProcessDate = val
	}
	if val, ok := respMap["card4no"]; ok {
		response.Card4No = val
	}
	if val, ok := respMap["auth_code"]; ok {
		response.AuthCode = val
	}
	if val, ok := respMap["gwsr"]; ok {
		response.RefundID = val
	}

	return response, nil
}

func (e *EcpayImpl) QueryPayment(config QueryConfig) (*PaymentResponse, error) {
	params := map[string]string{
		"MerchantID":      e.MerchantID,
		"MerchantTradeNo": config.MerchantTradeNo,
		"TimeStamp":       strconv.Itoa(int(time.Now().Unix())),
	}
	checkMac := NewPaymentMacValue(e.EcpayConfig).GenerateCheckMacValue(params)
	params["CheckMacValue"] = checkMac

	resp, err := e.client.R().SetFormData(params).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Cache-Control", "no-cache").
		Post(fmt.Sprintf("%s/Cashier/QueryTradeInfo/V5", e.getPaymentURL()))
	if err != nil {
		return nil, err
	}
	respString := resp.String()
	retParams, err := url.ParseQuery(respString)
	if err != nil {
		return nil, err
	}
	amount, _ := strconv.ParseFloat(retParams.Get("TradeAmt"), 64)

	var paymentResp *PaymentResponse = &PaymentResponse{}
	paymentResp.MerchantID = retParams.Get("MerchantID")
	paymentResp.TradeNo = retParams.Get("MerchantTradeNo")
	paymentResp.StoreID = retParams.Get("StoreID")
	paymentResp.BankTransactionID = retParams.Get("TradeNo")
	paymentResp.Amount = amount
	paymentResp.TradeDate = retParams.Get("TradeDate")
	paymentResp.PaymentType = retParams.Get("PaymentType")
	paymentResp.PaymentFee, _ = strconv.ParseFloat(retParams.Get("PaymentTypeChargeFee"), 64)
	paymentResp.PaymentDate, _ = time.Parse("2006/01/02 15:04:05", retParams.Get("PaymentDate"))
	paymentResp.RtnCode = retParams.Get("TradeStatus")
	paymentResp.By = "query"
	return paymentResp, nil
}

type RefundConfig struct {
	MerchantTradeNo   string
	BankTransactionID string
	RefundID          string
	Amount            float64
}

type RefundResponse struct {
	RtnCode string
	RtnMsg  string
}

func (r *RefundResponse) IsSuccess() bool {
	return r.RtnCode == "1"
}

func (e *EcpayImpl) RefundPayment(config RefundConfig) (*RefundResponse, error) {
	params := map[string]string{
		"MerchantID":      e.MerchantID,
		"CreditRefundId":  config.RefundID,
		"CreditAmount":    fmt.Sprintf("%.0f", config.Amount),
		"CreditCheckCode": e.CreditCheckKey,
	}
	checkMac := NewPaymentMacValue(e.EcpayConfig).GenerateCheckMacValue(params)
	params["CheckMacValue"] = checkMac

	resp, err := e.client.R().SetFormData(params).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Cache-Control", "no-cache").
		Post(fmt.Sprintf("%s/CreditDetail/QueryTrade/V2", e.getPaymentURL()))
	if err != nil {
		return nil, err
	}
	respString := resp.String()
	var result map[string]interface{} = make(map[string]interface{})
	err = json.Unmarshal([]byte(respString), &result)
	if err != nil {
		return nil, err
	}

	if _, ok := result["RtnCode"].(string); ok {
		return nil, fmt.Errorf("query credit card payment error: %v", result["RtnMsg"].(string))
	}

	rtnValue := result["RtnValue"].(map[string]interface{})
	rtnStatus := rtnValue["status"].(string)

	params = map[string]string{
		"MerchantID":      e.MerchantID,
		"MerchantTradeNo": config.MerchantTradeNo,
		"TradeNo":         config.BankTransactionID,
		"Action":          "",
		"TotalAmount":     fmt.Sprintf("%.0f", config.Amount),
	}

	var res *RefundResponse = &RefundResponse{}
	switch rtnStatus {
	case "已授權":
		params["Action"] = "N"
		checkMac := NewPaymentMacValue(e.EcpayConfig).GenerateCheckMacValue(params)
		params["CheckMacValue"] = checkMac

		resp, err := e.client.R().SetFormData(params).
			SetHeader("Content-Type", "application/x-www-form-urlencoded").
			SetHeader("Cache-Control", "no-cache").
			Post(fmt.Sprintf("%s/CreditDetail/DoAction", e.getPaymentURL()))
		if err != nil {
			return nil, err
		}
		respString := resp.String()
		retParams, err := url.ParseQuery(respString)
		if err != nil {
			return nil, err
		}
		res.RtnCode = retParams.Get("RtnCode")
		res.RtnMsg = retParams.Get("RtnMsg")
	case "已關帳":
		params["Action"] = "R"
		checkMac := NewPaymentMacValue(e.EcpayConfig).GenerateCheckMacValue(params)
		params["CheckMacValue"] = checkMac

		resp, err := e.client.R().SetFormData(params).
			SetHeader("Content-Type", "application/x-www-form-urlencoded").
			SetHeader("Cache-Control", "no-cache").
			Post(fmt.Sprintf("%s/CreditDetail/DoAction", e.getPaymentURL()))
		if err != nil {
			return nil, err
		}
		respString := resp.String()
		retParams, err := url.ParseQuery(respString)
		if err != nil {
			return nil, err
		}
		res.RtnCode = retParams.Get("RtnCode")
		res.RtnMsg = retParams.Get("RtnMsg")
	default:
		params["Action"] = "E"
		checkMac := NewPaymentMacValue(e.EcpayConfig).GenerateCheckMacValue(params)
		params["CheckMacValue"] = checkMac

		resp, err := e.client.R().SetFormData(params).
			SetHeader("Content-Type", "application/x-www-form-urlencoded").
			SetHeader("Cache-Control", "no-cache").
			Post(fmt.Sprintf("%s/CreditDetail/DoAction", e.getPaymentURL()))
		if err != nil {
			return nil, err
		}
		respString := resp.String()
		retParams, err := url.ParseQuery(respString)
		if err != nil {
			return nil, err
		}
		res.RtnCode = retParams.Get("RtnCode")
		res.RtnMsg = retParams.Get("RtnMsg")
		params["Action"] = "N"
		checkMac = NewPaymentMacValue(e.EcpayConfig).GenerateCheckMacValue(params)
		params["CheckMacValue"] = checkMac

		resp, err = e.client.R().SetFormData(params).
			SetHeader("Content-Type", "application/x-www-form-urlencoded").
			SetHeader("Cache-Control", "no-cache").
			Post(fmt.Sprintf("%s/CreditDetail/DoAction", e.getPaymentURL()))
		if err != nil {
			return nil, err
		}
		respString = resp.String()
		retParams, err = url.ParseQuery(respString)
		if err != nil {
			return nil, err
		}
		res.RtnCode = retParams.Get("RtnCode")
		res.RtnMsg = retParams.Get("RtnMsg")
	}
	if !res.IsSuccess() {
		return nil, fmt.Errorf("close credit card payment error: %v", res.RtnMsg)
	}
	return res, nil
}
