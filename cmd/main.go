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

		SenderName:     "測試人員",
		SenderPhone:    "0912345678",
		ServerReplyURL: "https://www.ecpay.com.tw/receive.php",
	})
	// resp, err := ec.ChooseShipStore(ecpayShipping.ChooseShipStoreConfig{
	// 	MerchantTradeNo:   "test123456789",
	// 	ShippingStoreType: "FAMI",
	// 	NeedPayment:       false,
	// 	ServerReplyURL:    "https://www.ecpay.com.tw/receive.php",
	// 	IsMobile:          false,
	// })
	resp, err := ec.CreateShipOrder(ecpayShipping.CreateShippingOrderConfig{
		MerchantTradeNo:   "test1237a",
		TradeDate:         time.Now(),
		ShippingStoreType: "711",
		Amount:            100,
		NeedPayment:       false,
		EntreeName:        "測試物品",
		ReceiverName:      "測試人員",
		ReceiverPhone:     "0912345678",
		ReceiverEmail:     "abc@gmail.com",
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
