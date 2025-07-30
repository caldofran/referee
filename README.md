# Referee - Cryptocurrency Arbitrage Simulation Bot

Referee is a professional-grade, high-performance cryptocurrency arbitrage simulation bot that operates on a local machine, tests strategies against live exchange data with a realistic cost model, and logs all results to a local database.

## Features

- **Real-time Price Streaming**: Connects to multiple cryptocurrency exchanges via WebSocket
- **Arbitrage Detection**: Identifies profitable trading opportunities across exchanges
- **Realistic Simulation**: Includes trading fees, network costs, and execution latency
- **Data Persistence**: Logs all simulated trades to PostgreSQL database
- **Visualization**: Metabase integration for data analysis and dashboards
- **Resilient Architecture**: Automatic reconnection with exponential backoff
- **Graceful Shutdown**: Proper signal handling and context cancellation

## Technology Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL with pgx/v5 driver
- **WebSocket**: gorilla/websocket for exchange connections
- **Configuration**: Viper for config management
- **Logging**: Structured logging with slog
- **Containerization**: Docker with multi-stage builds
- **Visualization**: Metabase for data analysis

## Quick Start

### Prerequisites

- Go 1.21 or newer
- Docker and Docker Compose
- Git

### Setup

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd Referee
   ```

2. **Start the infrastructure**:
   ```bash
   make docker-up
   ```

3. **Create configuration**:
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your settings
   ```

4. **Build and run**:
   ```bash
   make build
   make run
   ```

5. **Access Metabase** (optional):
   - Open http://localhost:3000
   - Follow the setup guide in `METABASE_SETUP.md`

## Configuration

The application uses `config.yaml` for configuration. Key settings:

```yaml
arbitrage:
  simulated_trade_volume_eur: 1000.0
  network_withdrawal_fee_eur: 5.0
  simulated_latency_ms: 50

database:
  host: "postgres"
  port: 5432
  user: "user"
  password: "password"
  dbname: "referee_sim"

exchanges:
  kraken:
    taker_fee_percent: 0.26
  binance:
    taker_fee_percent: 0.1
```

## Architecture

### Core Components

- **Arbitrage Engine**: Processes price ticks and identifies profitable opportunities
- **Exchange Clients**: WebSocket connections to cryptocurrency exchanges
- **Database Repository**: PostgreSQL interface for trade logging
- **Configuration Manager**: Viper-based config loading with environment variable support

### Data Flow

1. Exchange clients stream real-time price data via WebSocket
2. Price ticks are sent to a single channel (fan-in pattern)
3. Arbitrage engine processes each tick and identifies opportunities
4. Profitable trades are logged to the database with latency simulation
5. Metabase provides real-time visualization of the data

## Adding New Exchanges

To add a new exchange (e.g., Coinbase), follow these steps:

### 1. Create Exchange Client

Create a new file `internal/exchange/coinbase.go`:

```go
package exchange

import (
    "context"
    "log/slog"
    "time"
    "referee/internal/model"
    "github.com/gorilla/websocket"
)

type CoinbaseClient struct {
    logger *slog.Logger
}

func NewCoinbaseClient(logger *slog.Logger) *CoinbaseClient {
    return &CoinbaseClient{logger: logger}
}

func (c *CoinbaseClient) GetName() string {
    return "coinbase"
}

func (c *CoinbaseClient) StartStream(ctx context.Context, priceChan chan<- model.PriceTick, pair string) error {
    // Implement WebSocket connection to Coinbase
    // Follow the same pattern as KrakenClient and BinanceClient
    // Include resilient reconnection with exponential backoff
    // Parse messages and send model.PriceTick to priceChan
    // Respect ctx for graceful shutdown
    return nil
}
```

### 2. Add Configuration

Add the exchange configuration to `config.yaml`:

```yaml
exchanges:
  kraken:
    taker_fee_percent: 0.26
  binance:
    taker_fee_percent: 0.1
  coinbase:
    taker_fee_percent: 0.5  # Add appropriate fee
```

### 3. Register in Main Application

Add the new client to the clients list in `cmd/referee/main.go`:

```go
clients := []exchange.ExchangeClient{
    exchange.NewKrakenClient(logger),
    exchange.NewBinanceClient(logger),
    exchange.NewCoinbaseClient(logger),  // Add this line
}
```

### 4. Test the Implementation

```bash
make test
make build
make run
```

### Exchange Client Requirements

Your exchange client must:

- ✅ Implement the `ExchangeClient` interface
- ✅ Connect to the exchange's WebSocket API
- ✅ Subscribe to the appropriate trading pair (BTC/EUR)
- ✅ Parse incoming messages and extract bid/ask prices
- ✅ Send `model.PriceTick` objects to the provided channel
- ✅ Implement resilient reconnection with exponential backoff
- ✅ Respect context cancellation for graceful shutdown
- ✅ Log errors and important events using the provided logger

## Development

### Project Structure

```
referee/
├── cmd/referee/          # Main application entry point
├── internal/
│   ├── arbitrage/        # Core arbitrage logic
│   ├── config/           # Configuration management
│   ├── database/         # Database repository
│   ├── exchange/         # Exchange client implementations
│   └── model/            # Data models
├── pkg/                  # Public libraries (if needed)
├── config.example.yaml   # Configuration template
├── docker-compose.yml    # Infrastructure services
├── Dockerfile           # Application containerization
└── Makefile             # Development tasks
```

### Available Make Commands

```bash
make build      # Build the application
make test       # Run tests with coverage
make run        # Build and run the application
make clean      # Clean build artifacts
make docker-up  # Start infrastructure services
make docker-down # Stop infrastructure services
```

### Testing

```bash
# Run all tests
make test

# Run specific test
go test ./internal/arbitrage

# Run with coverage
go test -cover ./...
```

## Monitoring and Analysis

### Database Queries

Connect to the database to analyze performance:

```bash
docker exec -it referee_db psql -U user -d referee_sim
```

### Key Metrics

- **Total Trades**: `SELECT COUNT(*) FROM simulated_trades;`
- **Win Rate**: `SELECT COUNT(CASE WHEN net_profit_eur > 0 THEN 1 END) * 100.0 / COUNT(*) FROM simulated_trades;`
- **Total Profit**: `SELECT SUM(net_profit_eur) FROM simulated_trades;`
- **Best Exchange Pair**: `SELECT buy_exchange, sell_exchange, SUM(net_profit_eur) FROM simulated_trades GROUP BY buy_exchange, sell_exchange ORDER BY SUM(net_profit_eur) DESC;`

## Troubleshooting

### Common Issues

1. **WebSocket Connection Failures**:
   - Check internet connectivity
   - Verify exchange API endpoints are accessible
   - Review exchange rate limits

2. **Database Connection Issues**:
   - Ensure Docker services are running: `docker-compose ps`
   - Check database logs: `docker-compose logs postgres`

3. **No Arbitrage Opportunities**:
   - Verify exchange fees are configured correctly
   - Check if price spreads are sufficient to cover costs
   - Review network withdrawal fee settings

### Logs

The application uses structured JSON logging. Key log levels:
- `INFO`: Normal operation events
- `WARN`: Recoverable issues
- `ERROR`: Connection failures, parsing errors
- `DEBUG`: Detailed price tick information

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

[Add your license information here]

## Support

For issues and questions:
1. Check the troubleshooting section
2. Review the logs for error details
3. Open an issue with detailed information 