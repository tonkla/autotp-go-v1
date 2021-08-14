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

func Connect(dbName string) *DB {
	if dbName == "" {
		dbName = "autotp.db"
	}
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalln(err)
	}
	db.AutoMigrate(&t.Order{})
	return &DB{db: db}
}

func (d DB) IsEmptyZone(o t.Order) bool {
	var order t.Order
	d.db.Where(`bot_id = ? AND exchange = ? AND symbol = ? AND zone_price = ? AND side = ?
	AND (type = ? OR type = ?) AND status <> ? AND close_order_id = ''`,
		o.BotID, o.Exchange, o.Symbol, o.ZonePrice, o.Side, t.OrderTypeLimit, t.OrderTypeMarket,
		t.OrderStatusCanceled).First(&order)
	return order.OpenPrice == 0
}

func (d DB) GetOrderByID(id string) *t.Order {
	var order t.Order
	d.db.Where("id = ?", id).First(&order)
	if order.ID == "" {
		return nil
	}
	return &order
}

// GetActiveOrders returns all open orders that are not canceled
func (d DB) GetActiveOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND (type = ? OR type = ?) AND status <> ? AND close_order_id = ''",
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeLimit, t.OrderTypeMarket, t.OrderStatusCanceled).Find(&orders)
	return orders
}

// GetLimitOrder returns the LIMIT order that is not canceled
func (d DB) GetLimitOrder(o t.Order, slippage float64) *t.Order {
	var order t.Order
	if slippage > 0 {
		lowerPrice := o.OpenPrice - o.OpenPrice*slippage
		upperPrice := o.OpenPrice + o.OpenPrice*slippage
		d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND open_price BETWEEN ? AND ? AND side = ? AND type = ? AND status <> ? AND close_order_id = ''",
			o.BotID, o.Exchange, o.Symbol, lowerPrice, upperPrice, o.Side, t.OrderTypeLimit, t.OrderStatusCanceled).First(&order)
	} else {
		d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND open_price = ? AND side = ? AND type = ? AND status <> ? AND close_order_id = ''",
			o.BotID, o.Exchange, o.Symbol, o.OpenPrice, o.Side, t.OrderTypeLimit, t.OrderStatusCanceled).First(&order)
	}
	if order.OpenPrice == 0 {
		return nil
	}
	return &order
}

// GetLimitOrders returns the LIMIT orders that are not canceled
func (d DB) GetLimitOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND status <> ? AND close_order_id = ''",
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeLimit, t.OrderStatusCanceled).Find(&orders)
	return orders
}

// GetLimitOrdersBySide returns the LIMIT orders that are not canceled for the specified side
func (d DB) GetLimitOrdersBySide(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND side = ? AND type = ? AND status <> ? AND close_order_id = ''",
		o.BotID, o.Exchange, o.Symbol, o.Side, t.OrderTypeLimit, t.OrderStatusCanceled).Find(&orders)
	return orders
}

// GetNewOrders returns the orders that their status is NEW
func (d DB) GetNewOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND status = ?",
		o.BotID, o.Exchange, o.Symbol, t.OrderStatusNew).Find(&orders)
	return orders
}

// GetNewOrdersBySide returns the orders that their status is NEW for the specified side
func (d DB) GetNewOrdersBySide(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND status = ? AND side = ?",
		o.BotID, o.Exchange, o.Symbol, t.OrderStatusNew, o.Side).Find(&orders)
	return orders
}

// GetFilledOrders returns the orders that their status is FILLED
func (d DB) GetFilledOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND status = ? AND close_order_id = ''",
		o.BotID, o.Exchange, o.Symbol, t.OrderStatusFilled).Find(&orders)
	return orders
}

// GetFilledOrdersBySide returns the orders that their status is FILLED for the specified side
func (d DB) GetFilledOrdersBySide(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND status = ? AND side = ? AND close_order_id = ''",
		o.BotID, o.Exchange, o.Symbol, t.OrderStatusFilled, o.Side).Find(&orders)
	return orders
}

// GetProfitOrdersBySide returns the orders that are profitable for the specified side
func (d DB) GetProfitOrdersBySide(qo t.Order, tk t.Ticker) []t.Order {
	var orders []t.Order
	fee := tk.Price * 0.002 * 2 // 0.002=transaction fee at 0.2%, 2=open and closed fees
	for _, o := range d.GetFilledOrdersBySide(qo) {
		if o.Side == t.OrderSideBuy {
			profit := tk.Price - fee
			if o.OpenPrice < profit {
				orders = append(orders, o)
			}
		} else if o.Side == t.OrderSideSell {
			profit := tk.Price + fee
			if o.OpenPrice > profit {
				orders = append(orders, o)
			}
		}
	}
	return orders
}

// GetOppositeOrder returns the opposite order that will close this order
func (d DB) GetOppositeOrder(id string) *t.Order {
	var order t.Order
	d.db.Where("open_order_id = ? AND type = ? AND status <> ?",
		id, t.OrderTypeLimit, t.OrderStatusCanceled).First(&order)
	if order.ID == "" {
		return nil
	}
	return &order
}

// GetSLOrder returns the Stop Loss order of the order
func (d DB) GetSLOrder(openOrderID string) *t.Order {
	var order t.Order
	d.db.Where("open_order_id = ? AND type = ? AND status <> ?",
		openOrderID, t.OrderTypeSL, t.OrderStatusCanceled).First(&order)
	if order.ID == "" {
		return nil
	}
	return &order
}

// GetTPOrder returns the Take Profit order of the order
func (d DB) GetTPOrder(openOrderID string) *t.Order {
	var order t.Order
	d.db.Where("open_order_id = ? AND type = ? AND status <> ?",
		openOrderID, t.OrderTypeTP, t.OrderStatusCanceled).First(&order)
	if order.ID == "" {
		return nil
	}
	return &order
}

// GetSLOrders returns the STOP_LOSS_LIMIT orders that are not canceled
func (d DB) GetSLOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND status <> ? ORDER BY open_price ASC",
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeSL, t.OrderStatusCanceled).Find(&orders)
	return orders
}

// GetTPOrders returns the TAKE_PROFIT_LIMIT orders that are not canceled
func (d DB) GetTPOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND status <> ? ORDER BY open_price DESC",
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeTP, t.OrderStatusCanceled).Find(&orders)
	return orders
}

// GetNewSLOrders returns the STOP_LOSS_LIMIT orders that their status is NEW
func (d DB) GetNewSLOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND status = ? ORDER BY open_price ASC",
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeSL, t.OrderStatusNew).Find(&orders)
	return orders
}

// GetNewTPOrders returns the TAKE_PROFIT_LIMIT orders that their status is NEW
func (d DB) GetNewTPOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND status = ? ORDER BY open_price DESC",
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeTP, t.OrderStatusNew).Find(&orders)
	return orders
}

// CreateOrder performs SQL insert on the table orders
func (d DB) CreateOrder(order t.Order) error {
	return d.db.Create(&order).Error
}

// UpdateOrder performs SQL update on the table orders
func (d DB) UpdateOrder(order t.Order) error {
	return d.db.Updates(&order).Error
}
