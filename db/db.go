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
func (d *DB) GetOrder(order types.Order) *types.Order {
	return nil
}

// DoesOrderExists checks the order existence
func (d *DB) DoesOrderExists(o *types.Order) bool {
	var order types.Order
	d.db.Where("exchange = ? AND symbol = ? AND price = ? AND side = ? AND status <> ?",
		o.Exchange, o.Symbol, o.Price, o.Side, types.ORDER_STATUS_CLOSED).First(&order)
	return order.ID != 0
}

// CreateOrder performs SQL insert on the table orders
func (d *DB) CreateOrder(order *types.Order) error {
	return d.db.Create(&order).Error
}

// UpdateOrder performs SQL update on the table orders
func (d *DB) UpdateOrder(order *types.Order) error {
	return nil
}
