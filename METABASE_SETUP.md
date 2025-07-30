# Metabase Setup Guide for Referee

This guide will help you set up Metabase to visualize the arbitrage simulation data from the Referee project.

## Prerequisites

- Docker and Docker Compose installed
- Referee project running with database populated

## Step 1: Start the Services

First, ensure the Docker services are running:

```bash
make docker-up
```

This will start:
- PostgreSQL database on port 5432
- Metabase on port 3000

## Step 2: Access Metabase

1. Open your web browser and navigate to: `http://localhost:3000`
2. Wait for Metabase to finish initializing (this may take a few minutes)
3. You'll see the Metabase setup screen

## Step 3: Complete Metabase Setup

1. **Create Admin Account**:
   - Enter your email and password
   - Click "Let's get started"

2. **Add Database Connection**:
   - Click "Add your data"
   - Select "PostgreSQL"
   - Use these connection details:
     - **Host**: `postgres` (Docker service name)
     - **Port**: `5432`
     - **Database name**: `referee_sim`
     - **Username**: `user`
     - **Password**: `password`
   - Click "Test connection" to verify
   - Click "Save"

## Step 4: Create Your First Question

1. Click "Ask a question"
2. Select "Raw data"
3. Choose the `referee_sim` database
4. Select the `simulated_trades` table
5. Click "Done"

## Step 5: Build the Dashboard

### 5.1 Cumulative Net Profit Chart

1. Create a new question:
   - **Data source**: `simulated_trades`
   - **Aggregation**: Sum of `net_profit_eur`
   - **Group by**: `timestamp` (by day)
   - **Visualization**: Line chart
   - **Title**: "Cumulative Net Profit Over Time"

### 5.2 Recent Trades Table

1. Create a new question:
   - **Data source**: `simulated_trades`
   - **Limit**: 20 rows
   - **Sort by**: `timestamp` (descending)
   - **Visualization**: Table
   - **Title**: "20 Most Recent Trades"

### 5.3 Profit by Exchange

1. Create a new question:
   - **Data source**: `simulated_trades`
   - **Aggregation**: Sum of `net_profit_eur`
   - **Group by**: `buy_exchange`
   - **Visualization**: Bar chart
   - **Title**: "Total Profit by Buy Exchange"

## Step 6: Create Dashboard

1. Click "Save" on each question
2. Click "Add to dashboard" → "Create new dashboard"
3. Name it "Referee Arbitrage Analysis"
4. Arrange the visualizations as desired

## Step 7: Set Up Auto-refresh

1. In the dashboard, click the refresh icon
2. Set auto-refresh to every 30 seconds or 1 minute
3. This will keep the data current as new trades are logged

## Troubleshooting

### Connection Issues
- Ensure Docker services are running: `docker-compose ps`
- Check if PostgreSQL is accessible: `docker exec referee_db psql -U user -d referee_sim -c "SELECT 1;"`

### No Data Visible
- Verify the Referee application is running and logging trades
- Check the database directly: `docker exec referee_db psql -U user -d referee_sim -c "SELECT COUNT(*) FROM simulated_trades;"`

### Metabase Won't Start
- Check Docker logs: `docker-compose logs metabase`
- Ensure port 3000 is not in use by another application

## Advanced Queries

### Win/Loss Ratio
```sql
SELECT 
  COUNT(CASE WHEN net_profit_eur > 0 THEN 1 END) as winning_trades,
  COUNT(CASE WHEN net_profit_eur <= 0 THEN 1 END) as losing_trades,
  ROUND(COUNT(CASE WHEN net_profit_eur > 0 THEN 1 END) * 100.0 / COUNT(*), 2) as win_percentage
FROM simulated_trades;
```

### Average Profit by Hour
```sql
SELECT 
  EXTRACT(HOUR FROM timestamp) as hour,
  AVG(net_profit_eur) as avg_profit,
  COUNT(*) as trade_count
FROM simulated_trades 
GROUP BY EXTRACT(HOUR FROM timestamp)
ORDER BY hour;
```

### Best Trading Pairs
```sql
SELECT 
  buy_exchange,
  sell_exchange,
  COUNT(*) as trade_count,
  SUM(net_profit_eur) as total_profit,
  AVG(net_profit_eur) as avg_profit
FROM simulated_trades 
GROUP BY buy_exchange, sell_exchange
ORDER BY total_profit DESC;
``` 