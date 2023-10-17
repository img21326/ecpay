package ecpay

import "strings"

func TransferStoreType(storeType string) string {
	storeTypes := strings.Split(storeType, "_")
	storeType = storeTypes[len(storeTypes)-1]
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
