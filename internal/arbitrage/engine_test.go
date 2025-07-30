package arbitrage

import (
	"context"
	"log/slog"
	"os"
	"referee/internal/config"
	"referee/internal/model"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) LogTrade(ctx context.Context, trade model.SimulatedTrade) error {
	args := m.Called(ctx, trade)
	return args.Error(0)
}

func TestArbitrageEngine_ProcessTick(t *testing.T) {
	mockRepo := new(MockRepository)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := &config.Config{
		Arbitrage: config.ArbitrageConfig{
			SimulatedTradeVolumeEUR: 1000.0,
			NetworkWithdrawalFeeEUR: 5.0,
			SimulatedLatencyMS:      10,
		},
		Exchanges: map[string]config.ExchangeConfig{
			"kraken":  {TakerFeePercent: 0.26},
			"binance": {TakerFeePercent: 0.1},
		},
	}

	engine := NewArbitrageEngine(logger, mockRepo, cfg)

	// Test Case 1: No opportunity
	t.Run("no opportunity", func(t *testing.T) {
		tick1 := model.PriceTick{Exchange: "kraken", Pair: "BTC/EUR", Bid: 60000, Ask: 60050}
		engine.ProcessTick(context.Background(), tick1)
		mockRepo.AssertNotCalled(t, "LogTrade")
	})

	// Test Case 2: Profitable opportunity
	t.Run("profitable opportunity", func(t *testing.T) {
		// Create a fresh engine for this test
		engine2 := NewArbitrageEngine(logger, mockRepo, cfg)
		
		// Mock the LogTrade call
		mockRepo.On("LogTrade", mock.Anything, mock.Anything).Return(nil).Once()

		// First, add Kraken price
		tick1 := model.PriceTick{Exchange: "kraken", Pair: "BTC/EUR", Bid: 60000, Ask: 60050}
		engine2.ProcessTick(context.Background(), tick1)
		
		// Then add Binance price (should create profitable opportunity)
		tick2 := model.PriceTick{Exchange: "binance", Pair: "BTC/EUR", Bid: 61000, Ask: 61050}
		engine2.ProcessTick(context.Background(), tick2)

		time.Sleep(20 * time.Millisecond) // Wait for latency simulation
		mockRepo.AssertExpectations(t)
	})

	// Test Case 3: Opportunity made unprofitable by fees
	t.Run("unprofitable due to fees", func(t *testing.T) {
		// Reset mock for this sub-test
		mockRepo.Mock = mock.Mock{}
		mockRepo.AssertNotCalled(t, "LogTrade")

		engine.latestPrices["kraken"] = model.PriceTick{Exchange: "kraken", Pair: "BTC/EUR", Bid: 60000, Ask: 60001}
		tick3 := model.PriceTick{Exchange: "binance", Pair: "BTC/EUR", Bid: 60002, Ask: 60003}
		engine.ProcessTick(context.Background(), tick3)

		mockRepo.AssertNotCalled(t, "LogTrade")
	})
}
