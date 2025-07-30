package exchange

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"referee/internal/model"
)

// BinanceClient implements the ExchangeClient interface for Binance.
type BinanceClient struct {
	logger *slog.Logger
}

// NewBinanceClient creates a new BinanceClient.
func NewBinanceClient(logger *slog.Logger) *BinanceClient {
	return &BinanceClient{logger: logger}
}

func (b *BinanceClient) GetName() string {
	return "binance"
}

// StartStream connects to the Binance WebSocket API and streams BTC/EUR price ticks.
func (b *BinanceClient) StartStream(ctx context.Context, priceChan chan<- model.PriceTick, pair string) error {
	const wsURL = "wss://stream.binance.com:9443/ws/btceur@ticker"
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			b.logger.Info("BinanceClient: context cancelled, shutting down")
			return nil
		default:
			b.logger.Info("BinanceClient: connecting to WebSocket", "url", wsURL, "backoff", backoff)
			c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				b.logger.Error("BinanceClient: WebSocket connection failed", "error", err)
				select {
				case <-ctx.Done():
					return nil
				case <-time.After(backoff):
					backoff *= 2
					if backoff > 16*time.Second {
						backoff = 16 * time.Second
					}
				}
				continue
			}

			// Reset backoff on successful connection
			backoff = time.Second
			b.logger.Info("BinanceClient: connected successfully")

			// Handle incoming messages
			for {
				select {
				case <-ctx.Done():
					b.logger.Info("BinanceClient: context cancelled, closing connection")
					c.Close()
					return nil
				default:
					_, message, err := c.ReadMessage()
					if err != nil {
						b.logger.Error("BinanceClient: failed to read message", "error", err)
						c.Close()
						// Break out of message loop to trigger reconnection
						break
					}

					// Parse the message
					var tickerData map[string]interface{}
					if err := json.Unmarshal(message, &tickerData); err != nil {
						b.logger.Warn("BinanceClient: failed to parse message", "error", err)
						continue
					}

					// Extract bid and ask prices from Binance ticker format
					if bidStr, ok := tickerData["b"].(string); ok {
						if askStr, ok := tickerData["a"].(string); ok {
							bid, err := strconv.ParseFloat(bidStr, 64)
							if err != nil {
								b.logger.Warn("BinanceClient: failed to parse bid price", "error", err)
								continue
							}
							ask, err := strconv.ParseFloat(askStr, 64)
							if err != nil {
								b.logger.Warn("BinanceClient: failed to parse ask price", "error", err)
								continue
							}

							// Create and send price tick
							tick := model.PriceTick{
								Exchange: "binance",
								Pair:     "BTC/EUR",
								Bid:      bid,
								Ask:      ask,
							}

							select {
							case priceChan <- tick:
								b.logger.Debug("BinanceClient: sent price tick", "bid", bid, "ask", ask)
							case <-ctx.Done():
								b.logger.Info("BinanceClient: context cancelled while sending price tick")
								c.Close()
								return nil
							}
						}
					}
				}
			}
		}
	}
}
