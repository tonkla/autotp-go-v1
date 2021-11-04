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

func LogNew(o types.Order) {
	logOpen("NEW", o)
}

func LogFilled(o types.Order) {
	logOpen("FILLED", o)
}

func LogCanceled(o types.Order) {
	l := types.LogOpenOrder{
		Action: "CANCELED",
		Type:   o.Type,
		Qty:    o.Qty,
		Open:   o.OpenPrice,
		Zone:   o.ZonePrice,
	}
	log.Printf("%+v\n\n", l)
}

func LogNewF(o types.Order) {
	logOpenF("NEW", o)
}

func LogFilledF(o types.Order) {
	logOpenF("FILLED", o)
}

func LogCanceledF(o types.Order) {
	l := types.LogOpenFOrder{
		Action:  "CANCELED",
		Type:    o.Type,
		PosSide: o.PosSide,
		Qty:     o.Qty,
		Open:    o.OpenPrice,
	}
	log.Printf("%+v\n\n", l)
}

func LogClosed(open types.Order, close types.Order) {
	l := types.LogCloseOrder{
		Action: "CLOSED",
		Type:   close.Type,
		Qty:    close.Qty,
		Zone:   open.ZonePrice,
		Open:   open.OpenPrice,
		Close:  open.ClosePrice,
		Profit: open.PL,
	}
	log.Printf("%+v\n\n", l)
}

func LogClosedF(open types.Order, close types.Order) {
	l := types.LogCloseFOrder{
		Action:  "CLOSED",
		Type:    close.Type,
		PosSide: close.PosSide,
		Qty:     close.Qty,
		Open:    open.OpenPrice,
		Close:   open.ClosePrice,
		Profit:  open.PL,
	}
	log.Printf("%+v\n\n", l)
}

func logOpen(action string, o types.Order) {
	l := types.LogOpenOrder{
		Action: action,
		Type:   o.Type,
		Qty:    o.Qty,
		Open:   o.OpenPrice,
		Zone:   o.ZonePrice,
	}
	Log(l)
}

func logOpenF(action string, o types.Order) {
	l := types.LogOpenFOrder{
		Action:  action,
		Type:    o.Type,
		PosSide: o.PosSide,
		Qty:     o.Qty,
		Open:    o.OpenPrice,
	}
	Log(l)
}
