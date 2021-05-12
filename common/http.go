package common

import (
	"io"
	"net/http"
	"time"
)

// Get calls a URL with HTTP GET
func Get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Post calls a URL with HTTP POST
func Post(url string, data string) TradeResult {
	return TradeResult{Time: time.Now().Unix(), Symbol: "", Side: "", Price: 0, Qty: 0}
}
