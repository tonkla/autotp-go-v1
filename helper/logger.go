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
	l := types.LogOpenOrder{
		Action: "NEW",
		Type:   o.Type,
		Qty:    o.Qty,
		Open:   o.OpenPrice,
		Zone:   o.ZonePrice,
	}
	Logf("\n")
	Log(l)
}

func LogNewF(o types.Order) {
	l := types.LogOpenFOrder{
		Action:  "NEW",
		Type:    o.Type,
		PosSide: o.PosSide,
		Qty:     o.Qty,
		Open:    o.OpenPrice,
	}
	Logf("\n")
	Log(l)
}

func LogFilled(o types.Order) {
	l := types.LogOpenOrder{
		Action: "FILLED",
		Type:   o.Type,
		Qty:    o.Qty,
		Open:   o.OpenPrice,
		Zone:   o.ZonePrice,
	}
	Log(l)
}

func LogFilledF(o types.Order) {
	l := types.LogOpenFOrder{
		Action:  "FILLED",
		Type:    o.Type,
		PosSide: o.PosSide,
		Qty:     o.Qty,
		Open:    o.OpenPrice,
	}
	Log(l)
}

func LogCanceled(o types.Order) {
	l := types.LogOpenOrder{
		Action: "CANCELED",
		Type:   o.Type,
		Qty:    o.Qty,
		Open:   o.OpenPrice,
		Zone:   o.ZonePrice,
	}
	Log(l)
}

func LogCanceledF(o types.Order) {
	l := types.LogOpenFOrder{
		Action:  "CANCELED",
		Type:    o.Type,
		PosSide: o.PosSide,
		Qty:     o.Qty,
		Open:    o.OpenPrice,
	}
	Log(l)
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
	Log(l)
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
	Log(l)
}
