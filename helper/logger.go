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

func LogNew(o *types.Order) {
	logOpen("NEW", o)
}

func LogFilled(o *types.Order) {
	logOpen("FILLED", o)
}

func LogCanceled(o *types.Order) {
	logOpen("CANCELED", o)
}

func LogNewF(o *types.Order) {
	logOpenF("NEW", o)
}

func LogFilledF(o *types.Order) {
	logOpenF("FILLED", o)
}

func LogCanceledF(o *types.Order) {
	logOpenF("CANCELED", o)
}

func LogClosed(open *types.Order, close *types.Order) {
	log := types.LogCloseOrder{
		Action: "CLOSED",
		Type:   close.Type,
		Qty:    close.Qty,
		Close:  open.ClosePrice,
		Zone:   open.ZonePrice,
		Open:   open.OpenPrice,
		Profit: open.PL,
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

func logOpen(action string, o *types.Order) {
	log := types.LogOpenOrder{
		Action: action,
		Type:   o.Type,
		Qty:    o.Qty,
		Open:   o.OpenPrice,
		Zone:   o.ZonePrice,
		TP:     o.TPPrice,
	}
	Log(log)
}

func logOpenF(action string, o *types.Order) {
	log := types.LogOpenFOrder{
		Action:  action,
		Type:    o.Type,
		PosSide: o.PosSide,
		Qty:     o.Qty,
		Open:    o.OpenPrice,
	}
	Log(log)
}
