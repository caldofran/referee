package model

import "time"

// PriceTick represents a single price update from an exchange.
type PriceTick struct {
	Exchange string
	Pair     string
	Bid      float64
	Ask      float64
}

// SimulatedTrade represents a completed arbitrage trade to be logged.
type SimulatedTrade struct {
	ID             int64     `db:"id"`
	Timestamp      time.Time `db:"timestamp"`
	TradingPair    string    `db:"trading_pair"`
	BuyExchange    string    `db:"buy_exchange"`
	SellExchange   string    `db:"sell_exchange"`
	BuyPrice       float64   `db:"buy_price"`
	SellPrice      float64   `db:"sell_price"`
	VolumeEUR      float64   `db:"volume_eur"`
	GrossProfitEUR float64   `db:"gross_profit_eur"`
	TotalFeesEUR   float64   `db:"total_fees_eur"`
	NetProfitEUR   float64   `db:"net_profit_eur"`
}
