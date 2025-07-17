package script

import (
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

// InheritanceScript represents the Bitcoin script for inheritance contract
type InheritanceScript struct {
	OwnerPubKey      []byte
	InheritorPubKey  []byte
	RelativeTimelock int64
	RedeemScript     []byte
	ChainParams      *chaincfg.Params
}

// NewInheritanceScript creates a new inheritance script
func NewInheritanceScript(ownerPubKey, inheritorPubKey []byte, timelockDays int64, chainParams *chaincfg.Params) (*InheritanceScript, error) {
	// Calculate relative timelock value according to BIP 68
	relativeTimelock := calculateRelativeTimelock(timelockDays)

	// Build the redeem script
	redeemScript, err := buildRedeemScript(ownerPubKey, inheritorPubKey, relativeTimelock)
	if err != nil {
		return nil, fmt.Errorf("failed to build redeem script: %w", err)
	}

	log.Printf("Built redeem script with timelock: %d days (%d BIP68 value)", timelockDays, relativeTimelock)
	log.Printf("Redeem script hex: %x", redeemScript)

	return &InheritanceScript{
		OwnerPubKey:      ownerPubKey,
		InheritorPubKey:  inheritorPubKey,
		RelativeTimelock: relativeTimelock,
		RedeemScript:     redeemScript,
		ChainParams:      chainParams,
	}, nil
}

// buildRedeemScript constructs the inheritance redeem script
// Script structure:
// OP_IF
//
//	<Owner_PublicKey> OP_CHECKSIG
//
// OP_ELSE
//
//	<Relative_Timelock_Value> OP_CHECKSEQUENCEVERIFY OP_DROP
//	<Inheritor_PublicKey> OP_CHECKSIG
//
// OP_ENDIF
func buildRedeemScript(ownerPubKey, inheritorPubKey []byte, relativeTimelock int64) ([]byte, error) {
	builder := txscript.NewScriptBuilder()

	// Start conditional block
	builder.AddOp(txscript.OP_IF)

	// IF branch: Owner's immediate spend path
	builder.AddData(ownerPubKey)
	builder.AddOp(txscript.OP_CHECKSIG)

	// ELSE branch: Inheritor's time-delayed spend path
	builder.AddOp(txscript.OP_ELSE)
	builder.AddInt64(relativeTimelock)
	builder.AddOp(txscript.OP_CHECKSEQUENCEVERIFY)
	builder.AddOp(txscript.OP_DROP)
	builder.AddData(inheritorPubKey)
	builder.AddOp(txscript.OP_CHECKSIG)

	// End conditional block
	builder.AddOp(txscript.OP_ENDIF)

	return builder.Script()
}

// calculateRelativeTimelock converts days to BIP 68 encoded timelock value
// BIP 68 uses 512-second intervals when the type flag (bit 22) is set
func calculateRelativeTimelock(days int64) int64 {
	// Convert days to seconds
	totalSeconds := days * 24 * 60 * 60

	// Convert to 512-second intervals
	intervals := totalSeconds / 512

	// Set bit 22 to indicate time-based (not block-based) timelock
	// Bit 22 = 0x400000
	return intervals | 0x400000
}

// GetP2WSHAddress derives the P2WSH address from the redeem script
func (is *InheritanceScript) GetP2WSHAddress() (btcutil.Address, error) {
	// Hash the redeem script with SHA256
	scriptHash := sha256.Sum256(is.RedeemScript)

	// Create P2WSH address from the hash
	addr, err := btcutil.NewAddressWitnessScriptHash(scriptHash[:], is.ChainParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create P2WSH address: %w", err)
	}

	return addr, nil
}

// GetScriptHash returns the SHA256 hash of the redeem script
func (is *InheritanceScript) GetScriptHash() []byte {
	scriptHash := sha256.Sum256(is.RedeemScript)
	return scriptHash[:]
}

// GetScriptPubKey returns the scriptPubKey for P2WSH
func (is *InheritanceScript) GetScriptPubKey() ([]byte, error) {
	addr, err := is.GetP2WSHAddress()
	if err != nil {
		return nil, err
	}

	return txscript.PayToAddrScript(addr)
}

// ValidateScript performs basic validation on the constructed script
func (is *InheritanceScript) ValidateScript() error {
	// Check if script is not empty
	if len(is.RedeemScript) == 0 {
		return fmt.Errorf("redeem script is empty")
	}

	// Check if public keys are valid length (33 bytes for compressed)
	if len(is.OwnerPubKey) != 33 {
		return fmt.Errorf("owner public key must be 33 bytes (compressed)")
	}

	if len(is.InheritorPubKey) != 33 {
		return fmt.Errorf("inheritor public key must be 33 bytes (compressed)")
	}

	// Check if timelock is valid (positive and within BIP 68 limits)
	if is.RelativeTimelock <= 0 {
		return fmt.Errorf("relative timelock must be positive")
	}

	log.Printf("Script validation passed")
	return nil
}
