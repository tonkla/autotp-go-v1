package db

import (
	"log"
	"math"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	t "github.com/tonkla/autotp/types"
)

type DB struct {
	db *gorm.DB
}

// Connect returns an instance of the DB
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

// IsEmptyZone checks zone availability
func (d DB) IsEmptyZone(o t.Order) bool {
	var order t.Order
	d.db.Where(`bot_id = ? AND exchange = ? AND symbol = ? AND zone_price = ? AND side = ?
	AND (type = ? OR type = ?) AND status <> ? AND close_time = 0`,
		o.BotID, o.Exchange, o.Symbol, o.ZonePrice, o.Side, t.OrderTypeLimit, t.OrderTypeMarket,
		t.OrderStatusCanceled).First(&order)
	return order.OpenPrice == 0
}

// GetOrderByID returns an order by the specified ID
func (d DB) GetOrderByID(id string) *t.Order {
	var order t.Order
	d.db.Where("id = ?", id).First(&order)
	if order.ID == "" {
		return nil
	}
	return &order
}

// GetHighestNewBuyOrder returns the highest price, NEW, BUY order
func (d DB) GetHighestNewBuyOrder(o t.Order) *t.Order {
	var orders []t.Order
	d.db.Where(`bot_id = ? AND exchange = ? AND symbol = ? AND side = ? AND type = ?
	AND status = ? AND close_time = 0`, o.BotID, o.Exchange, o.Symbol, t.OrderSideBuy,
		t.OrderTypeLimit, t.OrderStatusNew).Order("zone_price desc").Limit(1).Find(&orders)
	if len(orders) == 0 {
		return nil
	}
	return &orders[0]
}

// GetLowestNewSellOrder returns the lowest price, NEW, SELL order
func (d DB) GetLowestNewSellOrder(o t.Order) *t.Order {
	var orders []t.Order
	d.db.Where(`bot_id = ? AND exchange = ? AND symbol = ? AND side = ? AND type = ?
	AND status = ? AND close_time = 0`, o.BotID, o.Exchange, o.Symbol, t.OrderSideSell,
		t.OrderTypeLimit, t.OrderStatusNew).Order("zone_price asc").Limit(1).Find(&orders)
	if len(orders) == 0 {
		return nil
	}
	return &orders[0]
}

// GetLowestFilledBuyOrder returns the lowest price, FILLED, BUY order
func (d DB) GetLowestFilledBuyOrder(o t.Order) *t.Order {
	var orders []t.Order
	d.db.Where(`bot_id = ? AND exchange = ? AND symbol = ? AND side = ? AND (type = ? OR type = ?)
	AND status = ? AND close_time = 0`, o.BotID, o.Exchange, o.Symbol, t.OrderSideBuy, t.OrderTypeLimit,
		t.OrderTypeMarket, t.OrderStatusFilled).Order("zone_price asc").Limit(1).Find(&orders)
	if len(orders) == 0 {
		return nil
	}
	return &orders[0]
}

// GetHighestFilledSellOrder returns the highest price, FILLED, SELL order
func (d DB) GetHighestFilledSellOrder(o t.Order) *t.Order {
	var orders []t.Order
	d.db.Where(`bot_id = ? AND exchange = ? AND symbol = ? AND side = ? AND (type = ? OR type = ?)
	AND status = ? AND close_time = 0`, o.BotID, o.Exchange, o.Symbol, t.OrderSideSell, t.OrderTypeLimit,
		t.OrderTypeMarket, t.OrderStatusFilled).Order("zone_price desc").Limit(1).Find(&orders)
	if len(orders) == 0 {
		return nil
	}
	return &orders[0]
}

// GetActiveOrders returns all open orders that are not canceled
func (d DB) GetActiveOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND (type = ? OR type = ?) AND status <> ? AND close_time = 0",
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeLimit, t.OrderTypeMarket, t.OrderStatusCanceled).Find(&orders)
	return orders
}

// GetActiveOrdersBySide returns all open orders that are not canceled for the specified side
func (d DB) GetActiveOrdersBySide(o t.Order) []t.Order {
	var orders []t.Order
	q := d.db.Where(`
	bot_id = ? AND exchange = ? AND symbol = ? AND side = ? AND (type = ? OR type = ?) AND status <> ? AND close_time = 0`,
		o.BotID, o.Exchange, o.Symbol, o.Side, t.OrderTypeLimit, t.OrderTypeMarket, t.OrderStatusCanceled)
	if o.Side == t.OrderSideBuy {
		q.Order("zone_price asc").Find(&orders)
	} else if o.Side == t.OrderSideSell {
		q.Order("zone_price desc").Find(&orders)
	}
	return orders
}

// GetLimitOrder returns the LIMIT order that is not canceled
func (d DB) GetLimitOrder(o t.Order, slippage float64) *t.Order {
	var order t.Order
	if slippage > 0 {
		lowerPrice := o.OpenPrice - o.OpenPrice*slippage
		upperPrice := o.OpenPrice + o.OpenPrice*slippage
		d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND open_price BETWEEN ? AND ? AND side = ? AND type = ? AND status <> ? AND close_time = 0",
			o.BotID, o.Exchange, o.Symbol, lowerPrice, upperPrice, o.Side, t.OrderTypeLimit, t.OrderStatusCanceled).First(&order)
	} else {
		d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND open_price = ? AND side = ? AND type = ? AND status <> ? AND close_time = 0",
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
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND status <> ? AND close_time = 0",
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeLimit, t.OrderStatusCanceled).Find(&orders)
	return orders
}

// GetLimitOrdersBySide returns the LIMIT orders that are not canceled
func (d DB) GetLimitOrdersBySide(o t.Order) []t.Order {
	var orders []t.Order
	q := d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND side = ? AND type = ? AND status <> ? AND close_time = 0",
		o.BotID, o.Exchange, o.Symbol, o.Side, t.OrderTypeLimit, t.OrderStatusCanceled)
	if o.Side == t.OrderSideBuy {
		if o.OpenTime > 0 {
			q.Order("open_time desc").Find(&orders)
		} else {
			q.Order("zone_price asc").Find(&orders)
		}
	} else if o.Side == t.OrderSideSell {
		if o.OpenTime > 0 {
			q.Order("open_time desc").Find(&orders)
		} else {
			q.Order("zone_price desc").Find(&orders)
		}
	}
	return orders
}

// GetFilledLimitBuyOrders returns the LIMIT BUY orders that their status is FILLED
func (d DB) GetFilledLimitBuyOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where(
		`bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND side = ? AND status = ? AND close_time = 0`,
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeLimit, t.OrderSideBuy, t.OrderStatusFilled).Find(&orders)
	return orders
}

// GetFilledLimitOrders returns the LIMIT LONG orders that their status is FILLED
func (d DB) GetFilledLimitLongOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where(
		`bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND pos_side = ? AND status = ? AND close_time = 0`,
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeLimit, t.OrderPosSideLong, t.OrderStatusFilled).Find(&orders)
	return orders
}

// GetFilledLimitOrders returns the LIMIT SHORT orders that their status is FILLED
func (d DB) GetFilledLimitShortOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where(
		`bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND pos_side = ? AND status = ? AND close_time = 0`,
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeLimit, t.OrderPosSideShort, t.OrderStatusFilled).Find(&orders)
	return orders
}

// GetNewOrders returns the orders that their status is NEW
func (d DB) GetNewOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND status = ?",
		o.BotID, o.Exchange, o.Symbol, t.OrderStatusNew).Find(&orders)
	return orders
}

// GetNewBuyOrders returns the BUY orders that their status is NEW
func (d DB) GetNewBuyOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND status = ? AND side = ?",
		o.BotID, o.Exchange, o.Symbol, t.OrderStatusNew, t.OrderSideBuy).Find(&orders)
	return orders
}

// GetHighestNewLongOrder returns the highest price NEW LONG order
func (d DB) GetHighestNewLongOrder(o t.Order) *t.Order {
	var orders []t.Order
	d.db.Where(
		`bot_id = ? AND exchange = ? AND symbol = ? AND pos_side = ? AND type = ? AND status = ?`,
		o.BotID, o.Exchange, o.Symbol, t.OrderPosSideLong, t.OrderTypeLimit, t.OrderStatusNew).
		Order("open_price desc").Limit(1).Find(&orders)
	if len(orders) == 0 {
		return nil
	}
	return &orders[0]
}

// GetLowestNewShortOrder returns the lowest price NEW SHORT order
func (d DB) GetLowestNewShortOrder(o t.Order) *t.Order {
	var orders []t.Order
	d.db.Where(
		`bot_id = ? AND exchange = ? AND symbol = ? AND pos_side = ? AND type = ? AND status = ?`,
		o.BotID, o.Exchange, o.Symbol, t.OrderPosSideShort, t.OrderTypeLimit, t.OrderStatusNew).
		Order("open_price asc").Limit(1).Find(&orders)
	if len(orders) == 0 {
		return nil
	}
	return &orders[0]
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
	d.db.Where("open_order_id = ? AND (type = ? OR type = ?) AND status <> ?",
		openOrderID, t.OrderTypeSL, t.OrderTypeFSL, t.OrderStatusCanceled).First(&order)
	if order.ID == "" {
		return nil
	}
	return &order
}

// GetTPOrder returns the Take Profit order of the order
func (d DB) GetTPOrder(openOrderID string) *t.Order {
	var order t.Order
	d.db.Where("open_order_id = ? AND (type = ? OR type =?) AND status <> ?",
		openOrderID, t.OrderTypeTP, t.OrderTypeFTP, t.OrderStatusCanceled).First(&order)
	if order.ID == "" {
		return nil
	}
	return &order
}

// GetSLOrders returns the STOP_LOSS_LIMIT orders that are not canceled
func (d DB) GetSLOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND status <> ? AND close_time = 0",
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeSL, t.OrderStatusCanceled).Order("open_price asc").Find(&orders)
	return orders
}

// GetTPOrders returns the TAKE_PROFIT_LIMIT orders that are not canceled
func (d DB) GetTPOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND status <> ? AND close_time = 0",
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeTP, t.OrderStatusCanceled).Order("open_price desc").Find(&orders)
	return orders
}

// GetHighestSLOrder returns the highest price SL order that is active
func (d DB) GetHighestSLOrder(o t.Order) *t.Order {
	var orders []t.Order
	d.db.Where(
		`bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND status <> ? AND close_time = 0`,
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeSL, t.OrderStatusCanceled).
		Order("open_price desc").Limit(1).Find(&orders)
	if len(orders) == 0 {
		return nil
	}
	return &orders[0]
}

// GetLowestTPOrder returns the lowest price TP order that is active
func (d DB) GetLowestTPOrder(o t.Order) *t.Order {
	var orders []t.Order
	d.db.Where(
		`bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND status <> ? AND close_time = 0`,
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeTP, t.OrderStatusCanceled).
		Order("open_price asc").Limit(1).Find(&orders)
	if len(orders) == 0 {
		return nil
	}
	return &orders[0]
}

// GetHighestSLLongOrder returns the highest price SL LONG order that is active
func (d DB) GetHighestSLLongOrder(o t.Order) *t.Order {
	var orders []t.Order
	d.db.Where(
		`bot_id = ? AND exchange = ? AND symbol = ? AND pos_side = ? AND type = ? AND status <> ? AND close_time = 0`,
		o.BotID, o.Exchange, o.Symbol, t.OrderPosSideLong, t.OrderTypeFSL, t.OrderStatusCanceled).
		Order("open_price desc").Limit(1).Find(&orders)
	if len(orders) == 0 {
		return nil
	}
	return &orders[0]
}

// GetLowestSLShortOrder returns the lowest price SL SHORT order that is active
func (d DB) GetLowestSLShortOrder(o t.Order) *t.Order {
	var orders []t.Order
	d.db.Where(
		`bot_id = ? AND exchange = ? AND symbol = ? AND pos_side = ? AND type = ? AND status <> ? AND close_time = 0`,
		o.BotID, o.Exchange, o.Symbol, t.OrderPosSideShort, t.OrderTypeFSL, t.OrderStatusCanceled).
		Order("open_price asc").Limit(1).Find(&orders)
	if len(orders) == 0 {
		return nil
	}
	return &orders[0]
}

// GetLowestTPLongOrder returns the lowest price TP LONG order that is active
func (d DB) GetLowestTPLongOrder(o t.Order) *t.Order {
	var orders []t.Order
	d.db.Where(
		`bot_id = ? AND exchange = ? AND symbol = ? AND pos_side = ? AND type = ? AND status <> ? AND close_time = 0`,
		o.BotID, o.Exchange, o.Symbol, t.OrderPosSideLong, t.OrderTypeFTP, t.OrderStatusCanceled).
		Order("open_price asc").Limit(1).Find(&orders)
	if len(orders) == 0 {
		return nil
	}
	return &orders[0]
}

// GetHighestTPShortOrder returns the highest price TP SHORT order that is active
func (d DB) GetHighestTPShortOrder(o t.Order) *t.Order {
	var orders []t.Order
	d.db.Where(
		`bot_id = ? AND exchange = ? AND symbol = ? AND pos_side = ? AND type = ? AND status <> ? AND close_time = 0`,
		o.BotID, o.Exchange, o.Symbol, t.OrderPosSideShort, t.OrderTypeFTP, t.OrderStatusCanceled).
		Order("open_price desc").Limit(1).Find(&orders)
	if len(orders) == 0 {
		return nil
	}
	return &orders[0]
}

// GetNewSLOrders returns the STOP_LOSS_LIMIT orders that their status is NEW
func (d DB) GetNewSLOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND status = ?",
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeSL, t.OrderStatusNew).Order("open_price asc").Find(&orders)
	return orders
}

// GetNewTPOrders returns the TAKE_PROFIT_LIMIT orders that their status is NEW
func (d DB) GetNewTPOrders(o t.Order) []t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND type = ? AND status = ?",
		o.BotID, o.Exchange, o.Symbol, t.OrderTypeTP, t.OrderStatusNew).Order("open_price desc").Find(&orders)
	return orders
}

// GetNearestOrder returns the nearest order of the specified order
func (d DB) GetNearestOrder(o t.Order) *t.Order {
	var orders []t.Order
	d.db.Where("bot_id = ? AND exchange = ? AND symbol = ? AND side = ? AND status <> ? AND close_time = 0",
		o.BotID, o.Exchange, o.Symbol, o.Side, t.OrderStatusCanceled).Find(&orders)
	if len(orders) == 0 {
		return nil
	}

	var norder t.Order
	for _, order := range orders {
		if norder.OpenPrice == 0 || math.Abs(order.OpenPrice-o.OpenPrice) < math.Abs(norder.OpenPrice-o.OpenPrice) {
			norder = order
		}
	}
	return &norder
}

// CreateOrder performs SQL insert on the table orders
func (d DB) CreateOrder(order t.Order) error {
	return d.db.Create(&order).Error
}

// UpdateOrder performs SQL update on the table orders
func (d DB) UpdateOrder(order t.Order) error {
	return d.db.Updates(&order).Error
}
