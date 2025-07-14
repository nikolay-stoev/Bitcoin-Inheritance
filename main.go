package main

import (
	"fmt"
	"log"

	"github.com/nikolay.stoev/bitcoin-inheritance/config"
	"github.com/nikolay.stoev/bitcoin-inheritance/keys"
	"github.com/nikolay.stoev/bitcoin-inheritance/script"
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
		// Initialize configuration based on network
		if testnet {
			cfg = config.NewTestnetConfig()
			log.Printf("Using testnet configuration")
		} else {
			cfg = config.NewMainnetConfig()
			log.Printf("Using mainnet configuration")
		}

		// Override timelock if specified
		if timelockDays > 0 {
			cfg.Contract.TimelockDays = timelockDays
		}

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

	// Display results
	log.Printf("\n=== Contract Generated Successfully ===")
	log.Printf("Funding Address (P2WSH): %s", p2wshAddr.EncodeAddress())
	log.Printf("Owner WIF: %s", inheritanceKeys.Owner.WIF.String())
	log.Printf("Inheritor WIF: %s", inheritanceKeys.Inheritor.WIF.String())
	log.Printf("Timelock: %d days", cfg.Contract.TimelockDays)

	log.Printf("\n=== Next Steps ===")
	log.Printf("1. Fund the contract address with testnet Bitcoin")
	log.Printf("2. Use a testnet faucet: https://testnet-faucet.mempool.co/")
	log.Printf("3. Save the transaction ID for spending later")

	return nil
}

func ownerWithdraw() error {
	log.Printf("=== Owner Withdrawal (Placeholder) ===")
	log.Printf("This would implement the owner's immediate withdrawal logic")
	log.Printf("- Load owner's private key from WIF")
	log.Printf("- Load contract details and UTXO information")
	log.Printf("- Build transaction using the IF path")
	log.Printf("- Sign with owner's key and OP_1 selector")
	log.Printf("- Broadcast transaction")

	return nil
}

func inheritorWithdraw() error {
	log.Printf("=== Inheritor Withdrawal (Placeholder) ===")
	log.Printf("This would implement the inheritor's time-delayed withdrawal logic")
	log.Printf("- Verify timelock has expired")
	log.Printf("- Load inheritor's private key from WIF")
	log.Printf("- Load contract details and UTXO information")
	log.Printf("- Build transaction using the ELSE path with correct nSequence")
	log.Printf("- Sign with inheritor's key and OP_0 selector")
	log.Printf("- Broadcast transaction")

	return nil
}
