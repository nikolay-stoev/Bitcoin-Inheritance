package contract

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

// ContractInfo represents the saved contract information
type ContractInfo struct {
	// Contract metadata
	ContractID   string    `json:"contract_id"`
	CreatedAt    time.Time `json:"created_at"`
	Network      string    `json:"network"`
	TimelockDays int64     `json:"timelock_days"`

	// Keys (WIF format for easy import)
	OwnerWIF     string `json:"owner_wif"`
	InheritorWIF string `json:"inheritor_wif"`

	// Script and address info
	RedeemScript string `json:"redeem_script"` // hex encoded
	P2WSHAddress string `json:"p2wsh_address"`
	ScriptHash   string `json:"script_hash"` // hex encoded

	// Funding status
	IsFunded      bool   `json:"is_funded"`
	FundingTxID   string `json:"funding_tx_id,omitempty"`
	FundingAmount int64  `json:"funding_amount,omitempty"` // satoshis
	FundingVout   uint32 `json:"funding_vout,omitempty"`
}

// SaveContractInfo saves contract information to a JSON file
func SaveContractInfo(contractInfo *ContractInfo) error {
	// Create contracts directory if it doesn't exist
	contractsDir := "contracts"
	if err := os.MkdirAll(contractsDir, 0755); err != nil {
		return fmt.Errorf("failed to create contracts directory: %w", err)
	}

	// Generate filename based on contract ID
	filename := fmt.Sprintf("%s.json", contractInfo.ContractID)
	filepath := filepath.Join(contractsDir, filename)

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(contractInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal contract info: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write contract file: %w", err)
	}

	return nil
}

// LoadContractInfo loads contract information from a JSON file
func LoadContractInfo(contractID string) (*ContractInfo, error) {
	filename := fmt.Sprintf("%s.json", contractID)
	filepath := filepath.Join("contracts", filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read contract file: %w", err)
	}

	var contractInfo ContractInfo
	if err := json.Unmarshal(data, &contractInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal contract info: %w", err)
	}

	return &contractInfo, nil
}

// ListContracts returns a list of all saved contract IDs
func ListContracts() ([]string, error) {
	contractsDir := "contracts"

	// Check if directory exists
	if _, err := os.Stat(contractsDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	files, err := os.ReadDir(contractsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read contracts directory: %w", err)
	}

	var contractIDs []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			contractID := file.Name()[:len(file.Name())-5] // remove .json extension
			contractIDs = append(contractIDs, contractID)
		}
	}

	return contractIDs, nil
}

// GenerateContractID generates a unique contract ID based on the P2WSH address
func GenerateContractID(p2wshAddr btcutil.Address, chainParams *chaincfg.Params) string {
	// Use first 8 characters of the address and network prefix
	addrStr := p2wshAddr.EncodeAddress()
	networkPrefix := "testnet"
	if chainParams.Net == chaincfg.MainNetParams.Net {
		networkPrefix = "mainnet"
	}

	return fmt.Sprintf("%s_%s", networkPrefix, addrStr[len(addrStr)-8:])
}

// UpdateFundingStatus updates the funding status of a contract
func UpdateFundingStatus(contractID, txID string, vout uint32, amount int64) error {
	contractInfo, err := LoadContractInfo(contractID)
	if err != nil {
		return fmt.Errorf("failed to load contract: %w", err)
	}

	contractInfo.IsFunded = true
	contractInfo.FundingTxID = txID
	contractInfo.FundingVout = vout
	contractInfo.FundingAmount = amount

	return SaveContractInfo(contractInfo)
}
