package db

import (
	"github.com/tonkla/autotp/types"
)

// GetOrder performs SQL select on the table orders
func GetOrder(order types.Order) *types.Order {
	return nil
}

func DoesOrderExists(order *types.Order) bool {
	return false
}

// CreateOrder performs SQL insert on the table orders
func CreateOrder(order *types.Order) error {
	return nil
}

// UpdateOrder performs SQL update on the table orders
func UpdateOrder(order *types.Order) {}
