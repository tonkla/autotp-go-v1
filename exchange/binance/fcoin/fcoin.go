package fcoin

import (
	"fmt"

	"github.com/tonkla/autotp/common"
)

const (
	urlBase   = "https://dapi.binance.com/dapi/v1"
	pathTrade = "/order"
)

func Trade() {
	url := fmt.Sprintf("%s%s", urlBase, pathTrade)
	data := ""
	common.Post(url, data)
}
