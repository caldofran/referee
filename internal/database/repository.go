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
	LogPriceTick(ctx context.Context, tick model.PriceTick) error
	Migrate(ctx context.Context) error
}

// PostgresRepository is the PostgreSQL implementation of the Repository.
type PostgresRepository struct {
	Pool *pgxpool.Pool
}

// LogPriceTick inserts a new price tick into the database.
func (r *PostgresRepository) LogPriceTick(ctx context.Context, tick model.PriceTick) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	query := `INSERT INTO price_ticks (timestamp, exchange, pair, bid, ask) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.Pool.Exec(ctx, query, time.Now(), tick.Exchange, tick.Pair, tick.Bid, tick.Ask)
	return err
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

// Migrate creates the necessary database tables if they do not exist.
func (r *PostgresRepository) Migrate(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Create simulated_trades table
	tradesTableQuery := `
		CREATE TABLE IF NOT EXISTS simulated_trades (
			id SERIAL PRIMARY KEY,
			timestamp TIMESTAMPTZ NOT NULL,
			trading_pair VARCHAR(20) NOT NULL,
			buy_exchange VARCHAR(50) NOT NULL,
			sell_exchange VARCHAR(50) NOT NULL,
			buy_price NUMERIC(20, 8) NOT NULL,
			sell_price NUMERIC(20, 8) NOT NULL,
			volume_eur NUMERIC(20, 8) NOT NULL,
			gross_profit_eur NUMERIC(20, 8) NOT NULL,
			total_fees_eur NUMERIC(20, 8) NOT NULL,
			net_profit_eur NUMERIC(20, 8) NOT NULL
		);`
	if _, err := r.Pool.Exec(ctx, tradesTableQuery); err != nil {
		return err
	}

	// Create price_ticks table
	ticksTableQuery := `
		CREATE TABLE IF NOT EXISTS price_ticks (
			id SERIAL PRIMARY KEY,
			timestamp TIMESTAMPTZ NOT NULL,
			exchange VARCHAR(50) NOT NULL,
			pair VARCHAR(20) NOT NULL,
			bid NUMERIC(20, 8) NOT NULL,
			ask NUMERIC(20, 8) NOT NULL
		);`
	if _, err := r.Pool.Exec(ctx, ticksTableQuery); err != nil {
		return err
	}

	return nil
}
