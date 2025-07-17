package config

import (
	"log"
	"os"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	// Bitcoin network parameters (testnet/mainnet)
	ChainParams *chaincfg.Params

	// RPC client configuration for btcd
	RPCConfig RPCConfig

	// Contract settings
	Contract ContractConfig
}

// RPCConfig holds RPC connection settings
type RPCConfig struct {
	Host         string
	User         string
	Pass         string
	HTTPPostMode bool
	DisableTLS   bool
}

// ContractConfig holds inheritance contract specific settings
type ContractConfig struct {
	// Timelock duration in days
	TimelockDays int64

	// Default transaction fee in satoshis
	DefaultFee int64
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file - exit if not found
	if err := godotenv.Load(); err != nil {
		log.Fatalf(".env file not found: %v", err)
	}

	network := getEnvString("BITCOIN_NETWORK", "testnet")

	var cfg *Config
	if network == "mainnet" {
		cfg = createMainnetConfig()
	} else {
		cfg = createTestnetConfig()
	}

	// Override with environment variables if present
	if timelockDays := getEnvInt64("TIMELOCK_DAYS", cfg.Contract.TimelockDays); timelockDays > 0 {
		cfg.Contract.TimelockDays = timelockDays
	}
	if defaultFee := getEnvInt64("DEFAULT_FEE_SATOSHIS", cfg.Contract.DefaultFee); defaultFee > 0 {
		cfg.Contract.DefaultFee = defaultFee
	}

	return cfg
}

// createTestnetConfig creates a testnet configuration from environment variables
func createTestnetConfig() *Config {
	return &Config{
		ChainParams: &chaincfg.TestNet3Params,
		RPCConfig: RPCConfig{
			Host:         getRequiredEnvString("TESTNET_RPC_HOST"),
			User:         getRequiredEnvString("TESTNET_RPC_USER"),
			Pass:         getRequiredEnvString("TESTNET_RPC_PASS"),
			HTTPPostMode: getEnvBool("TESTNET_RPC_HTTP_POST_MODE", true),
			DisableTLS:   getEnvBool("TESTNET_RPC_DISABLE_TLS", false),
		},
		Contract: ContractConfig{
			TimelockDays: getEnvInt64("TIMELOCK_DAYS", 180),
			DefaultFee:   getEnvInt64("DEFAULT_FEE_SATOSHIS", 2000),
		},
	}
}

// createMainnetConfig creates a mainnet configuration from environment variables
func createMainnetConfig() *Config {
	return &Config{
		ChainParams: &chaincfg.MainNetParams,
		RPCConfig: RPCConfig{
			Host:         getRequiredEnvString("MAINNET_RPC_HOST"),
			User:         getRequiredEnvString("MAINNET_RPC_USER"),
			Pass:         getRequiredEnvString("MAINNET_RPC_PASS"),
			HTTPPostMode: getEnvBool("MAINNET_RPC_HTTP_POST_MODE", true),
			DisableTLS:   getEnvBool("MAINNET_RPC_DISABLE_TLS", false),
		},
		Contract: ContractConfig{
			TimelockDays: getEnvInt64("TIMELOCK_DAYS", 180),
			DefaultFee:   getEnvInt64("DEFAULT_FEE_SATOSHIS", 2000),
		},
	}
}

// Helper functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getRequiredEnvString(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
		log.Printf("Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		log.Printf("Invalid boolean value for %s: %s, using default: %t", key, value, defaultValue)
	}
	return defaultValue
}
