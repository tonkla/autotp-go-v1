package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Sign signs a payload with a Binance API secret key
func Sign(payload string, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
