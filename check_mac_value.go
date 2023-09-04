package ecpay_shipping

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

type CheckMacValueService interface {
	GenerateCheckMacValue(params map[string]string) string
}

type shipMacService struct {
	ec EcpayConfig
}

type paymentMacService struct {
	ec EcpayConfig
}

func NewShipMacValue(ec EcpayConfig) CheckMacValueService {
	return &shipMacService{
		ec: ec,
	}
}

func (c *shipMacService) GenerateCheckMacValue(params map[string]string) string {
	encodedParams := NewECPayValuesFromMap(params).Encode()
	encodedParams = fmt.Sprintf("HashKey=%s&%s&HashIV=%s", c.ec.HashKey, encodedParams, c.ec.HashIV)
	encodedParams = FormUrlEncode(encodedParams)
	encodedParams = strings.ToLower(encodedParams)
	sum := md5.Sum([]byte(encodedParams))
	checkMac := strings.ToUpper(hex.EncodeToString(sum[:]))
	return checkMac
}

func NewPaymentMacValue(ec EcpayConfig) CheckMacValueService {
	return &paymentMacService{
		ec: ec,
	}
}

func (c *paymentMacService) GenerateCheckMacValue(params map[string]string) string {
	encodedParams := NewECPayValuesFromMap(params).Encode()
	encodedParams = fmt.Sprintf("HashKey=%s&%s&HashIV=%s", c.ec.HashKey, encodedParams, c.ec.HashIV)
	encodedParams = FormUrlEncode(encodedParams)
	encodedParams = strings.ToLower(encodedParams)
	sum := sha256.Sum256([]byte(encodedParams))
	checkMac := strings.ToUpper(hex.EncodeToString(sum[:]))
	return checkMac
}

func FormUrlEncode(s string) string {
	s = url.QueryEscape(s)
	s = strings.ReplaceAll(s, "%21", "!")
	s = strings.ReplaceAll(s, "%2A", "*")
	s = strings.ReplaceAll(s, "%28", "(")
	s = strings.ReplaceAll(s, "%29", ")")
	s = strings.ReplaceAll(s, "~", "%7e")
	return s
}

type LowerStringSlice []string

func (p LowerStringSlice) Len() int           { return len(p) }
func (p LowerStringSlice) Less(i, j int) bool { return strings.ToLower(p[i]) < strings.ToLower(p[j]) }
func (p LowerStringSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type ECPayValues struct {
	url.Values
}

func NewECPayValuesFromMap(values map[string]string) *ECPayValues {
	v := ECPayValues{Values: url.Values{}}
	for key, value := range values {
		v.Set(key, value)
	}
	return &v
}

func (v ECPayValues) Encode() string {
	if v.Values == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(v.Values))
	for k := range v.Values {
		keys = append(keys, k)
	}
	sort.Sort(LowerStringSlice(keys))
	for _, k := range keys {
		vs := v.Values[k]
		//keyEscaped := url.QueryEscape(k)
		keyEscaped := k
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(v)
		}
	}
	return buf.String()
}
