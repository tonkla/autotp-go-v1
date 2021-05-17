package fcoin

import (
	"fmt"

	"github.com/tonkla/autotp/helper"
)

const (
	urlBase   = "https://dapi.binance.com/dapi/v1"
	pathTrade = "/order"
)

func Trade() {
	url := fmt.Sprintf("%s%s", urlBase, pathTrade)
	data := ""
	helper.Post(url, data)
}
