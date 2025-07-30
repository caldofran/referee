package database

import (
	"context"
	"referee/internal/model"
)

// Repository defines the standard interface for database operations.
type Repository interface {
	LogTrade(ctx context.Context, trade model.SimulatedTrade) error
}
