package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"
	h "github.com/tonkla/autotp/helper"
	t "github.com/tonkla/autotp/types"
)

type Client struct {
	BaseURL   string
	ApiKey    string
	SecretKey string
}

// Sign signs a payload with a Binance API secret key
func Sign(payload string, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func NewHeader(apiKey string) http.Header {
	var header http.Header = make(map[string][]string)
	header.Set("X-MBX-APIKEY", apiKey)
	return header
}

// Build a base query string
func BuildBaseQS(payload *strings.Builder, symbol string) {
	fmt.Fprintf(payload, "timestamp=%d&recvWindow=50000&symbol=%s", h.Now13(), symbol)
}

// GetTicker returns the latest ticker
func GetTicker(baseURL string, symbol string) *t.Ticker {
	var url strings.Builder

	fmt.Fprintf(&url, "%s/ticker/price?symbol=%s", baseURL, symbol)
	data, err := h.Get(url.String())
	if err != nil {
		return nil
	}
	r := gjson.ParseBytes(data)
	return &t.Ticker{
		Exchange: t.ExcBinance,
		Symbol:   r.Get("symbol").String(),
		Price:    r.Get("price").Float(),
		Time:     r.Get("time").Int(),
	}
}

// GetOrderBook returns an order book (market depth)
func GetOrderBook(baseURL string, symbol string, limit int) *t.OrderBook {
	var url strings.Builder

	fmt.Fprintf(&url, "%s/depth?symbol=%s&limit=%d", baseURL, symbol, limit)
	data, err := h.Get(url.String())
	if err != nil {
		return nil
	}

	var bids, asks []t.ExOrder
	result := gjson.ParseBytes(data)
	for _, bid := range result.Get("bids").Array() {
		b := bid.Array()
		bids = append(bids, t.ExOrder{
			Price: b[0].Float(),
			Qty:   b[1].Float(),
		})
	}
	for _, ask := range result.Get("asks").Array() {
		a := ask.Array()
		asks = append(asks, t.ExOrder{
			Price: a[0].Float(),
			Qty:   a[1].Float(),
		})
	}
	return &t.OrderBook{
		Symbol: symbol,
		Bids:   bids,
		Asks:   asks,
	}
}

// GetHistoricalPrices returns historical prices in a format of k-lines/candlesticks
func GetHistoricalPrices(baseURL string, symbol string, timeframe string, limit int) []t.HistoricalPrice {
	var url strings.Builder

	fmt.Fprintf(&url, "%s/klines?symbol=%s&interval=%s&limit=%d", baseURL, symbol, timeframe, limit)
	data, err := h.Get(url.String())
	if err != nil {
		return nil
	}

	var hPrices []t.HistoricalPrice
	for _, data := range gjson.ParseBytes(data).Array() {
		d := data.Array()
		p := t.HistoricalPrice{
			Symbol: symbol,
			Time:   d[0].Int(),
			Open:   d[1].Float(),
			High:   d[2].Float(),
			Low:    d[3].Float(),
			Close:  d[4].Float(),
		}
		hPrices = append(hPrices, p)
	}
	return hPrices
}

// GetOrderByID returns the order by its ID
func GetOrderByID(c Client, symbol string, ID string, refID string) (*t.Order, error) {
	if symbol == "" || (ID == "" && refID == "") {
		return nil, nil
	}

	var payload, url strings.Builder

	BuildBaseQS(&payload, symbol)
	if refID != "" {
		fmt.Fprintf(&payload, "&orderId=%s", refID)
	}
	if ID != "" {
		fmt.Fprintf(&payload, "&origClientOrderId=%s", ID)
	}

	signature := Sign(payload.String(), c.SecretKey)

	fmt.Fprintf(&url, "%s/order?%s&signature=%s", c.BaseURL, payload.String(), signature)
	data, err := h.GetH(url.String(), NewHeader(c.ApiKey))
	if err != nil {
		return nil, err
	}

	r := gjson.ParseBytes(data)

	if r.Get("code").Int() < 0 {
		h.Log("GetOrderByID", r)
		return nil, errors.New(r.Get("msg").String())
	}

	return &t.Order{
		ID:         ID,
		RefID:      refID,
		Symbol:     symbol,
		Status:     r.Get("status").String(),
		UpdateTime: r.Get("updateTime").Int(),
	}, nil
}

// GetOrder returns the order by its IDs
func GetOrder(c Client, o t.Order) (*t.Order, error) {
	exo, err := GetOrderByID(c, o.Symbol, o.ID, o.RefID)
	if err != nil {
		return nil, err
	}
	if exo == nil {
		return nil, nil
	}
	o.Status = exo.Status
	o.UpdateTime = exo.UpdateTime
	return &o, nil
}
