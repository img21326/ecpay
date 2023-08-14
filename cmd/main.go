package main

import (
	"fmt"
	"time"

	ecpayShipping "github.com/img21326/ecpay_shipping"
)

func main() {

	ec := ecpayShipping.NewEcpayShipping(ecpayShipping.EcpayConfig{
		MerchantID:   "2000933",
		HashKey:      "XBERn1YOvpM9nfZc",
		HashIV:       "h1ONHk4P4yqbl5LK",
		IsProduction: false,

		SenderName:     "test",
		SenderPhone:    "0912345678",
		ServerReplyURL: "http://localhost:8000/test",
	})

	// resp, err := ec.ChooseShipStore(ecpayShipping.ChooseShipStoreConfig{
	// 	MerchantTradeNo:   "test123456789",
	// 	ShippingStoreType: "FAMI",
	// 	NeedPayment:       false,
	// 	ServerReplyURL:    "https://www.ecpay.com.tw/receive.php",
	// 	IsMobile:          false,
	// })
	resp, err := ec.CreateShipOrder(ecpayShipping.CreateShippingOrderConfig{
		MerchantTradeNo:   "4b4797b817",
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
	fmt.Printf("%+v\n", resp)

	r, err := ec.ParseShipOrderResponse(resp)
	if err != nil {
		fmt.Printf("get parse error: %+v\n", err)
	}
	fmt.Printf("parse: %+v", r)
}
