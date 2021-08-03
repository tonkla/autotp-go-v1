package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/tidwall/gjson"

	h "github.com/tonkla/autotp/helper"
	t "github.com/tonkla/autotp/types"
)

type Client struct {
	baseURL   string
	apiKey    string
	secretKey string
}

// NewSpotClient returns Binance Spot client
func NewSpotClient(apiKey string, secretKey string) Client {
	return Client{
		baseURL:   "https://api.binance.com/api/v3",
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

// NewTestSpotClient returns Binance Test Network Spot client
func NewTestSpotClient(apiKey string, secretKey string) Client {
	return Client{
		baseURL:   "https://testnet.binance.vision/api/v3",
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

// NewFuturesClient returns Binance USDⓈ-M Futures client
func NewFuturesClient(apiKey string, secretKey string) Client {
	return Client{
		baseURL:   "https://fapi.binance.com/fapi/v1",
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

// Sign signs a payload with a Binance API secret key
func Sign(payload string, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func newHeader(apiKey string) http.Header {
	var header http.Header = make(map[string][]string)
	header.Set("X-MBX-APIKEY", apiKey)
	return header
}

// Build a base query string
func buildBaseQS(payload *strings.Builder, symbol string) {
	fmt.Fprintf(payload, "timestamp=%d&recvWindow=20000&symbol=%s", h.Now13(), symbol)
}

// Public APIs -----------------------------------------------------------------

// GetTicker returns the latest ticker
func (c Client) GetTicker(symbol string) *t.Ticker {
	var url strings.Builder

	fmt.Fprintf(&url, "%s/ticker/price?symbol=%s", c.baseURL, symbol)
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
func (c Client) GetOrderBook(symbol string, limit int) *t.OrderBook {
	var url strings.Builder

	fmt.Fprintf(&url, "%s/depth?symbol=%s&limit=%d", c.baseURL, symbol, limit)
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
func (c Client) GetHistoricalPrices(symbol string, timeframe string, limit int) []t.HistoricalPrice {
	var url strings.Builder

	fmt.Fprintf(&url, "%s/klines?symbol=%s&interval=%s&limit=%d", c.baseURL, symbol, timeframe, limit)
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

// Private APIs ----------------------------------------------------------------

// GetOrderByIDs returns the order by its IDs
func (c Client) GetOrderByIDs(symbol string, ID string, refID string) *t.Order {
	if symbol == "" || (ID == "" && refID == "") {
		return nil
	}

	var payload, url strings.Builder

	buildBaseQS(&payload, symbol)
	if refID != "" {
		fmt.Fprintf(&payload, "&orderId=%s", refID)
	}
	if ID != "" {
		fmt.Fprintf(&payload, "&origClientOrderId=%s", ID)
	}

	signature := Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/order?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := h.GetH(url.String(), newHeader(c.apiKey))
	if err != nil {
		h.Log(err)
		return nil
	}

	r := gjson.ParseBytes(data)

	if r.Get("code").Int() < 0 {
		h.Log(r)
		return nil
	}

	return &t.Order{
		ID:         ID,
		RefID:      refID,
		Symbol:     symbol,
		Status:     r.Get("status").String(),
		UpdateTime: r.Get("updateTime").Int(),
	}
}

// GetOrder returns the order by its IDs
func (c Client) GetOrder(o t.Order) *t.Order {
	exo := c.GetOrderByIDs(o.Symbol, o.ID, o.RefID)
	if exo == nil {
		return nil
	}
	o.Status = exo.Status
	o.UpdateTime = exo.UpdateTime
	return &o
}

// GetOpenOrders returns open orders
func (c Client) GetOpenOrders(symbol string) []t.Order {
	var payload, url strings.Builder

	buildBaseQS(&payload, symbol)

	signature := Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/openOrders?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := h.GetH(url.String(), newHeader(c.apiKey))
	if err != nil {
		return nil
	}

	rs := gjson.ParseBytes(data)

	if rs.Get("code").Int() < 0 {
		h.Log(rs)
		return nil
	}

	var orders []t.Order
	for _, r := range rs.Array() {
		order := t.Order{
			Symbol:     symbol,
			ID:         r.Get("clientOrderId").String(),
			RefID:      r.Get("orderId").String(),
			Side:       r.Get("side").String(),
			Status:     r.Get("status").String(),
			Type:       r.Get("type").String(),
			OpenTime:   r.Get("time").Int(),
			Qty:        r.Get("origQty").Float(),
			OpenPrice:  r.Get("price").Float(),
			UpdateTime: r.Get("updateTime").Int(),
		}
		orders = append(orders, order)
	}
	return orders
}

// GetAllOrders returns all account orders; active, canceled, or filled
func (c Client) GetAllOrders(symbol string, limit int, startTime int, endTime int) []t.Order {
	var payload, url strings.Builder

	buildBaseQS(&payload, symbol)

	if limit > 0 {
		fmt.Fprintf(&payload, "&limit=%d", limit)
	}
	if startTime > 0 {
		fmt.Fprintf(&payload, "&startTime=%d", startTime)
	}
	if endTime > 0 {
		fmt.Fprintf(&payload, "&endTime=%d", endTime)
	}

	signature := Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/allOrders?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := h.GetH(url.String(), newHeader(c.apiKey))
	if err != nil {
		return nil
	}

	rs := gjson.ParseBytes(data)

	if rs.Get("code").Int() < 0 {
		h.Log(rs)
		return nil
	}

	var orders []t.Order
	for _, r := range rs.Array() {
		order := t.Order{
			Symbol:    symbol,
			ID:        r.Get("clientOrderId").String(),
			RefID:     r.Get("orderId").String(),
			Side:      r.Get("side").String(),
			Status:    r.Get("status").String(),
			Type:      r.Get("type").String(),
			OpenTime:  r.Get("time").Int(),
			Qty:       r.Get("origQty").Float(),
			OpenPrice: r.Get("price").Float(),
		}
		orders = append(orders, order)
	}
	return orders
}

// PlaceLimitOrder places a limit order on the exchange
func (c Client) PlaceLimitOrder(o t.Order) *t.Order {
	if o.Type != t.OrderTypeLimit {
		return nil
	}

	var payload, url strings.Builder

	buildBaseQS(&payload, o.Symbol)
	fmt.Fprintf(&payload, "&newClientOrderId=%s&side=%s&type=%s&quantity=%f&price=%f&timeInForce=GTC",
		o.ID, o.Side, o.Type, o.Qty, o.OpenPrice)

	signature := Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/order?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := h.Post(url.String(), newHeader(c.apiKey))
	if err != nil {
		return nil
	}

	r := gjson.ParseBytes(data)

	if r.Get("code").Int() < 0 {
		h.Log(r)
		return nil
	}

	status := r.Get("status").String()
	if status != t.OrderStatusNew && status != t.OrderStatusFilled {
		return nil
	}
	o.Status = status
	o.RefID = r.Get("orderId").String()
	o.OpenTime = r.Get("transactTime").Int()
	price := r.Get("price").Float()
	if price > 0 {
		o.OpenPrice = price
	}
	return &o
}

// PlaceStopOrder places a stop order on the exchange
func (c Client) PlaceStopOrder(o t.Order) *t.Order {
	if o.Type != t.OrderTypeSL && o.Type != t.OrderTypeTP {
		return nil
	}

	var payload, url strings.Builder

	buildBaseQS(&payload, o.Symbol)
	fmt.Fprintf(&payload, "&newClientOrderId=%s&side=%s&type=%s&quantity=%f&price=%f&stopPrice=%f&timeInForce=GTC",
		o.ID, o.Side, o.Type, o.Qty, o.OpenPrice, o.StopPrice)

	signature := Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/order?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := h.Post(url.String(), newHeader(c.apiKey))
	if err != nil {
		return nil
	}

	r := gjson.ParseBytes(data)

	if r.Get("code").Int() < 0 {
		h.Log(r)
		return nil
	}

	o.RefID = r.Get("orderId").String()
	o.OpenTime = r.Get("transactTime").Int()
	return &o
}

// PlaceMarketOrder places a market order on the exchange
func (c Client) PlaceMarketOrder(o t.Order) *t.Order {
	if o.Type != t.OrderTypeMarket {
		return nil
	}

	var payload, url strings.Builder

	buildBaseQS(&payload, o.Symbol)
	fmt.Fprintf(&payload, "&newClientOrderId=%s&side=%s&type=%s&quantity=%f",
		o.ID, o.Side, o.Type, o.Qty)

	signature := Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/order?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := h.Post(url.String(), newHeader(c.apiKey))
	if err != nil {
		return nil
	}

	r := gjson.ParseBytes(data)

	if r.Get("code").Int() < 0 {
		h.Log(r)
		return nil
	}

	o.RefID = r.Get("orderId").String()
	o.OpenTime = r.Get("transactTime").Int()
	o.OpenPrice = r.Get("price").Float()
	o.Status = r.Get("status").String()
	return &o
}

// CancelOrder cancels an order on the exchange
func (c Client) CancelOrder(o t.Order) *t.Order {
	var payload, url strings.Builder

	buildBaseQS(&payload, o.Symbol)
	fmt.Fprintf(&payload, "&orderId=%s&origClientOrderId=%s", o.RefID, o.ID)

	signature := Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/order?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := h.Delete(url.String(), newHeader(c.apiKey))
	if err != nil {
		return nil
	}
	status := gjson.ParseBytes(data).Get("status").String()
	if status != t.OrderStatusCanceled {
		return nil
	}
	o.Status = status
	return &o
}
