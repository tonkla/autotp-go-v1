# AutoTP

AutoTP aims to be an automated trading platform that supports multiple exchanges and programmable strategies (robot/EA). Inspired by MetaTrader and MQL4.

## Setup

1. Build AutoTP with `go build -o autotp main.go`.
2. Create a trading robot `kzm.go`, then compile it with `go build kzm.go`.
3. Copy a `kzm` binary into the same place of `autotp`.
4. Run AutoTP with minimum required flags: `API Key`, `Exchange`, `Symbol`, `Bot`, and `Database` (SQLite).

## Usage

```
$ autotp trade --apiKey API_KEY --secret SECRET --exchange BINANCE --symbol BNBBUSD --bot kzm --database autotp.db
$ autotp trade -k API_KEY -x SECRET -e BINANCE -s BNBBUSD -b kzm -d autotp.db
```

Or using a config file,

```
$ autotp trade --config config.yml
$ autotp trade -c config.yml
$ autotp trade -k API_KEY -x SECRET -c config.yml
```

### (Work-in-Progress) Supported Exchanges

- [Binance](https://github.com/binance/binance-spot-api-docs)
- [Bitkub](https://github.com/bitkub/bitkub-official-api-docs)
- [Satang Pro](https://docs.satangcorp.com/)

Read-only exchange,

- [SET : The Stock Exchange of Thailand](https://marketdata.set.or.th/mkt/marketsummary.do?language=en&country=US)
