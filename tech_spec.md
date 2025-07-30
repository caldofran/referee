Of course. Here is the definitive, all-inclusive technical specification for **Project Referee**.

This document is designed to be the single source of truth that a developer with no prior context can use to understand and build the application from scratch. It consolidates all previous specifications, Q&A clarifications, and implementation details into one logical flow.

---

### **Technical Specification: Project Referee (Version 1.3 - Definitive Build Spec)**

**1. Project Mission & Name**

*   **Project Name:** **Referee**
*   **Mission:** To build a professional-grade, high-performance cryptocurrency arbitrage **simulation bot**. Referee will operate on a local machine, test strategies against live exchange data with a realistic cost model, and log all results to a local database. Its success will be measured by its stability, the accuracy of its simulations, and the clarity of the data it provides for analysis.
*   **Initial Scope:** The initial build will focus exclusively on validating the core strategy for the **`BTC/EUR`** trading pair, using the highest liquidity providers. The design must be extensible, but the implementation will be limited to this pair.

**2. Core Principles & Best Practices**

*   **Standard Project Layout:** The project will use the standard Go project layout (`cmd/`, `internal/`, `pkg/`) to ensure a clean separation of concerns.
*   **Dependency Injection (DI):** Dependencies (like database connections and loggers) will be initialized in `main` and injected into the components that need them.
*   **Context for Graceful Shutdown:** The application will use Go's `context` package to manage its lifecycle, ensuring clean shutdowns on OS interrupt signals (Ctrl+C).
*   **Structured Logging:** All log output will use Go's standard structured logging library, `slog`, configured for JSON output.
*   **Comprehensive Testing:** The goal is the highest test coverage possible, using interfaces, mocks for unit tests, and container-based integration tests.

**3. Technology Stack**

*   **Language:** **Go** (version 1.21 or newer)
*   **Project Name / Go Module:** `referee`
*   **Configuration Management:** **Viper** (`spf13/viper`)
*   **API Interaction:** **`gorilla/websocket`**
*   **Logging:** **`slog`** (standard library)
*   **Database:** **PostgreSQL** (latest stable version)
*   **Go Database Driver:** **`pgx/v5`**
*   **Testing Framework:** Go's built-in testing package + **`testify/assert`** & **`testify/mock`**.
*   **Environment Setup:** **Docker** and **Docker Compose**.
*   **Visualization Tool:** **Metabase**.

**4. Configuration Model (`config.yaml`)**

Configuration will be managed by Viper. It will load from a `config.yaml` file and allow overrides from environment variables for security.

*   **File Structure (`config.example.yaml`):**
    ```yaml
    # Configuration for the arbitrage simulation
    arbitrage:
      # The fixed trade size for every simulated trade, in EUR.
      simulated_trade_volume_eur: 1000.0
      # A constant fee representing the cost of moving assets between exchanges.
      network_withdrawal_fee_eur: 5.0
      # A delay in milliseconds to simulate network and execution latency.
      simulated_latency_ms: 50

    # PostgreSQL database connection details.
    # IMPORTANT: Use environment variables for sensitive values in production.
    database:
      host: "postgres" # Docker service name
      port: 5432
      user: "user"
      # For local dev, you can set password here.
      # For production/git, set this via environment variable: REFEREE_DB_PASSWORD
      password: "password"
      dbname: "referee_sim"

    # Per-exchange specific settings.
    # The key (e.g., "kraken") must match the exchange name returned by the client.
    exchanges:
      kraken:
        taker_fee_percent: 0.26
      binance:
        taker_fee_percent: 0.1
    ```
*   **Secret Management:** Viper must be configured to automatically bind to environment variables (e.g., `REFEREE_DB_PASSWORD`), which will always take precedence over the values in the YAML file. This prevents secrets from being committed to version control.

**5. Project Structure (Standard Go Layout)**

```
referee/
├── cmd/
│   └── referee/
│       └── main.go         # Entry point, DI container, signal handling
├── internal/
│   ├── arbitrage/
│   │   ├── engine.go       # Core arbitrage logic, state management
│   │   └── engine_test.go
│   ├── config/             # Configuration loading (Viper)
│   ├── database/           # PostgreSQL repository (interface and implementation)
│   ├── exchange/           # Exchange client implementations
│   │   ├── client.go       # Defines the common ExchangeClient interface
│   │   ├── binance.go
│   │   └── kraken.go
│   └── model/              # Defines core data models (PriceTick, Trade, etc.)
├── pkg/
│   └── (empty for now)     # For public-facing libraries if needed later
├── go.mod
├── go.sum
├── Makefile                # For developer task automation
├── Dockerfile              # For building the production container
├── config.example.yaml
└── docker-compose.yml
```

**6. Core Interfaces & Models**

*   **`internal/model/models.go`:** Defines the core data structs.
    ```go
    package model

    import "time"

    // PriceTick represents a single price update from an exchange.
    type PriceTick struct {
        Exchange string
        Pair     string
        Bid      float64
        Ask      float64
    }

    // SimulatedTrade represents a completed arbitrage trade to be logged.
    type SimulatedTrade struct {
        ID              int64     `db:"id"`
        Timestamp       time.Time `db:"timestamp"`
        TradingPair     string    `db:"trading_pair"`
        BuyExchange     string    `db:"buy_exchange"`
        SellExchange    string    `db:"sell_exchange"`
        BuyPrice        float64   `db:"buy_price"`
        SellPrice       float64   `db:"sell_price"`
        VolumeEUR       float64   `db:"volume_eur"`
        GrossProfitEUR  float64   `db:"gross_profit_eur"`
        TotalFeesEUR    float64   `db:"total_fees_eur"`
        NetProfitEUR    float64   `db:"net_profit_eur"`
    }
    ```
*   **`internal/exchange/client.go`:** Defines the `ExchangeClient` interface.
    ```go
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
    ```
*   **`internal/database/repository.go`:** Defines the `Repository` interface.
    ```go
    package database

    import (
        "context"
        "referee/internal/model"
    )

    // Repository defines the standard interface for database operations.
    type Repository interface {
        LogTrade(ctx context.Context, trade model.SimulatedTrade) error
    }
    ```

**7. Database Schema**

The PostgreSQL database will contain a single table, `simulated_trades`, created with the following SQL statement.

```sql
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
);
```

**8. Core Logic & Implementation Details**

*   **`internal/arbitrage/engine.go` - Arbitrage Engine**
    *   **State Management:** The engine will maintain an internal, in-memory `map[string]model.PriceTick` to store the latest price from each exchange. When a new tick arrives, the engine updates this map and then compares the new tick against the stored prices of all other exchanges to find opportunities.
    *   **Profit Calculation:** A trade is profitable if the gross profit exceeds the combined fees. The formulas are:
        *   `VolumeInCrypto = ConfiguredVolumeEUR / BuyPrice`
        *   `GrossProfitEUR = (SellPrice - BuyPrice) * VolumeInCrypto`
        *   `BuyLegFee = (BuyPrice * VolumeInCrypto) * BuyExchangeTakerFeePercent`
        *   `SellLegFee = (SellPrice * VolumeInCrypto) * SellExchangeTakerFeePercent`
        *   `TotalFeesEUR = BuyLegFee + SellLegFee + ConfiguredNetworkFeeEUR`
        *   `NetProfitEUR = GrossProfitEUR - TotalFeesEUR`
    *   **Latency Simulation:** When a profitable trade is identified, the engine must call `time.Sleep()` for the `simulated_latency_ms` duration defined in the config. After the delay, it will log the trade to the database using the prices that *initially* identified the opportunity.
    *   **Future Consideration (Slippage):** This version of the engine does not model slippage (price impact from trade volume). The architecture should allow for future replacement of `PriceTick` with order book data to enable more advanced simulation in a V2.

*   **`internal/exchange/*.go` - Exchange Clients**
    *   **Resilience:** All client implementations of the `ExchangeClient` interface **must** be resilient. If a WebSocket connection drops, the client must not exit. It must log the error and automatically attempt to reconnect using an **exponential backoff** strategy (e.g., wait 1s, 2s, 4s, 8s...). The reconnection loop must respect the application's master `context` to allow for graceful shutdown.

*   **`cmd/referee/main.go` - Main Application**
    *   **Data Flow:** The application will use a "fan-in" concurrency pattern. A single `chan model.PriceTick` will be created in `main`. This channel will be passed to every running `ExchangeClient` goroutine. The `ArbitrageEngine` goroutine will be the sole consumer of this channel.

**9. Local Environment & Workflow**

*   **Docker Compose Setup (`docker-compose.yml`)**
    ```yaml
    version: '3.8'

    services:
      postgres:
        image: postgres:16
        container_name: referee_db
        environment:
          POSTGRES_USER: user
          POSTGRES_PASSWORD: password
          POSTGRES_DB: referee_sim
        ports:
          - "5432:5432"
        volumes:
          - postgres_data:/var/lib/postgresql/data
        restart: unless-stopped

      metabase:
        image: metabase/metabase:latest
        container_name: referee_dashboard
        ports:
          - "3000:3000"
        depends_on:
          - postgres
        restart: unless-stopped

    volumes:
      postgres_data:
    ```
*   **Developer Makefile (`Makefile`)**
    ```makefile
    # Makefile for Project Referee

    .PHONY: all build test run clean docker-up docker-down

    # Set Go binary name
    BINARY_NAME=referee

    all: test build

    # Build the Go application
    build:
    	@echo "Building Go binary..."
    	@go build -o ./bin/$(BINARY_NAME) ./cmd/referee

    # Run tests with coverage
    test:
    	@echo "Running tests..."
    	@go test -v -cover ./...

    # Run the application (requires config.yaml)
    run: build
    	@echo "Starting Referee..."
    	@./bin/$(BINARY_NAME)

    # Clean up binaries
    clean:
    	@echo "Cleaning up..."
    	@go clean
    	@rm -f ./bin/$(BINARY_NAME)

    # Start docker-compose services in the background
    docker-up:
    	@echo "Starting Docker services (Postgres & Metabase)..."
    	@docker-compose up -d

    # Stop and remove docker-compose services
    docker-down:
    	@echo "Stopping Docker services..."
    	@docker-compose down
    ```

**10. Containerization & Security (`Dockerfile`)**

A `Dockerfile` must be created to produce a minimal and secure production image.

*   **Multi-Stage Build:** The Dockerfile must use a multi-stage build. The first stage uses a standard Go build image to compile the application. The final stage copies the compiled binary into a minimal "distroless" or Alpine-based image.
*   **Security:** The final image must contain only the compiled binary and any necessary certificates. It must be configured to run as a **non-root user** to reduce potential attack surface.

**11. Step-by-Step Prompts to Guide AI-Assisted Development**

This sequence is the full, actionable build plan.

**Phase 1: Foundation & Structure**
1.  **Project Init:** "Initialize a new Go project named `referee`. Create the standard project layout, the `docker-compose.yml`, the `Makefile`, and the core interface/model files (`client.go`, `repository.go`, `models.go`) with the content from the tech spec."
2.  **Configuration:** "In `internal/config`, use `viper` to load configuration from `config.yaml` into a Go struct. Create the `config.example.yaml` with the clarified structure for `arbitrage`, `database`, and per-exchange `exchanges` settings. **Crucially, configure Viper to automatically bind and override keys with environment variables (e.g., `REFEREE_DB_PASSWORD`).**"

**Phase 2: Database Layer (Test-Driven Development)**
3.  **DB Repository Test:** "In `internal/database/`, write an **integration test** for the `PostgresRepository`. The test should use a library like `testcontainers-go` to programmatically start a PostgreSQL container, ensuring tests are hermetic and don't depend on `docker-compose`."
4.  **DB Repository Implementation:** "In `internal/database/`, create a `PostgresRepository` that implements the `Repository` interface using a `pgxpool` connection pool. The `LogTrade` method must use a `context.WithTimeout` for resilience."

**Phase 3: The Arbitrage Engine (Test-Driven Development)**
5.  **Engine Test:** "In `internal/arbitrage/engine_test.go`, write a table-driven unit test for the `ArbitrageEngine`. Test cases must cover: a) no opportunity, b) a profitable opportunity, c) an opportunity made unprofitable by per-exchange fees and network fees. Use `testify/mock` for the `Repository`."
6.  **Engine Implementation:** "In `internal/arbitrage/engine.go`, create the `ArbitrageEngine`. It must:
    *   Accept a logger, repository, and config struct.
    *   Maintain an in-memory `map[string]model.PriceTick` for state.
    *   When a profitable opportunity is found, `time.Sleep()` for `simulated_latency_ms` *before* logging the trade with the original prices.
    *   Correctly calculate net profit using per-exchange taker fees and the global network withdrawal fee."

**Phase 4: Exchange Clients & Main Application**
7.  **Exchange Client (Kraken):** "In `internal/exchange/`, create a `KrakenClient` that implements the `ExchangeClient` interface for the `BTC/EUR` pair. The `StartStream` method **must feature a resilient WebSocket reconnection loop with exponential backoff** and must respect the parent `context` for graceful shutdown."
8.  **Exchange Client (Binance):** "Following the same resilience pattern, create a `BinanceClient` for the `btceur` stream."
9.  **Tying it all together in `main`:** "In `cmd/referee/main.go`, write the main application logic. It must:
    *   Initialize `slog` and load the configuration.
    *   Create the `pgxpool` connection and instantiate `PostgresRepository`.
    *   Instantiate the `ArbitrageEngine`, injecting the logger, repository, and config.
    *   Create a list of `ExchangeClient`s.
    *   Set up a `context.WithCancel` for graceful shutdown.
    *   Launch all clients and the engine in goroutines, implementing the **fan-in pattern** where all clients write to a single channel that the engine reads from."

**Phase 5: Containerization, Visualization & Expansion**
10. **Dockerfile:** "Create a multi-stage `Dockerfile` in the project root. The first stage should build the Go binary in a standard Go build image. The final stage should copy the compiled binary into a minimal `gcr.io/distroless/static-debian11` or `alpine` image. Ensure the final container runs as a non-root user."
11. **Visualization Guide:** "Generate a step-by-step guide for setting up Metabase. The guide should detail how to:
    *   Access Metabase at `localhost:3000`.
    *   Connect to the Dockerized `referee_db` PostgreSQL container (hint: use the Docker network service name `postgres` as the host).
    *   Create a new 'Question' that queries the `simulated_trades` table.
    *   Build a dashboard that includes: a time-series chart of cumulative net profit, a table of the 20 most recent trades, and a bar chart showing total profit grouped by `buy_exchange`."
12. **Expansion Plan:** "To conclude, add a note to the project's README explaining how to add a new exchange (e.g., Coinbase). The explanation should state that a developer only needs to create a new `coinbase.go` file that implements the `ExchangeClient` interface and add it to the client list in `main.go`."
