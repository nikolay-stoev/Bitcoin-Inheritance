package config

import (
	"github.com/btcsuite/btcd/chaincfg"
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

// NewTestnetConfig creates a new configuration for testnet
func NewTestnetConfig() *Config {
	return &Config{
		ChainParams: &chaincfg.TestNet3Params,
		RPCConfig: RPCConfig{
			Host:         "localhost:18334", // Default btcd testnet RPC port
			User:         "testuser",        // Sample RPC username
			Pass:         "testpass",        // Sample RPC password
			HTTPPostMode: true,
			DisableTLS:   false,
		},
		Contract: ContractConfig{
			TimelockDays: 180,  // 180 days default timelock
			DefaultFee:   2000, // 2000 satoshis default fee
		},
	}
}

// NewMainnetConfig creates a new configuration for mainnet
func NewMainnetConfig() *Config {
	return &Config{
		ChainParams: &chaincfg.MainNetParams,
		RPCConfig: RPCConfig{
			Host:         "localhost:8334", // Default btcd mainnet RPC port
			User:         "mainuser",       // Sample RPC username
			Pass:         "mainpass",       // Sample RPC password
			HTTPPostMode: true,
			DisableTLS:   false,
		},
		Contract: ContractConfig{
			TimelockDays: 180,  // 180 days default timelock
			DefaultFee:   2000, // 2000 satoshis default fee
		},
	}
}
