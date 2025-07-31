package exchange

import (
	"fmt"
	"log/slog"
	"referee/internal/config"
)

// NewClient creates a new exchange client based on the given name and configuration.
func NewClient(name string, logger *slog.Logger, cfg *config.ExchangeConfig) (ExchangeClient, error) {
	switch name {
	case "kraken":
		return NewKrakenClient(logger), nil
	case "binance":
		return NewBinanceClient(logger), nil
	default:
		return nil, fmt.Errorf("unknown exchange: %s", name)
	}
}
