package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/nikolay.stoev/bitcoin-inheritance/config"
	"github.com/nikolay.stoev/bitcoin-inheritance/contract"
	"github.com/nikolay.stoev/bitcoin-inheritance/keys"
	"github.com/nikolay.stoev/bitcoin-inheritance/rpc"
	"github.com/nikolay.stoev/bitcoin-inheritance/script"
	"github.com/nikolay.stoev/bitcoin-inheritance/transaction"
	"github.com/spf13/cobra"
)

var (
	// Global configuration
	cfg *config.Config

	// Command line flags
	testnet      bool
	timelockDays int64
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "bitcoin-inheritance",
	Short: "Bitcoin Inheritance Protocol using Conditional Timelocks",
	Long: `A Bitcoin inheritance application that creates time-locked contracts
using OP_IF/OP_ELSE conditional logic and OP_CHECKSEQUENCEVERIFY timelocks.

The contract allows:
- Owner to spend funds at any time
- Inheritor to spend funds after the timelock expires`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load configuration from environment variables
		cfg = config.LoadConfig()

		// Override network if testnet flag is explicitly set to false
		if !testnet {
			// Force mainnet configuration
			cfg = config.LoadConfig()
			log.Printf("Using mainnet configuration (forced by --testnet=false)")
		} else {
			log.Printf("Using configuration from environment (.env file or system env vars)")
		}

		// Override timelock if specified via command line
		if timelockDays > 0 {
			cfg.Contract.TimelockDays = timelockDays
			log.Printf("Timelock overridden via command line: %d days", timelockDays)
		}

		log.Printf("Network: %s", cfg.ChainParams.Name)
		log.Printf("Timelock duration: %d days", cfg.Contract.TimelockDays)
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new inheritance contract",
	Long: `Generate a new inheritance contract with fresh keys for owner and inheritor.
This creates the redeem script and derives the P2WSH funding address.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return generateContract()
	},
}

var showCmd = &cobra.Command{
	Use:   "show [contract-id]",
	Short: "Show details of a specific inheritance contract",
	Long:  `Show detailed information about a specific inheritance contract by its ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return showContract(args[0])
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved inheritance contracts",
	Long:  `List all inheritance contracts that have been generated and saved locally.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listContracts()
	},
}

var ownerWithdrawCmd = &cobra.Command{
	Use:   "owner-withdraw",
	Short: "Create owner withdrawal transaction",
	Long: `Create and sign a transaction for the owner to withdraw funds immediately.
This uses the IF path of the contract script.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ownerWithdraw()
	},
}

var inheritorWithdrawCmd = &cobra.Command{
	Use:   "inheritor-withdraw",
	Short: "Create inheritor withdrawal transaction",
	Long: `Create and sign a transaction for the inheritor to withdraw funds after timelock.
This uses the ELSE path of the contract script and requires the timelock to have expired.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return inheritorWithdraw()
	},
}

func init() {
	// Add persistent flags
	rootCmd.PersistentFlags().BoolVar(&testnet, "testnet", true, "Use testnet (default: true)")
	rootCmd.PersistentFlags().Int64Var(&timelockDays, "timelock-days", 0, "Timelock duration in days (default: 180)")

	// Add subcommands
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(ownerWithdrawCmd)
	rootCmd.AddCommand(inheritorWithdrawCmd)
}

func generateContract() error {
	log.Printf("=== Generating Bitcoin Inheritance Contract ===")

	// Step 1: Generate keys for owner and inheritor
	log.Printf("Step 1: Generating cryptographic keys...")
	inheritanceKeys, err := keys.GenerateInheritanceKeys(cfg.ChainParams)
	if err != nil {
		return fmt.Errorf("failed to generate keys: %w", err)
	}

	// Step 2: Create the inheritance script
	log.Printf("Step 2: Building inheritance script...")
	ownerPubKey := inheritanceKeys.Owner.GetCompressedPubKeyBytes()
	inheritorPubKey := inheritanceKeys.Inheritor.GetCompressedPubKeyBytes()

	inheritanceScript, err := script.NewInheritanceScript(
		ownerPubKey,
		inheritorPubKey,
		cfg.Contract.TimelockDays,
		cfg.ChainParams,
	)
	if err != nil {
		return fmt.Errorf("failed to create inheritance script: %w", err)
	}

	// Step 3: Validate the script
	log.Printf("Step 3: Validating script...")
	if err := inheritanceScript.ValidateScript(); err != nil {
		return fmt.Errorf("script validation failed: %w", err)
	}

	// Step 4: Generate P2WSH address
	log.Printf("Step 4: Generating P2WSH funding address...")
	p2wshAddr, err := inheritanceScript.GetP2WSHAddress()
	if err != nil {
		return fmt.Errorf("failed to generate P2WSH address: %w", err)
	}

	// Step 5: Save contract details and provide funding instructions
	log.Printf("Step 5: Saving contract details and providing funding instructions...")

	// Generate contract ID
	contractID := contract.GenerateContractID(p2wshAddr, cfg.ChainParams)

	// Create contract info structure
	contractInfo := &contract.ContractInfo{
		ContractID:   contractID,
		CreatedAt:    time.Now(),
		Network:      cfg.ChainParams.Name,
		TimelockDays: cfg.Contract.TimelockDays,
		OwnerWIF:     inheritanceKeys.Owner.WIF.String(),
		InheritorWIF: inheritanceKeys.Inheritor.WIF.String(),
		RedeemScript: fmt.Sprintf("%x", inheritanceScript.RedeemScript),
		P2WSHAddress: p2wshAddr.EncodeAddress(),
		ScriptHash:   fmt.Sprintf("%x", inheritanceScript.GetScriptHash()),
		IsFunded:     false,
	}

	// Save contract to file
	if err := contract.SaveContractInfo(contractInfo); err != nil {
		log.Printf("Warning: Failed to save contract info: %v", err)
	} else {
		log.Printf("Contract details saved to: contracts/%s.json", contractID)
	}

	// Test RPC connection (optional)
	rpcClient := rpc.NewRPCClient(&cfg.RPCConfig)
	if err := rpcClient.TestConnection(); err != nil {
		log.Printf("Warning: RPC connection test failed: %v", err)
		log.Printf("You can still fund the contract manually using the address above")
	} else {
		log.Printf("RPC connection successful - ready for automated operations")
		// TODO: Implement automated funding and transaction broadcasting
	}

	// Provide funding instructions
	log.Printf("\n=== Next Steps ===")
	log.Printf("1. Send Bitcoin to the contract address: %s", p2wshAddr.EncodeAddress())
	log.Printf("2. The contract will be active once funded")
	log.Printf("3. Use 'owner-withdraw' command to spend as owner (immediate)")
	log.Printf("4. Use 'inheritor-withdraw' command to spend as inheritor (after %d days)", cfg.Contract.TimelockDays)
	log.Printf("5. Contract ID for future reference: %s", contractID)

	return nil
}

func showContract(contractID string) error {
	log.Printf("=== Contract Details: %s ===", contractID)

	contractInfo, err := contract.LoadContractInfo(contractID)
	if err != nil {
		return fmt.Errorf("failed to load contract: %w", err)
	}

	log.Printf("Contract ID: %s", contractInfo.ContractID)
	log.Printf("Network: %s", contractInfo.Network)
	log.Printf("Created: %s", contractInfo.CreatedAt.Format("2006-01-02 15:04:05 MST"))
	log.Printf("Timelock: %d days", contractInfo.TimelockDays)
	log.Printf("")
	log.Printf("Funding Address (P2WSH): %s", contractInfo.P2WSHAddress)
	log.Printf("Script Hash: %s", contractInfo.ScriptHash)
	log.Printf("Redeem Script: %s", contractInfo.RedeemScript)
	log.Printf("")
	log.Printf("Owner WIF: %s", contractInfo.OwnerWIF)
	log.Printf("Inheritor WIF: %s", contractInfo.InheritorWIF)
	log.Printf("")
	log.Printf("Funding Status: %t", contractInfo.IsFunded)
	if contractInfo.IsFunded {
		log.Printf("Funding Transaction: %s:%d", contractInfo.FundingTxID, contractInfo.FundingVout)
		log.Printf("Funding Amount: %d satoshis", contractInfo.FundingAmount)
	} else {
		log.Printf("To fund this contract, send Bitcoin to: %s", contractInfo.P2WSHAddress)
	}

	return nil
}

func listContracts() error {
	log.Printf("=== Saved Inheritance Contracts ===")

	contractIDs, err := contract.ListContracts()
	if err != nil {
		return fmt.Errorf("failed to list contracts: %w", err)
	}

	if len(contractIDs) == 0 {
		log.Printf("No contracts found. Use 'generate' command to create a new contract.")
		return nil
	}

	for i, contractID := range contractIDs {
		contractInfo, err := contract.LoadContractInfo(contractID)
		if err != nil {
			log.Printf("%d. %s (error loading: %v)", i+1, contractID, err)
			continue
		}

		log.Printf("%d. Contract ID: %s", i+1, contractInfo.ContractID)
		log.Printf("   Network: %s", contractInfo.Network)
		log.Printf("   Created: %s", contractInfo.CreatedAt.Format("2006-01-02 15:04:05"))
		log.Printf("   Timelock: %d days", contractInfo.TimelockDays)
		log.Printf("   Address: %s", contractInfo.P2WSHAddress)
		log.Printf("   Funded: %t", contractInfo.IsFunded)
		if contractInfo.IsFunded {
			log.Printf("   Funding: %d satoshis (txid: %s:%d)",
				contractInfo.FundingAmount, contractInfo.FundingTxID, contractInfo.FundingVout)
		}
		log.Printf("")
	}

	return nil
}

func ownerWithdraw() error {
	log.Printf("=== Owner Withdrawal ===")

	// Step 1: Get contract ID from user
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter contract ID: ")
	contractID, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read contract ID: %w", err)
	}
	contractID = strings.TrimSpace(contractID)

	// Step 2: Load contract details and UTXO information
	log.Printf("Step 1: Loading contract details...")
	contractInfo, err := contract.LoadContractInfo(contractID)
	if err != nil {
		return fmt.Errorf("failed to load contract: %w", err)
	}

	if !contractInfo.IsFunded {
		return fmt.Errorf("contract is not funded yet")
	}

	log.Printf("Contract found: %s", contractInfo.P2WSHAddress)
	log.Printf("Funding UTXO: %s:%d (%d satoshis)",
		contractInfo.FundingTxID, contractInfo.FundingVout, contractInfo.FundingAmount)

	// Step 3: Load owner's private key from WIF
	log.Printf("Step 2: Loading owner's private key...")
	ownerKeys, err := keys.KeyPairFromWIF(contractInfo.OwnerWIF, cfg.ChainParams)
	if err != nil {
		return fmt.Errorf("failed to load owner keys: %w", err)
	}

	// Step 4: Get owner's destination address
	fmt.Print("Enter destination address for withdrawal: ")
	destAddrStr, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read destination address: %w", err)
	}
	destAddrStr = strings.TrimSpace(destAddrStr)

	destAddr, err := btcutil.DecodeAddress(destAddrStr, cfg.ChainParams)
	if err != nil {
		return fmt.Errorf("invalid destination address: %w", err)
	}

	// Step 5: Parse funding transaction hash
	fundingHash, err := chainhash.NewHashFromStr(contractInfo.FundingTxID)
	if err != nil {
		return fmt.Errorf("invalid funding transaction hash: %w", err)
	}

	// Step 6: Parse redeem script
	redeemScript, err := hex.DecodeString(contractInfo.RedeemScript)
	if err != nil {
		return fmt.Errorf("failed to decode redeem script: %w", err)
	}

	// Step 7: Create UTXO for the contract
	contractUTXO := &transaction.UTXO{
		TxHash:   fundingHash,
		Vout:     contractInfo.FundingVout,
		Amount:   btcutil.Amount(contractInfo.FundingAmount),
		PkScript: nil, // Will be filled by the signing process
	}

	// Step 8: Build transaction using the IF path
	log.Printf("Step 3: Building withdrawal transaction...")

	// Set a reasonable fee (500 satoshis)
	fee := btcutil.Amount(500)

	txBuilder := transaction.NewTransactionBuilder(cfg.ChainParams, fee)
	tx, err := txBuilder.BuildOwnerWithdrawTx(contractUTXO, destAddr, redeemScript)
	if err != nil {
		return fmt.Errorf("failed to build transaction: %w", err)
	}

	// Step 9: Sign with owner's key and OP_1 selector
	log.Printf("Step 4: Signing transaction...")
	if err := txBuilder.SignOwnerTransaction(tx, contractUTXO, redeemScript, ownerKeys.PrivateKey); err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Step 10: Validate transaction
	if err := txBuilder.ValidateTransaction(tx); err != nil {
		return fmt.Errorf("transaction validation failed: %w", err)
	}

	// Step 11: Serialize transaction for broadcasting
	txHex, err := txBuilder.SerializeTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to serialize transaction: %w", err)
	}

	log.Printf("Transaction built successfully!")
	log.Printf("Transaction hex: %s", txHex)

	// Step 12: Ask user for confirmation before broadcasting
	fmt.Print("Do you want to broadcast this transaction? (y/N): ")
	confirm, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		log.Printf("Transaction not broadcast (user cancelled)")
		return nil
	}

	// Step 13: Broadcast transaction
	log.Printf("Step 5: Broadcasting transaction...")
	rpcClient := rpc.NewRPCClient(&cfg.RPCConfig)

	txid, err := rpcClient.BroadcastTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	log.Printf("✅ Transaction broadcast successfully!")
	log.Printf("Transaction ID: %s", txid)
	log.Printf("Owner withdrawal completed!")

	return nil
}

func inheritorWithdraw() error {
	log.Printf("=== Inheritor Withdrawal ===")

	// Step 1: Get contract ID from user
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter contract ID: ")
	contractID, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read contract ID: %w", err)
	}
	contractID = strings.TrimSpace(contractID)

	// Step 2: Load contract details and UTXO information
	log.Printf("Step 1: Loading contract details...")
	contractInfo, err := contract.LoadContractInfo(contractID)
	if err != nil {
		return fmt.Errorf("failed to load contract: %w", err)
	}

	if !contractInfo.IsFunded {
		return fmt.Errorf("contract is not funded yet")
	}

	log.Printf("Contract found: %s", contractInfo.P2WSHAddress)
	log.Printf("Funding UTXO: %s:%d (%d satoshis)",
		contractInfo.FundingTxID, contractInfo.FundingVout, contractInfo.FundingAmount)

	// Step 3: Verify timelock has expired
	// Calculate the required timelock in blocks (assuming 10 minutes per block)
	relativeTimelock := contractInfo.TimelockDays * 24 * 6 // days * hours * blocks per hour
	log.Printf("Step 2: Verifying timelock has expired...")
	log.Printf("Required timelock: %d blocks (%d days)", relativeTimelock, contractInfo.TimelockDays)
	log.Printf("Note: This implementation requires manual verification that enough blocks have passed")

	// Step 4: Load inheritor's private key from WIF
	log.Printf("Step 3: Loading inheritor's private key...")
	inheritorKeys, err := keys.KeyPairFromWIF(contractInfo.InheritorWIF, cfg.ChainParams)
	if err != nil {
		return fmt.Errorf("failed to load inheritor keys: %w", err)
	}

	// Step 5: Get inheritor's destination address
	fmt.Print("Enter destination address for withdrawal: ")
	destAddrStr, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read destination address: %w", err)
	}
	destAddrStr = strings.TrimSpace(destAddrStr)

	destAddr, err := btcutil.DecodeAddress(destAddrStr, cfg.ChainParams)
	if err != nil {
		return fmt.Errorf("invalid destination address: %w", err)
	}

	// Step 6: Parse funding transaction hash
	fundingHash, err := chainhash.NewHashFromStr(contractInfo.FundingTxID)
	if err != nil {
		return fmt.Errorf("invalid funding transaction hash: %w", err)
	}

	// Step 7: Parse redeem script
	redeemScript, err := hex.DecodeString(contractInfo.RedeemScript)
	if err != nil {
		return fmt.Errorf("failed to decode redeem script: %w", err)
	}

	// Step 8: Create UTXO for the contract
	contractUTXO := &transaction.UTXO{
		TxHash:   fundingHash,
		Vout:     contractInfo.FundingVout,
		Amount:   btcutil.Amount(contractInfo.FundingAmount),
		PkScript: nil, // Will be filled by the signing process
	}

	// Step 9: Build transaction using the ELSE path with correct nSequence
	log.Printf("Step 4: Building withdrawal transaction...")

	// Set a reasonable fee (500 satoshis)
	fee := btcutil.Amount(500)

	txBuilder := transaction.NewTransactionBuilder(cfg.ChainParams, fee)
	tx, err := txBuilder.BuildInheritorWithdrawTx(contractUTXO, destAddr, redeemScript, relativeTimelock)
	if err != nil {
		return fmt.Errorf("failed to build transaction: %w", err)
	}

	// Step 10: Sign with inheritor's key and OP_0 selector
	log.Printf("Step 5: Signing transaction...")
	if err := txBuilder.SignInheritorTransaction(tx, contractUTXO, redeemScript, inheritorKeys.PrivateKey); err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Step 11: Validate transaction
	if err := txBuilder.ValidateTransaction(tx); err != nil {
		return fmt.Errorf("transaction validation failed: %w", err)
	}

	// Step 12: Serialize transaction for broadcasting
	txHex, err := txBuilder.SerializeTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to serialize transaction: %w", err)
	}

	log.Printf("Transaction built successfully!")
	log.Printf("Transaction hex: %s", txHex)

	// Step 13: Ask user for confirmation before broadcasting
	fmt.Print("Do you want to broadcast this transaction? (y/N): ")
	confirm, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		log.Printf("Transaction not broadcast (user cancelled)")
		return nil
	}

	// Step 14: Broadcast transaction
	log.Printf("Step 6: Broadcasting transaction...")
	rpcClient := rpc.NewRPCClient(&cfg.RPCConfig)

	txid, err := rpcClient.BroadcastTransaction(tx)
	if err != nil {
		return fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	log.Printf("✅ Transaction broadcast successfully!")
	log.Printf("Transaction ID: %s", txid)
	log.Printf("Inheritor withdrawal completed!")

	return nil
}
