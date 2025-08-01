package arbitrage

import (
	"context"
	"log/slog"
	"os"
	"referee/internal/config"
	"referee/internal/model"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) LogTrade(ctx context.Context, trade model.SimulatedTrade) error {
	args := m.Called(ctx, trade)
	return args.Error(0)
}

func (m *MockRepository) LogPriceTick(ctx context.Context, tick model.PriceTick) error {
	args := m.Called(ctx, tick)
	return args.Error(0)
}

func (m *MockRepository) Migrate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestArbitrageEngine_CheckArbitrage(t *testing.T) {
	mockRepo := new(MockRepository)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := &config.Config{
		Arbitrage: config.ArbitrageConfig{
			SimulatedTradeVolumeEUR: 1000.0,
			NetworkWithdrawalFeeEUR: 5.0,
			SimulatedLatencyMS:      10,
			TradingPair:             "BTC/EUR",
		},
		Exchanges: map[string]config.ExchangeConfig{
			"kraken":  {TakerFeePercent: 0.26},
			"binance": {TakerFeePercent: 0.1},
		},
	}

	engine := NewArbitrageEngine(logger, mockRepo, cfg)

	t.Run("no opportunity", func(t *testing.T) {
		engine.latestPrices = map[string]model.PriceTick{
			"kraken":  {Exchange: "kraken", Pair: "BTC/EUR", Bid: 60000, Ask: 60050},
			"binance": {Exchange: "binance", Pair: "BTC/EUR", Bid: 60000, Ask: 60050},
		}
		engine.checkArbitrage(context.Background())
		mockRepo.AssertNotCalled(t, "LogTrade")
	})

	t.Run("profitable opportunity", func(t *testing.T) {
		mockRepo.On("LogTrade", mock.Anything, mock.Anything).Return(nil).Once()
		engine.latestPrices = map[string]model.PriceTick{
			"kraken":  {Exchange: "kraken", Pair: "BTC/EUR", Bid: 60000, Ask: 60050},
			"binance": {Exchange: "binance", Pair: "BTC/EUR", Bid: 61000, Ask: 61050},
		}
		engine.checkArbitrage(context.Background())
		mockRepo.AssertExpectations(t)
	})

	t.Run("unprofitable due to fees", func(t *testing.T) {
		mockRepo.Mock = mock.Mock{}
		engine.latestPrices = map[string]model.PriceTick{
			"kraken":  {Exchange: "kraken", Pair: "BTC/EUR", Bid: 60000, Ask: 60001},
			"binance": {Exchange: "binance", Pair: "BTC/EUR", Bid: 60002, Ask: 60003},
		}
		engine.checkArbitrage(context.Background())
		mockRepo.AssertNotCalled(t, "LogTrade")
	})
}
