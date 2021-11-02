# AutoTP

AutoTP aims to be an automated trading platform that supports various exchanges. Inspired by MetaTrader and [MQL4](https://docs.mql4.com/).

## Usage

1. Craft your trading strategy, put it inside `strategy/`
2. Add the strategy into `strategy/strategy.go`
3. Compile it with `go build -o autotp main.go`
4. Copy `config.yml.example` to `config.yml`, configure your preferred parameters
5. Run `./autotp -c config.yml`, or `./monit` for infinite running until the world ends

### (Work-in-Progress) Supported Exchanges

- [Binance](https://github.com/binance/binance-spot-api-docs)
- [FTX](https://docs.ftx.us/)
- [Bitkub](https://github.com/bitkub/bitkub-official-api-docs)
- [Satang Pro](https://docs.satangcorp.com/)

Read-only exchange,

- [SET: The Stock Exchange of Thailand](https://marketdata.set.or.th/mkt/marketsummary.do?language=en&country=US)
