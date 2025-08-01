package arbitrage

import (
	"context"
	"log/slog"
	"referee/internal/config"
	"referee/internal/database"
	"referee/internal/model"
	"sync"
	"time"
)

// ArbitrageEngine holds the logic for identifying and executing arbitrage opportunities.
type ArbitrageEngine struct {
	logger         *slog.Logger
	repo           database.Repository
	cfg            *config.Config
	latestPrices   map[string]model.PriceTick
	priceMutex     sync.RWMutex
	checkInterval  time.Duration
}

// NewArbitrageEngine creates a new instance of the ArbitrageEngine.
func NewArbitrageEngine(logger *slog.Logger, repo database.Repository, cfg *config.Config) *ArbitrageEngine {
	return &ArbitrageEngine{
		logger:        logger,
		repo:          repo,
		cfg:           cfg,
		latestPrices:  make(map[string]model.PriceTick),
		checkInterval: time.Duration(cfg.Arbitrage.CheckIntervalMS) * time.Millisecond,
	}
}

// Start begins the arbitrage checking loop.
func (e *ArbitrageEngine) Start(ctx context.Context) {
	ticker := time.NewTicker(e.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.checkArbitrage(ctx)
		}
	}
}

// ProcessTick processes a new price tick to check for arbitrage opportunities.
func (e *ArbitrageEngine) ProcessTick(ctx context.Context, tick model.PriceTick) {
	e.priceMutex.Lock()
	defer e.priceMutex.Unlock()
	e.latestPrices[tick.Exchange] = tick
}

// checkArbitrage checks for arbitrage opportunities between all exchanges.
func (e *ArbitrageEngine) checkArbitrage(ctx context.Context) {
	e.priceMutex.RLock()
	defer e.priceMutex.RUnlock()

	exchanges := make([]string, 0, len(e.latestPrices))
	for ex := range e.latestPrices {
		exchanges = append(exchanges, ex)
	}

	for i := 0; i < len(exchanges); i++ {
		for j := i + 1; j < len(exchanges); j++ {
			ex1 := exchanges[i]
			ex2 := exchanges[j]

			price1 := e.latestPrices[ex1]
			price2 := e.latestPrices[ex2]

			if price1.Ask < price2.Bid {
				e.evaluateAndExecute(ctx, ex1, ex2, price1.Ask, price2.Bid)
			}
			if price2.Ask < price1.Bid {
				e.evaluateAndExecute(ctx, ex2, ex1, price2.Ask, price1.Bid)
			}
		}
	}
}

// evaluateAndExecute checks if an arbitrage opportunity is profitable and executes it.
func (e *ArbitrageEngine) evaluateAndExecute(ctx context.Context, buyExchange, sellExchange string, buyPrice, sellPrice float64) {
	volumeInCrypto := e.cfg.Arbitrage.SimulatedTradeVolumeEUR / buyPrice
	grossProfitEUR := (sellPrice - buyPrice) * volumeInCrypto

	buyLegFee := (buyPrice * volumeInCrypto) * (e.cfg.Exchanges[buyExchange].TakerFeePercent / 100)
	sellLegFee := (sellPrice * volumeInCrypto) * (e.cfg.Exchanges[sellExchange].TakerFeePercent / 100)
	totalFeesEUR := buyLegFee + sellLegFee + e.cfg.Arbitrage.NetworkWithdrawalFeeEUR

	netProfitEUR := grossProfitEUR - totalFeesEUR

	if netProfitEUR > 0 {
		e.logger.Info("Profitable arbitrage opportunity found",
			"buyExchange", buyExchange,
			"sellExchange", sellExchange,
			"buyPrice", buyPrice,
			"sellPrice", sellPrice,
			"netProfit", netProfitEUR,
		)

		time.Sleep(time.Duration(e.cfg.Arbitrage.SimulatedLatencyMS) * time.Millisecond)

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
