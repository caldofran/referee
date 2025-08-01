package arbitrage

import (
	"context"
	"log/slog"
	"referee/internal/config"
	"referee/internal/database"
	"referee/internal/model"
	"time"
)

// ArbitrageEngine holds the logic for identifying and executing arbitrage opportunities.
type ArbitrageEngine struct {
	logger       *slog.Logger
	repo         database.Repository
	cfg          *config.Config
	latestPrices map[string]model.PriceTick
}

// NewArbitrageEngine creates a new instance of the ArbitrageEngine.
func NewArbitrageEngine(logger *slog.Logger, repo database.Repository, cfg *config.Config) *ArbitrageEngine {
	return &ArbitrageEngine{
		logger:       logger,
		repo:         repo,
		cfg:          cfg,
		latestPrices: make(map[string]model.PriceTick),
	}
}

// ProcessTick processes a new price tick to check for arbitrage opportunities.
func (e *ArbitrageEngine) ProcessTick(ctx context.Context, tick model.PriceTick) {
	// Log the incoming price tick
	if err := e.repo.LogPriceTick(ctx, tick); err != nil {
		e.logger.Error("Failed to log price tick", "error", err)
	}

	// Update the latest price for this exchange
	e.latestPrices[tick.Exchange] = tick

	// Check for arbitrage opportunities with other exchanges
	for exchange, latestTick := range e.latestPrices {
		if exchange == tick.Exchange {
			continue // Skip comparing with itself
		}

		// Check if we can buy on one exchange and sell on another
		if tick.Ask < latestTick.Bid {
			// Buy on tick.Exchange, sell on exchange
			e.checkAndExecuteArbitrage(ctx, tick.Exchange, exchange, tick.Ask, latestTick.Bid)
		} else if latestTick.Ask < tick.Bid {
			// Buy on exchange, sell on tick.Exchange
			e.checkAndExecuteArbitrage(ctx, exchange, tick.Exchange, latestTick.Ask, tick.Bid)
		}
	}
}

// checkAndExecuteArbitrage checks if an arbitrage opportunity is profitable and executes it.
func (e *ArbitrageEngine) checkAndExecuteArbitrage(ctx context.Context, buyExchange, sellExchange string, buyPrice, sellPrice float64) {
	// Calculate profit using the formulas from the tech spec
	volumeInCrypto := e.cfg.Arbitrage.SimulatedTradeVolumeEUR / buyPrice
	grossProfitEUR := (sellPrice - buyPrice) * volumeInCrypto

	// Calculate fees
	buyLegFee := (buyPrice * volumeInCrypto) * (e.cfg.Exchanges[buyExchange].TakerFeePercent / 100)
	sellLegFee := (sellPrice * volumeInCrypto) * (e.cfg.Exchanges[sellExchange].TakerFeePercent / 100)
	totalFeesEUR := buyLegFee + sellLegFee + e.cfg.Arbitrage.NetworkWithdrawalFeeEUR

	// Calculate net profit
	netProfitEUR := grossProfitEUR - totalFeesEUR

	// Check if the trade is profitable
	if netProfitEUR > 0 {
		e.logger.Info("Profitable arbitrage opportunity found",
			"buyExchange", buyExchange,
			"sellExchange", sellExchange,
			"buyPrice", buyPrice,
			"sellPrice", sellPrice,
			"netProfit", netProfitEUR,
		)

		// Simulate latency before logging the trade
		time.Sleep(time.Duration(e.cfg.Arbitrage.SimulatedLatencyMS) * time.Millisecond)

		// Log the trade
		trade := model.SimulatedTrade{
			Timestamp:      time.Now(),
			TradingPair:    e.cfg.Arbitrage.TradingPair,
			BuyExchange:    buyExchange,
			SellExchange:   sellExchange,
			BuyPrice:       buyPrice,
			SellPrice:      sellPrice,
			VolumeEUR:      e.cfg.Arbitrage.SimulatedTradeVolumeEUR,
			GrossProfitEUR: grossProfitEUR,
			TotalFeesEUR:   totalFeesEUR,
			NetProfitEUR:   netProfitEUR,
		}

		if err := e.repo.LogTrade(ctx, trade); err != nil {
			e.logger.Error("Failed to log trade", "error", err)
		}
	}
}
