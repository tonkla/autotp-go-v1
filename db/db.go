package db

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	t "github.com/tonkla/autotp/types"
)

type DB struct {
	db *gorm.DB
}

func Connect() *DB {
	db, err := gorm.Open(sqlite.Open("autotp.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalln(err)
	}
	db.AutoMigrate(&t.Order{})
	return &DB{db: db}
}

// GetActiveOrder returns the order that its status is not CLOSED
func (d DB) GetActiveOrder(o t.Order, slippage float64) *t.Order {
	var order t.Order
	if slippage > 0 {
		lowerPrice := o.OpenPrice - (o.OpenPrice * slippage)
		upperPrice := o.OpenPrice + (o.OpenPrice * slippage)
		d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND open_price BETWEEN ? AND ? AND side = ? AND status <> ? AND status <> ?",
			o.BotID, o.Exchange, o.Symbol, lowerPrice, upperPrice, o.Side,
			t.OrderStatusClosed, t.OrderStatusCanceled).First(&order)
	} else {
		d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND open_price = ? AND side = ? AND status <> ? AND status <> ?",
			o.BotID, o.Exchange, o.Symbol, o.OpenPrice, o.Side,
			t.OrderStatusClosed, t.OrderStatusCanceled).First(&order)
	}
	if order.OpenPrice == 0 {
		return nil
	}
	return &order
}

// GetActiveOrders returns the orders that their status is not CLOSED
func (d DB) GetActiveOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND side = ? AND status <> ? AND status <> ?",
		o.BotID, o.Exchange, o.Symbol, o.Side, t.OrderStatusClosed, t.OrderStatusCanceled).Find(&orders)
	return orders
}

// GetNewOrders returns the orders that their status is NEW
func (d DB) GetNewOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND status = ?",
		o.BotID, o.Exchange, o.Symbol, t.OrderStatusNew).Find(&orders)
	return orders
}

// GetFilledOrders returns the orders that their status is FILLED
func (d DB) GetFilledOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND status = ?",
		o.BotID, o.Exchange, o.Symbol, t.OrderStatusFilled).Find(&orders)
	return orders
}

// GetProfitOrders returns the orders that are profitable
func (d DB) GetProfitOrders(o t.Order, tk t.Ticker) []t.Order {
	var orders []t.Order
	fee := tk.Price * 0.002 * 2 // 0.002=transaction fee at 0.2%, 2=open and closed fees
	if o.Side == t.OrderSideBuy {
		profit := tk.Price - fee
		d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND side = ? AND status = ? AND open_price < ?",
			o.BotID, o.Exchange, o.Symbol, o.Side, t.OrderStatusFilled, profit).Find(&orders)
	} else if o.Side == t.OrderSideSell {
		profit := tk.Price + fee
		d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND side = ? AND status = ? AND open_price > ?",
			o.BotID, o.Exchange, o.Symbol, o.Side, t.OrderStatusFilled, profit).Find(&orders)
	}
	return orders
}

// IsOrderActive checks the order is active
func (d DB) IsOrderActive(o t.Order, slippage float64) bool {
	return d.GetActiveOrder(o, slippage).RefID1 > 0
}

// CreateOrder performs SQL insert on the table orders
func (d DB) CreateOrder(order t.Order) error {
	return d.db.Create(&order).Error
}

// UpdateOrder performs SQL update on the table orders
func (d DB) UpdateOrder(order t.Order) error {
	return d.db.Where("ref_id1 = ?", order.RefID1).Updates(&order).Error
}
