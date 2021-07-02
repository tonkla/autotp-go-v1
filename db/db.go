package db

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/tonkla/autotp/types"
)

type DB struct {
	db *gorm.DB
}

func Connect() *DB {
	db, err := gorm.Open(sqlite.Open("autotp.db"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalln(err)
	}
	db.AutoMigrate(&types.Order{})
	return &DB{db: db}
}

// GetActiveOrder returns the order that its status is not CLOSED
func (d DB) GetActiveOrder(o types.Order, slippage float64) *types.Order {
	var order types.Order
	if slippage > 0 {
		lowerPrice := o.OpenPrice - (o.OpenPrice * slippage)
		upperPrice := o.OpenPrice + (o.OpenPrice * slippage)
		d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND open_price BETWEEN ? AND ? AND side = ? AND status <> ?",
			o.BotID, o.Exchange, o.Symbol, lowerPrice, upperPrice, o.Side, types.ORDER_STATUS_CLOSED).First(&order)
	} else {
		d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND open_price = ? AND side = ? AND status <> ?",
			o.BotID, o.Exchange, o.Symbol, o.OpenPrice, o.Side, types.ORDER_STATUS_CLOSED).First(&order)
	}
	return &order
}

// GetActiveOrders returns the orders that their status is not CLOSED
func (d DB) GetActiveOrders(o types.Order) []types.Order {
	var orders []types.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND side = ? AND status <> ?",
		o.BotID, o.Exchange, o.Symbol, o.Side, types.ORDER_STATUS_CLOSED).Find(&orders)
	return orders
}

// GetProfitOrders returns the orders that are profitable
func (d DB) GetProfitOrders(o types.Order) []types.Order {
	var orders []types.Order
	fee := o.ClosePrice * 0.002 // tx fee is 0.2%
	if o.Side == types.ORDER_SIDE_BUY {
		priceWithFee := o.ClosePrice - fee
		d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND side = ? AND status = ? AND open_price < ?",
			o.BotID, o.Exchange, o.Symbol, o.Side, types.ORDER_STATUS_FILLED, priceWithFee).Find(&orders)
	} else if o.Side == types.ORDER_SIDE_SELL {
		priceWithFee := o.ClosePrice + fee
		d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND side = ? AND status = ? AND open_price > ?",
			o.BotID, o.Exchange, o.Symbol, o.Side, types.ORDER_STATUS_FILLED, priceWithFee).Find(&orders)
	}
	return orders
}

// IsOrderActive checks the order is active
func (d DB) IsOrderActive(o types.Order, slippage float64) bool {
	return d.GetActiveOrder(o, slippage).ID > 0
}

// CreateOrder performs SQL insert on the table orders
func (d DB) CreateOrder(order types.Order) error {
	return d.db.Create(&order).Error
}

// UpdateOrder performs SQL update on the table orders
func (d DB) UpdateOrder(order types.Order) error {
	return d.db.Save(&order).Error
}
