package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "autotp",
	Short: "AutoTP server",
	Long:  "AutoTP server",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var tradeCmd = &cobra.Command{
	Use:   "trade",
	Short: "Trade with your trusted robot",
	Long:  `Trade with your trusted robot`,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey := cmd.Flags().Lookup("apiKey").Value
		exchange := cmd.Flags().Lookup("exchange").Value
		symbol := cmd.Flags().Lookup("symbol").Value
		bot := cmd.Flags().Lookup("bot").Value
		// database := cmd.Flags().Lookup("database").Value
		// autotp.Trade()
		fmt.Println(apiKey, exchange, symbol, bot)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of AutoTP",
	Long:  `All software has versions. This is AutoTP's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("AutoTP v0.1")
	},
}

func init() {
	var apiKey string
	var exchange string
	var symbol string
	var bot string
	// var database string

	tradeCmd.Flags().StringVarP(&apiKey, "apiKey", "k", "", "API Key (required)")
	tradeCmd.MarkFlagRequired("apiKey")
	tradeCmd.Flags().StringVarP(&exchange, "exchange", "e", "", "Exchange (required)")
	tradeCmd.MarkFlagRequired("exchange")
	tradeCmd.Flags().StringVarP(&symbol, "symbol", "s", "", "Symbol (required)")
	tradeCmd.MarkFlagRequired("symbol")
	tradeCmd.Flags().StringVarP(&bot, "bot", "b", "", "Bot (required)")
	tradeCmd.MarkFlagRequired("bot")
	// tradeCmd.Flags().StringVarP(&database, "database", "d", "", "SQLite Database (required)")
	// tradeCmd.MarkFlagRequired("database")

	rootCmd.AddCommand(tradeCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
