package config

import (
	"github.com/spf13/viper"
	"strings"
)

// Config stores all configuration for the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	Arbitrage ArbitrageConfig
	Database  DatabaseConfig
	Exchanges map[string]ExchangeConfig
}

// ArbitrageConfig defines the arbitrage-related settings.
type ArbitrageConfig struct {
	SimulatedTradeVolumeEUR float64 `mapstructure:"simulated_trade_volume_eur"`
	NetworkWithdrawalFeeEUR float64 `mapstructure:"network_withdrawal_fee_eur"`
	SimulatedLatencyMS      int     `mapstructure:"simulated_latency_ms"`
}

// DatabaseConfig defines the database connection settings.
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// ExchangeConfig defines settings for a specific exchange.
type ExchangeConfig struct {
	TakerFeePercent float64 `mapstructure:"taker_fee_percent"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
