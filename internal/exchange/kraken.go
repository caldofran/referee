package exchange

import (
	"context"
	"log/slog"
	"time"

	"referee/internal/model"
)

// KrakenClient implements the ExchangeClient interface for Kraken.
type KrakenClient struct {
	logger *slog.Logger
}

// NewKrakenClient creates a new KrakenClient.
func NewKrakenClient(logger *slog.Logger) *KrakenClient {
	return &KrakenClient{logger: logger}
}

func (k *KrakenClient) GetName() string {
	return "kraken"
}

// StartStream connects to the Kraken WebSocket API and streams BTC/EUR price ticks.
func (k *KrakenClient) StartStream(ctx context.Context, priceChan chan<- model.PriceTick, pair string) error {
	// Placeholder: Connect to Kraken WebSocket, subscribe to ticker, handle reconnection with backoff
	// Use exponential backoff for reconnection attempts
	// Respect ctx for graceful shutdown
	// Parse messages and send model.PriceTick to priceChan
	// Log errors using k.logger
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			k.logger.Info("KrakenClient: context cancelled, shutting down")
			return nil
		default:
			// TODO: Connect and handle messages
			k.logger.Info("KrakenClient: connecting to WebSocket (placeholder)")
			// Simulate a connection attempt and error
			time.Sleep(backoff)
			backoff *= 2
			if backoff > 16*time.Second {
				backoff = 16 * time.Second
			}
		}
	}
}