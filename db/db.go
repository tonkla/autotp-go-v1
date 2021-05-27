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

// GetOrder performs SQL select on the table orders
func (d *DB) GetActiveOrder(o *types.Order, slippage float64) *types.Order {
	var order types.Order
	if slippage > 0 {
		lowerPrice := o.Price - (o.Price * slippage)
		upperPrice := o.Price + (o.Price * slippage)
		d.db.Where("exchange = ? AND symbol = ? AND price >= ? AND price <= ? AND side = ? AND status <> ?",
			o.Exchange, o.Symbol, lowerPrice, upperPrice, o.Side, types.ORDER_STATUS_CLOSED).First(&order)
	} else {
		d.db.Where("exchange = ? AND symbol = ? AND price = ? AND side = ? AND status <> ?",
			o.Exchange, o.Symbol, o.Price, o.Side, types.ORDER_STATUS_CLOSED).First(&order)
	}
	return &order
}

// IsOrderActive checks the order is active
func (d *DB) IsOrderActive(o *types.Order, slippage float64) bool {
	return d.GetActiveOrder(o, slippage).ID > 0
}

// CreateOrder performs SQL insert on the table orders
func (d *DB) CreateOrder(order *types.Order) error {
	return d.db.Create(&order).Error
}

// UpdateOrder performs SQL update on the table orders
func (d *DB) UpdateOrder(order *types.Order) error {
	return d.db.Save(&order).Error
}
