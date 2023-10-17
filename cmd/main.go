package main

import (
	"fmt"
	"time"

	ecpayShipping "github.com/img21326/ecpay"
)

func main() {

	ec := ecpayShipping.NewEcpay(ecpayShipping.EcpayConfig{
		MerchantID:   "2000933",
		HashKey:      "XBERn1YOvpM9nfZc",
		HashIV:       "h1ONHk4P4yqbl5LK",
		IsProduction: false,

		SenderName:         "test",
		SenderPhone:        "0912345678",
		ShipServerReplyURL: "http://localhost:8000/test",
	})

	// resp, err := ec.ChooseShipStore(ecpayShipping.ChooseShipStoreConfig{
	// 	MerchantTradeNo:   "test123456789",
	// 	ShippingStoreType: "FAMI",
	// 	NeedPayment:       false,
	// 	ServerReplyURL:    "https://www.ecpay.com.tw/receive.php",
	// 	IsMobile:          false,
	// })
	resp, err := ec.CreateShipOrder(ecpayShipping.CreateShippingOrderConfig{
		MerchantTradeNo:   "4b4797b81715",
		TradeDate:         time.Now(),
		ShippingStoreType: "711",
		Amount:            150,
		NeedPayment:       true,
		EntreeName:        "name",
		ReceiverName:      "你好嗎",
		ReceiverPhone:     "0912345679",
		ReceiverEmail:     "test@gmail.com",
		ReceiverStoreID:   "131386",
		ClientReplyURL:    "",
	})
	if err != nil {
		fmt.Printf("get error: %+v\n", err)
		return
	}
	fmt.Printf("create string: %+v\n", resp)

	r, err := ec.ParseShipOrderResponse(resp)
	if err != nil {
		fmt.Printf("get parse error: %+v\n", err)
	}
	fmt.Printf("create parse: %+v\n", r)

	resp2, err := ec.QueryShip(ecpayShipping.QueryShipConfig{
		MerchantTradeNo: "4b4797b81715"})
	if err != nil {
		fmt.Printf("get error: %+v\n", err)
		return
	}
	fmt.Printf("query: %+v\n", resp2)

	// ec := ecpayShipping.NewEcpay(ecpayShipping.EcpayConfig{
	// 	MerchantID:   "3002607",
	// 	HashKey:      "pwFHCqoQZGmho4w6",
	// 	HashIV:       "EkRm7iFT261dpevs",
	// 	IsProduction: false,

	// 	SenderName:            "test",
	// 	SenderPhone:           "0912345678",
	// 	ShipServerReplyURL:    "http://localhost:8000/test",
	// 	PaymentServerReplyURL: "http://localhost:8000/test",
	// })

	// resp, err := ec.CreatePaymentOrder(ecpayShipping.PaymentConfig{
	// 	MerchantTradeNo: "test1ddss",
	// 	TradeDate:       time.Now(),
	// 	Amount:          150,
	// 	EntreeName:      "name",
	// 	SupportPayments: []string{"Credit", "WebATM"},
	// 	StoreID:         "131386",
	// 	ClientReplyURL:  "",
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("%+v\n", resp)
}
