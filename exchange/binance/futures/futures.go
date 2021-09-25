package binance

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	b "github.com/tonkla/autotp/exchange/binance"
	h "github.com/tonkla/autotp/helper"
	t "github.com/tonkla/autotp/types"
)

type Client struct {
	baseURL   string
	apiKey    string
	secretKey string
}

// NewFuturesClient returns Binance USDⓈ-M Futures client
func NewFuturesClient(apiKey string, secretKey string) Client {
	return Client{
		baseURL:   "https://fapi.binance.com/fapi/v1",
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

// GetTicker returns the latest ticker
func (c Client) GetTicker(symbol string) *t.Ticker {
	return b.GetTicker(c.baseURL, symbol)
}

// GetOrderBook returns an order book (market depth
func (c Client) GetOrderBook(symbol string, limit int) *t.OrderBook {
	return b.GetOrderBook(c.baseURL, symbol, limit)
}

// GetHistoricalPrices returns historical prices in a format of k-lines/candlesticks
func (c Client) GetHistoricalPrices(symbol string, timeframe string, limit int) []t.HistoricalPrice {
	return b.GetHistoricalPrices(c.baseURL, symbol, timeframe, limit)
}

// PlaceLimitOrder places a limit order
func (c Client) PlaceLimitOrder(o t.Order) (*t.Order, error) {
	if o.Type != t.OrderTypeLimit {
		return nil, nil
	}

	var payload, url strings.Builder

	b.BuildBaseQS(&payload, o.Symbol)
	fmt.Fprintf(&payload, "&newClientOrderId=%s&side=%s&positionSide=%s&type=%s&quantity=%f&price=%f&timeInForce=GTC",
		o.ID, o.Side, o.PosSide, o.Type, o.Qty, o.OpenPrice)

	signature := b.Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/order?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := h.Post(url.String(), b.NewHeader(c.apiKey))
	if err != nil {
		return nil, err
	}

	r := gjson.ParseBytes(data)

	if r.Get("code").Int() < 0 {
		h.Log("PlaceLimitOrder", r)
		return nil, errors.New(r.Get("msg").String())
	}

	status := r.Get("status").String()
	if status != t.OrderStatusNew && status != t.OrderStatusFilled {
		return nil, nil
	}
	o.Status = status
	o.RefID = r.Get("orderId").String()
	o.OpenTime = r.Get("transactTime").Int()
	price := r.Get("price").Float()
	if price > 0 {
		o.OpenPrice = price
	}
	return &o, nil
}

// PlaceStopOrder places a stop order
func (c Client) PlaceStopOrder(o t.Order) (*t.Order, error) {
	if o.Type != t.OrderTypeFSL && o.Type != t.OrderTypeFTP {
		return nil, nil
	}

	var payload, url strings.Builder

	b.BuildBaseQS(&payload, o.Symbol)
	fmt.Fprintf(&payload, "&newClientOrderId=%s&side=%s&positionSide=%s&type=%s&quantity=%f&price=%f&stopPrice=%f&timeInForce=GTC",
		o.ID, o.Side, o.PosSide, o.Type, o.Qty, o.OpenPrice, o.StopPrice)

	signature := b.Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/order?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := h.Post(url.String(), b.NewHeader(c.apiKey))
	if err != nil {
		return nil, err
	}

	r := gjson.ParseBytes(data)

	if r.Get("code").Int() < 0 {
		h.Log("PlaceStopOrder", r)
		return nil, errors.New(r.Get("msg").String())
	}

	o.RefID = r.Get("orderId").String()
	o.OpenTime = r.Get("transactTime").Int()
	return &o, nil
}

// GetTradeList returns trades list for a specified symbol
func (c Client) GetTradeList(symbol string, limit int64) ([]t.Order, error) {
	var payload, url strings.Builder

	b.BuildBaseQS(&payload, symbol)
	fmt.Fprintf(&payload, "&limit=%d", limit)

	signature := b.Sign(payload.String(), c.secretKey)

	fmt.Fprintf(&url, "%s/userTrades?%s&signature=%s", c.baseURL, payload.String(), signature)
	data, err := h.GetH(url.String(), b.NewHeader(c.apiKey))
	if err != nil {
		return nil, err
	}

	rs := gjson.ParseBytes(data)

	if rs.Get("code").Int() < 0 {
		h.Log("GetTradeList", rs)
		return nil, errors.New(rs.Get("msg").String())
	}

	var orders []t.Order
	for _, r := range rs.Array() {
		order := t.Order{
			Symbol:     symbol,
			RefID:      r.Get("orderId").String(),
			Side:       r.Get("side").String(),
			PosSide:    r.Get("positionSide").String(),
			Qty:        r.Get("qty").Float(),
			OpenPrice:  r.Get("price").Float(),
			OpenTime:   r.Get("time").Int(),
			Commission: r.Get("commission").Float(),
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// GetOrder returns the order by its IDs
func (c Client) GetOrder(o t.Order) (*t.Order, error) {
	cc := b.CClient{
		BaseURL:   c.baseURL,
		ApiKey:    c.apiKey,
		SecretKey: c.secretKey,
	}
	return b.GetOrder(cc, o)
}
