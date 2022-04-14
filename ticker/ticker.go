package ticker

import "time"

// Ticker represents an entity that can be bought or sold.
type Ticker string

const (
	BTCUSDTicker Ticker = "BTC_USD"
)

// Bar represents a ticker buy/sell operation.
type Bar struct {
	Ticker Ticker
	Time   time.Time
	Price  string // decimal value. example: "0", "10", "12.2", "13.2345122"
}
