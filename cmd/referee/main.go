package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"referee/internal/arbitrage"
	"referee/internal/config"
	"referee/internal/database"
	"referee/internal/exchange"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
	"referee/internal/model"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	logger.Info("Starting Referee arbitrage simulation bot")

	// Load configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	logger.Info("Configuration loaded successfully")

	// Create database connection pool
	pool, err := pgxpool.New(context.Background(), cfg.Database.DSN())
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	logger.Info("Database connection established")

	// Create repository
	repo := &database.PostgresRepository{Pool: pool}

	// Create arbitrage engine
	engine := arbitrage.NewArbitrageEngine(logger, repo, &cfg)
	logger.Info("Arbitrage engine initialized")

	// Create exchange clients based on configuration
	clients := make([]exchange.ExchangeClient, 0, len(cfg.Exchanges))
	for name, exchangeCfg := range cfg.Exchanges {
		client, err := exchange.NewClient(name, logger, &exchangeCfg)
		if err != nil {
			logger.Error("Failed to create exchange client", "exchange", name, "error", err)
			os.Exit(1)
		}
		clients = append(clients, client)
	}
	logger.Info("Exchange clients created", "count", len(clients))

	// Set up context for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Use an errgroup to manage goroutines
	eg, gCtx := errgroup.WithContext(ctx)

	// Create the fan-in channel for price ticks
	priceChan := make(chan model.PriceTick, 100)

	// Start the arbitrage engine goroutine
	eg.Go(func() error {
		logger.Info("Starting arbitrage engine")
		for {
			select {
			case <-gCtx.Done():
				logger.Info("Arbitrage engine shutting down")
				return gCtx.Err()
			case tick := <-priceChan:
				engine.ProcessTick(gCtx, tick)
			}
		}
	})

	// Start all exchange clients in goroutines
	for _, client := range clients {
		c := client // capture range variable
		eg.Go(func() error {
			logger.Info("Starting exchange client", "exchange", c.GetName())
			if err := c.StartStream(gCtx, priceChan, cfg.Arbitrage.TradingPair); err != nil {
				logger.Error("Exchange client error", "exchange", c.GetName(), "error", err)
				return err
			}
			return nil
		})
	}

	// Wait for shutdown signal or an error from a goroutine
	logger.Info("Referee is running. Press Ctrl+C to stop.")
	if err := eg.Wait(); err != nil && err != context.Canceled {
		logger.Error("Application error", "error", err)
	}

	logger.Info("Graceful shutdown completed")
}
