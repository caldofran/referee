package exchange

import (
	"context"
	"log/slog"
	"time"

	"encoding/json"
	"github.com/gorilla/websocket"
	"referee/internal/model"
	"strconv"
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
	const wsURL = "wss://ws.kraken.com"
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			k.logger.Info("KrakenClient: context cancelled, shutting down")
			return nil
		default:
			k.logger.Info("KrakenClient: connecting to WebSocket", "url", wsURL, "backoff", backoff)
			c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				k.logger.Error("KrakenClient: WebSocket connection failed", "error", err)
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

			// Send subscription message for BTC/EUR ticker
			subscription := map[string]interface{}{
				"event": "subscribe",
				"pair":  []string{"XBT/EUR"},
				"subscription": map[string]string{
					"name": "ticker",
				},
			}
							if err := c.WriteJSON(subscription); err != nil {
					k.logger.Error("KrakenClient: failed to send subscription", "error", err)
					if closeErr := c.Close(); closeErr != nil {
						k.logger.Warn("KrakenClient: failed to close connection", "error", closeErr)
					}
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
			k.logger.Info("KrakenClient: subscription sent successfully")

			// Handle incoming messages
			for {
				select {
				case <-ctx.Done():
					k.logger.Info("KrakenClient: context cancelled, closing connection")
					if closeErr := c.Close(); closeErr != nil {
						k.logger.Warn("KrakenClient: failed to close connection", "error", closeErr)
					}
					return nil
				default:
					_, message, err := c.ReadMessage()
					if err != nil {
						k.logger.Error("KrakenClient: failed to read message", "error", err)
						if closeErr := c.Close(); closeErr != nil {
							k.logger.Warn("KrakenClient: failed to close connection", "error", closeErr)
						}
						// Break out of message loop to trigger reconnection
						break
					}

					// Parse the message - Kraken sends both objects and arrays
					var msgObj map[string]interface{}
					var msgArray []interface{}
					
					// Try to parse as object first (for subscription confirmations)
					if err := json.Unmarshal(message, &msgObj); err == nil {
						// Handle subscription confirmation
						if event, ok := msgObj["event"].(string); ok && event == "subscriptionStatus" {
							k.logger.Info("KrakenClient: subscription confirmed")
							continue
						}
						// If it's an object but not a subscription confirmation, skip it
						continue
					}
					
					// Try to parse as array (for ticker data: [channelID, tickerData, pair, channelName])
					if err := json.Unmarshal(message, &msgArray); err != nil {
						k.logger.Warn("KrakenClient: failed to parse message", "error", err)
						continue
					}
					
					// Check if it's a ticker array with at least 2 elements
					if len(msgArray) >= 2 {
						if tickerData, ok := msgArray[1].(map[string]interface{}); ok {
							// Extract bid and ask prices
							if bidStr, ok := tickerData["b"].([]interface{}); ok && len(bidStr) > 0 {
								if askStr, ok := tickerData["a"].([]interface{}); ok && len(askStr) > 0 {
									bid, err := strconv.ParseFloat(bidStr[0].(string), 64)
									if err != nil {
										k.logger.Warn("KrakenClient: failed to parse bid price", "error", err)
										continue
									}
									ask, err := strconv.ParseFloat(askStr[0].(string), 64)
									if err != nil {
										k.logger.Warn("KrakenClient: failed to parse ask price", "error", err)
										continue
									}

									// Create and send price tick
									tick := model.PriceTick{
										Exchange: "kraken",
										Pair:     "BTC/EUR",
										Bid:      bid,
										Ask:      ask,
									}

									select {
									case priceChan <- tick:
										k.logger.Debug("KrakenClient: sent price tick", "bid", bid, "ask", ask)
									case <-ctx.Done():
										k.logger.Info("KrakenClient: context cancelled while sending price tick")
										if closeErr := c.Close(); closeErr != nil {
											k.logger.Warn("KrakenClient: failed to close connection", "error", closeErr)
										}
										return nil
									}
								}
							}
						}
					}
				}
			}
		}
	}
}
