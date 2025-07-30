package database

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"referee/internal/model"
)

var (
	pool *pgxpool.Pool
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Define the PostgreSQL container request
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	// Create and start the PostgreSQL container
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("could not start postgres container: %s", err)
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Fatalf("could not stop postgres container: %s", err)
		}
	}()

	// Get the container's mapped port and host
	host, err := pgContainer.Host(ctx)
	if err != nil {
		log.Fatalf("could not get container host: %s", err)
	}
	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("could not get mapped port: %s", err)
	}

	// Create the database connection string
	connStr := "postgres://testuser:testpassword@" + host + ":" + port.Port() + "/testdb"

	// Create a new connection pool
	pool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("could not connect to database: %s", err)
	}
	defer pool.Close()

	// Create the table
	createTableSQL := `
	CREATE TABLE simulated_trades (
		id SERIAL PRIMARY KEY,
		timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
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
	_, err = pool.Exec(ctx, createTableSQL)
	if err != nil {
		log.Fatalf("could not create table: %s", err)
	}

	// Run the tests
	code := m.Run()

	os.Exit(code)
}

func TestPostgresRepository_LogTrade(t *testing.T) {
	ctx := context.Background()
	repo := &PostgresRepository{Pool: pool}

	trade := model.SimulatedTrade{
		Timestamp:      time.Now(),
		TradingPair:    "BTC/EUR",
		BuyExchange:    "kraken",
		SellExchange:   "binance",
		BuyPrice:       60000.0,
		SellPrice:      60100.0,
		VolumeEUR:      1000.0,
		GrossProfitEUR: 1.66666667,
		TotalFeesEUR:   1.86,
		NetProfitEUR:   -0.19333333,
	}

	err := repo.LogTrade(ctx, trade)
	assert.NoError(t, err)

	// Verify the trade was logged
	var loggedTrade model.SimulatedTrade
	err = pool.QueryRow(ctx, "SELECT trading_pair, buy_exchange, sell_exchange, buy_price, sell_price, volume_eur, gross_profit_eur, total_fees_eur, net_profit_eur FROM simulated_trades WHERE buy_exchange = 'kraken'").Scan(
		&loggedTrade.TradingPair, &loggedTrade.BuyExchange, &loggedTrade.SellExchange, &loggedTrade.BuyPrice, &loggedTrade.SellPrice, &loggedTrade.VolumeEUR, &loggedTrade.GrossProfitEUR, &loggedTrade.TotalFeesEUR, &loggedTrade.NetProfitEUR,
	)
	assert.NoError(t, err)
	assert.Equal(t, trade.TradingPair, loggedTrade.TradingPair)
	assert.Equal(t, trade.BuyExchange, loggedTrade.BuyExchange)
	assert.Equal(t, trade.SellExchange, loggedTrade.SellExchange)
}
