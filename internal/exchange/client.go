package exchange

import (
	"context"
	"referee/internal/model"
)

// ExchangeClient defines the standard interface for all exchange clients.
type ExchangeClient interface {
	GetName() string
	StartStream(ctx context.Context, priceChan chan<- model.PriceTick, pair string) error
}
