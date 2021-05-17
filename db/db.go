package db

import "github.com/tonkla/autotp/types"

// GetRecord performs SQL select on the table records
func GetRecord(id string) *types.Record {
	return nil
}

// GetRecordByPrice performs SQL select on the table records by target price
func GetRecordByPrice(price float64, side string) *types.Record {
	return nil
}

// CreateRecord performs SQL insert on the table records
func CreateRecord(result types.TradeResult) error {
	return nil
}

// UpdateRecord performs SQL update on the table records
func UpdateRecord(id string) {}
