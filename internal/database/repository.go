package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"referee/internal/model"
)

// Repository defines the standard interface for database operations.
type Repository interface {
	LogTrade(ctx context.Context, trade model.SimulatedTrade) error
}

// PostgresRepository is the PostgreSQL implementation of the Repository.
type PostgresRepository struct {
	Pool *pgxpool.Pool
}

// LogTrade inserts a new simulated trade into the database.
func (r *PostgresRepository) LogTrade(ctx context.Context, trade model.SimulatedTrade) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO simulated_trades (
			timestamp, trading_pair, buy_exchange, sell_exchange, buy_price,
			sell_price, volume_eur, gross_profit_eur, total_fees_eur, net_profit_eur
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.Pool.Exec(ctx, query,
		trade.Timestamp,
		trade.TradingPair,
		trade.BuyExchange,
		trade.SellExchange,
		trade.BuyPrice,
		trade.SellPrice,
		trade.VolumeEUR,
		trade.GrossProfitEUR,
		trade.TotalFeesEUR,
		trade.NetProfitEUR,
	)

	return err
}
