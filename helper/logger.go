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
	logOpen("NEW", o)
}

func LogFilledF(o *types.Order) {
	logOpen("FILLED", o)
}

func LogCanceledF(o *types.Order) {
	logOpen("CANCELED", o)
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

func logOpen(action string, o *types.Order) {
	log := types.LogOpenFOrder{
		Action:  action,
		Type:    o.Type,
		PosSide: o.PosSide,
		Qty:     o.Qty,
		Open:    o.OpenPrice,
	}
	Log(log)
}
