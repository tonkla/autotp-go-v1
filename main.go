// package main

// import (
// 	"log"

// 	"github.com/spf13/viper"

// 	"github.com/tonkla/autotp/exchange"
// )

// type cfgExchange struct {
// 	Name    string
// 	Symbols []string
// }

// func main() {
// 	viper.SetConfigFile("config.yml")
// 	err := viper.ReadInConfig()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	var cfgEx map[string][]cfgExchange
// 	viper.Unmarshal(&cfgEx)
// 	for _, c := range cfgEx["exchanges"] {
// 		ex := exchange.New(c.Name)
// 		if ex != nil {
// 			// ex.Watch(c.Symbols)
// 		} else {
// 			log.Printf("Exchange '%s' does not implement.\n", c.Name)
// 		}
// 	}
// }
package main

import (
	"github.com/tonkla/autotp/cmd"
)

func main() {
	cmd.Execute()
}
