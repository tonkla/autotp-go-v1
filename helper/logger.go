package helper

import (
	"log"

	"github.com/tonkla/autotp/types"
)

func Log(v ...interface{}) {
	log.Printf("%+v\n", v)
}

func Logf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func LogNewF(o *types.Order) {
	log := types.LogOpenFOrder{
		Action:  "NEW",
		Type:    o.Type,
		PosSide: o.PosSide,
		Qty:     o.Qty,
		Open:    o.OpenPrice,
	}
	Log(log)
}

func LogFilledF(o *types.Order) {
	log := types.LogCloseFOrder{
		Action:  "FILLED",
		Type:    o.Type,
		PosSide: o.PosSide,
		Qty:     o.Qty,
		Open:    o.OpenPrice,
	}
	Log(log)
}

func LogCanceledF(o *types.Order) {
	log := types.LogCloseFOrder{
		Action:  "CANCELED",
		Type:    o.Type,
		PosSide: o.PosSide,
		Qty:     o.Qty,
		Open:    o.OpenPrice,
	}
	Log(log)
}

func LogClosedF(open *types.Order, close *types.Order) {
	log := types.LogCloseFOrder{
		Action:  "CLOSED",
		Type:    close.Type,
		PosSide: close.PosSide,
		Qty:     close.Qty,
		Close:   open.ClosePrice,
		Open:    open.OpenPrice,
		Profit:  open.PL,
	}
	Log(log)
}
