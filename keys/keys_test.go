package keys

import (
	"bytes"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
)

func TestGenerateInheritanceKeys(t *testing.T) {
	// Test key generation for testnet
	keys, err := GenerateInheritanceKeys(&chaincfg.TestNet3Params)
	if err != nil {
		t.Fatalf("Failed to generate inheritance keys: %v", err)
	}

	// Verify owner keys
	if keys.Owner == nil {
		t.Error("Owner keys are nil")
	}
	if keys.Owner.PrivateKey == nil {
		t.Error("Owner private key is nil")
	}
	if keys.Owner.PublicKey == nil {
		t.Error("Owner public key is nil")
	}
	if keys.Owner.WIF == nil {
		t.Error("Owner WIF is nil")
	}

	// Verify inheritor keys
	if keys.Inheritor == nil {
		t.Error("Inheritor keys are nil")
	}
	if keys.Inheritor.PrivateKey == nil {
		t.Error("Inheritor private key is nil")
	}
	if keys.Inheritor.PublicKey == nil {
		t.Error("Inheritor public key is nil")
	}
	if keys.Inheritor.WIF == nil {
		t.Error("Inheritor WIF is nil")
	}

	// Test compressed public key length
	ownerPubKeyBytes := keys.Owner.GetCompressedPubKeyBytes()
	if len(ownerPubKeyBytes) != 33 {
		t.Errorf("Expected owner public key length 33, got %d", len(ownerPubKeyBytes))
	}

	inheritorPubKeyBytes := keys.Inheritor.GetCompressedPubKeyBytes()
	if len(inheritorPubKeyBytes) != 33 {
		t.Errorf("Expected inheritor public key length 33, got %d", len(inheritorPubKeyBytes))
	}

	// Test P2WPKH address generation
	ownerAddr, err := keys.Owner.GetP2WPKHAddress()
	if err != nil {
		t.Fatalf("Failed to generate owner P2WPKH address: %v", err)
	}
	if ownerAddr == nil {
		t.Error("Owner P2WPKH address is nil")
	}

	inheritorAddr, err := keys.Inheritor.GetP2WPKHAddress()
	if err != nil {
		t.Fatalf("Failed to generate inheritor P2WPKH address: %v", err)
	}
	if inheritorAddr == nil {
		t.Error("Inheritor P2WPKH address is nil")
	}

	// Verify that addresses are different
	if ownerAddr.EncodeAddress() == inheritorAddr.EncodeAddress() {
		t.Error("Owner and inheritor addresses should be different")
	}
}

func TestNewKeyPair(t *testing.T) {
	// Test individual key pair generation
	keyPair, err := NewKeyPair(&chaincfg.TestNet3Params)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	if keyPair.PrivateKey == nil {
		t.Error("Private key is nil")
	}
	if keyPair.PublicKey == nil {
		t.Error("Public key is nil")
	}
	if keyPair.WIF == nil {
		t.Error("WIF is nil")
	}
	if keyPair.ChainParams == nil {
		t.Error("ChainParams is nil")
	}

	// Test WIF string format
	wifStr := keyPair.WIF.String()
	if wifStr == "" {
		t.Error("WIF string is empty")
	}

	// Test WIF roundtrip
	restoredKeyPair, err := KeyPairFromWIF(wifStr, &chaincfg.TestNet3Params)
	if err != nil {
		t.Fatalf("Failed to restore key pair from WIF: %v", err)
	}

	// Verify keys match by comparing serialized forms
	if !bytes.Equal(keyPair.PrivateKey.Serialize(), restoredKeyPair.PrivateKey.Serialize()) {
		t.Error("Private keys don't match after WIF roundtrip")
	}
}
