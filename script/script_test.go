package script

import (
	"bytes"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
)

// Test helper to create valid compressed public keys
func createTestPubKeys() ([]byte, []byte) {
	// Valid compressed public key examples (33 bytes each)
	ownerPubKey := []byte{
		0x03, 0x2e, 0x58, 0xd0, 0x8c, 0xa4, 0x5c, 0x7d, 0xa8, 0x7b, 0x2f, 0xc6, 0x9c, 0x5b, 0x8a, 0x5e,
		0x1a, 0x3b, 0x4c, 0x5d, 0x6e, 0x7f, 0x8a, 0x9b, 0x0c, 0x1d, 0x2e, 0x3f, 0x4a, 0x5b, 0x6c, 0x7d, 0x8e,
	}
	inheritorPubKey := []byte{
		0x02, 0x4a, 0x5b, 0x6c, 0x7d, 0x8e, 0x9f, 0xa1, 0xb2, 0xc3, 0xd4, 0xe5, 0xf6, 0x07, 0x18, 0x29,
		0x3a, 0x4b, 0x5c, 0x6d, 0x7e, 0x8f, 0x90, 0xa1, 0xb2, 0xc3, 0xd4, 0xe5, 0xf6, 0x07, 0x18, 0x29, 0x3a,
	}
	return ownerPubKey, inheritorPubKey
}

func TestNewInheritanceScript_ValidInput(t *testing.T) {
	ownerPubKey, inheritorPubKey := createTestPubKeys()
	timelockDays := int64(365) // 1 year
	chainParams := &chaincfg.TestNet3Params

	script, err := NewInheritanceScript(ownerPubKey, inheritorPubKey, timelockDays, chainParams)
	if err != nil {
		t.Fatalf("NewInheritanceScript failed: %v", err)
	}

	// Verify script structure
	if script == nil {
		t.Fatal("Script is nil")
	}

	// Verify public keys are stored correctly
	if !bytes.Equal(script.OwnerPubKey, ownerPubKey) {
		t.Error("Owner public key not stored correctly")
	}

	if !bytes.Equal(script.InheritorPubKey, inheritorPubKey) {
		t.Error("Inheritor public key not stored correctly")
	}

	// Verify chain parameters
	if script.ChainParams != chainParams {
		t.Error("Chain parameters not stored correctly")
	}

	// Verify redeem script is generated
	if len(script.RedeemScript) == 0 {
		t.Error("Redeem script is empty")
	}

	// Verify relative timelock is calculated correctly
	expectedTimelock := calculateRelativeTimelock(timelockDays)
	if script.RelativeTimelock != expectedTimelock {
		t.Errorf("Expected relative timelock %d, got %d", expectedTimelock, script.RelativeTimelock)
	}

	// Verify script validation passes
	if err := script.ValidateScript(); err != nil {
		t.Errorf("Script validation failed: %v", err)
	}
}

func TestNewInheritanceScript_DifferentTimelocks(t *testing.T) {
	ownerPubKey, inheritorPubKey := createTestPubKeys()
	chainParams := &chaincfg.TestNet3Params

	testCases := []struct {
		name         string
		timelockDays int64
		expectError  bool
	}{
		{"1 day", 1, false},
		{"30 days", 30, false},
		{"365 days", 365, false},
		{"1000 days", 1000, false},
		{"Zero days", 0, false}, // Should work but result in zero timelock
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			script, err := NewInheritanceScript(ownerPubKey, inheritorPubKey, tc.timelockDays, chainParams)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if script == nil {
				t.Fatal("Script is nil")
			}

			// Verify timelock calculation
			expectedTimelock := calculateRelativeTimelock(tc.timelockDays)
			if script.RelativeTimelock != expectedTimelock {
				t.Errorf("Expected relative timelock %d, got %d", expectedTimelock, script.RelativeTimelock)
			}

			// Verify redeem script is generated
			if len(script.RedeemScript) == 0 {
				t.Error("Redeem script is empty")
			}
		})
	}
}

func TestNewInheritanceScript_InvalidPublicKeys(t *testing.T) {
	chainParams := &chaincfg.TestNet3Params
	timelockDays := int64(365)

	testCases := []struct {
		name            string
		ownerPubKey     []byte
		inheritorPubKey []byte
		expectError     bool
	}{
		{
			name:            "Nil owner public key",
			ownerPubKey:     nil,
			inheritorPubKey: make([]byte, 33),
			expectError:     true,
		},
		{
			name:            "Nil inheritor public key",
			ownerPubKey:     make([]byte, 33),
			inheritorPubKey: nil,
			expectError:     true,
		},
		{
			name:            "Empty owner public key",
			ownerPubKey:     []byte{},
			inheritorPubKey: make([]byte, 33),
			expectError:     true,
		},
		{
			name:            "Empty inheritor public key",
			ownerPubKey:     make([]byte, 33),
			inheritorPubKey: []byte{},
			expectError:     true,
		},
		{
			name:            "Short owner public key",
			ownerPubKey:     make([]byte, 32),
			inheritorPubKey: make([]byte, 33),
			expectError:     true,
		},
		{
			name:            "Short inheritor public key",
			ownerPubKey:     make([]byte, 33),
			inheritorPubKey: make([]byte, 32),
			expectError:     true,
		},
		{
			name:            "Long owner public key",
			ownerPubKey:     make([]byte, 34),
			inheritorPubKey: make([]byte, 33),
			expectError:     true,
		},
		{
			name:            "Long inheritor public key",
			ownerPubKey:     make([]byte, 33),
			inheritorPubKey: make([]byte, 34),
			expectError:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			script, err := NewInheritanceScript(tc.ownerPubKey, tc.inheritorPubKey, timelockDays, chainParams)

			if tc.expectError {
				// The error might come from script validation rather than construction
				if err == nil && script != nil {
					// Try validation to see if it catches the error
					if validationErr := script.ValidateScript(); validationErr == nil {
						t.Error("Expected validation error but got none")
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if script == nil {
					t.Error("Script is nil")
				}
			}
		})
	}
}

func TestNewInheritanceScript_DifferentChainParams(t *testing.T) {
	ownerPubKey, inheritorPubKey := createTestPubKeys()
	timelockDays := int64(365)

	chainParams := []*chaincfg.Params{
		&chaincfg.TestNet3Params,
		&chaincfg.MainNetParams,
		&chaincfg.RegressionNetParams,
	}

	for _, params := range chainParams {
		t.Run(params.Name, func(t *testing.T) {
			script, err := NewInheritanceScript(ownerPubKey, inheritorPubKey, timelockDays, params)
			if err != nil {
				t.Fatalf("NewInheritanceScript failed: %v", err)
			}

			if script == nil {
				t.Fatal("Script is nil")
			}

			if script.ChainParams != params {
				t.Error("Chain parameters not stored correctly")
			}

			// Verify P2WSH address can be generated for this chain
			addr, err := script.GetP2WSHAddress()
			if err != nil {
				t.Errorf("Failed to generate P2WSH address: %v", err)
			}
			if addr == nil {
				t.Error("P2WSH address is nil")
			}
		})
	}
}

func TestNewInheritanceScript_ScriptConsistency(t *testing.T) {
	ownerPubKey, inheritorPubKey := createTestPubKeys()
	timelockDays := int64(365)
	chainParams := &chaincfg.TestNet3Params

	// Create multiple scripts with same parameters
	script1, err := NewInheritanceScript(ownerPubKey, inheritorPubKey, timelockDays, chainParams)
	if err != nil {
		t.Fatalf("First script creation failed: %v", err)
	}

	script2, err := NewInheritanceScript(ownerPubKey, inheritorPubKey, timelockDays, chainParams)
	if err != nil {
		t.Fatalf("Second script creation failed: %v", err)
	}

	// Verify scripts are identical
	if !bytes.Equal(script1.RedeemScript, script2.RedeemScript) {
		t.Error("Scripts with identical parameters should generate identical redeem scripts")
	}

	if script1.RelativeTimelock != script2.RelativeTimelock {
		t.Error("Scripts with identical parameters should have identical relative timelocks")
	}

	// Verify addresses are identical
	addr1, err := script1.GetP2WSHAddress()
	if err != nil {
		t.Fatalf("Failed to get address from first script: %v", err)
	}

	addr2, err := script2.GetP2WSHAddress()
	if err != nil {
		t.Fatalf("Failed to get address from second script: %v", err)
	}

	if addr1.EncodeAddress() != addr2.EncodeAddress() {
		t.Error("Scripts with identical parameters should generate identical addresses")
	}
}

func TestNewInheritanceScript_RedeemScriptStructure(t *testing.T) {
	ownerPubKey, inheritorPubKey := createTestPubKeys()
	timelockDays := int64(365)
	chainParams := &chaincfg.TestNet3Params

	script, err := NewInheritanceScript(ownerPubKey, inheritorPubKey, timelockDays, chainParams)
	if err != nil {
		t.Fatalf("NewInheritanceScript failed: %v", err)
	}

	redeemScript := script.RedeemScript

	// Verify script starts with OP_IF (0x63)
	if len(redeemScript) == 0 || redeemScript[0] != 0x63 {
		t.Error("Redeem script should start with OP_IF")
	}

	// Verify script contains the owner public key
	if !bytes.Contains(redeemScript, ownerPubKey) {
		t.Error("Redeem script should contain owner public key")
	}

	// Verify script contains the inheritor public key
	if !bytes.Contains(redeemScript, inheritorPubKey) {
		t.Error("Redeem script should contain inheritor public key")
	}

	// Verify script contains OP_CHECKSIG (0xac)
	if !bytes.Contains(redeemScript, []byte{0xac}) {
		t.Error("Redeem script should contain OP_CHECKSIG")
	}

	// Verify script contains OP_CHECKSEQUENCEVERIFY (0xb2)
	if !bytes.Contains(redeemScript, []byte{0xb2}) {
		t.Error("Redeem script should contain OP_CHECKSEQUENCEVERIFY")
	}

	// Verify script ends with OP_ENDIF (0x68)
	if len(redeemScript) == 0 || redeemScript[len(redeemScript)-1] != 0x68 {
		t.Error("Redeem script should end with OP_ENDIF")
	}
}

func TestNewInheritanceScript_NegativeTimelock(t *testing.T) {
	ownerPubKey, inheritorPubKey := createTestPubKeys()
	timelockDays := int64(-1)
	chainParams := &chaincfg.TestNet3Params

	script, err := NewInheritanceScript(ownerPubKey, inheritorPubKey, timelockDays, chainParams)
	if err != nil {
		t.Fatalf("NewInheritanceScript failed: %v", err)
	}

	// Script creation should succeed, but validation should fail
	if err := script.ValidateScript(); err == nil {
		t.Error("Expected validation to fail for negative timelock")
	}
}

func TestNewInheritanceScript_TimelockCalculation(t *testing.T) {
	ownerPubKey, inheritorPubKey := createTestPubKeys()
	chainParams := &chaincfg.TestNet3Params

	testCases := []struct {
		name          string
		days          int64
		expectedBit22 bool // Whether bit 22 should be set (time-based)
	}{
		{"1 day", 1, true},
		{"30 days", 30, true},
		{"365 days", 365, true},
		{"0 days", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			script, err := NewInheritanceScript(ownerPubKey, inheritorPubKey, tc.days, chainParams)
			if err != nil {
				t.Fatalf("NewInheritanceScript failed: %v", err)
			}

			// Check if bit 22 is set (time-based timelock)
			bit22Set := (script.RelativeTimelock & 0x400000) != 0
			if bit22Set != tc.expectedBit22 {
				t.Errorf("Expected bit 22 set: %v, got: %v", tc.expectedBit22, bit22Set)
			}

			// Verify the timelock value without the type bit
			timelockValue := script.RelativeTimelock & 0x3FFFFF
			expectedIntervals := (tc.days * 24 * 60 * 60) / 512
			if timelockValue != expectedIntervals {
				t.Errorf("Expected timelock intervals %d, got %d", expectedIntervals, timelockValue)
			}
		})
	}
}

func TestNewInheritanceScript_SameKeysError(t *testing.T) {
	ownerPubKey, _ := createTestPubKeys()
	// Use the same key for both owner and inheritor
	inheritorPubKey := ownerPubKey
	timelockDays := int64(365)
	chainParams := &chaincfg.TestNet3Params

	script, err := NewInheritanceScript(ownerPubKey, inheritorPubKey, timelockDays, chainParams)
	if err != nil {
		t.Fatalf("NewInheritanceScript failed: %v", err)
	}

	// While the script creation succeeds, this creates a potentially problematic scenario
	// where the owner and inheritor are the same entity
	if !bytes.Equal(script.OwnerPubKey, script.InheritorPubKey) {
		t.Error("Owner and inheritor public keys should be identical in this test")
	}

	// The script should still be valid from a technical perspective
	if err := script.ValidateScript(); err != nil {
		t.Errorf("Script validation should pass even with same keys: %v", err)
	}
}

func TestNewInheritanceScript_P2WSHAddressGeneration(t *testing.T) {
	ownerPubKey, inheritorPubKey := createTestPubKeys()
	timelockDays := int64(365)
	chainParams := &chaincfg.TestNet3Params

	script, err := NewInheritanceScript(ownerPubKey, inheritorPubKey, timelockDays, chainParams)
	if err != nil {
		t.Fatalf("NewInheritanceScript failed: %v", err)
	}

	// Test P2WSH address generation
	addr, err := script.GetP2WSHAddress()
	if err != nil {
		t.Fatalf("Failed to generate P2WSH address: %v", err)
	}

	if addr == nil {
		t.Fatal("P2WSH address is nil")
	}

	addrStr := addr.EncodeAddress()
	if addrStr == "" {
		t.Error("P2WSH address string is empty")
	}

	// For testnet, addresses should start with "tb1"
	if !strings.HasPrefix(addrStr, "tb1") {
		t.Errorf("Testnet P2WSH address should start with 'tb1', got: %s", addrStr)
	}
}

func TestNewInheritanceScript_ScriptHash(t *testing.T) {
	ownerPubKey, inheritorPubKey := createTestPubKeys()
	timelockDays := int64(365)
	chainParams := &chaincfg.TestNet3Params

	script, err := NewInheritanceScript(ownerPubKey, inheritorPubKey, timelockDays, chainParams)
	if err != nil {
		t.Fatalf("NewInheritanceScript failed: %v", err)
	}

	// Test script hash generation
	scriptHash := script.GetScriptHash()
	if len(scriptHash) != 32 {
		t.Errorf("Script hash should be 32 bytes, got %d", len(scriptHash))
	}

	// Test script pubkey generation
	scriptPubKey, err := script.GetScriptPubKey()
	if err != nil {
		t.Fatalf("Failed to generate script pubkey: %v", err)
	}

	if len(scriptPubKey) == 0 {
		t.Error("Script pubkey is empty")
	}

	// P2WSH scriptPubKey should be 34 bytes (OP_0 + 32-byte hash)
	if len(scriptPubKey) != 34 {
		t.Errorf("P2WSH scriptPubKey should be 34 bytes, got %d", len(scriptPubKey))
	}

	// Should start with OP_0 (0x00)
	if scriptPubKey[0] != 0x00 {
		t.Errorf("P2WSH scriptPubKey should start with OP_0, got 0x%02x", scriptPubKey[0])
	}

	// Second byte should be 0x20 (32 bytes push)
	if scriptPubKey[1] != 0x20 {
		t.Errorf("P2WSH scriptPubKey second byte should be 0x20, got 0x%02x", scriptPubKey[1])
	}
}

// Benchmark tests for performance measurement
func BenchmarkNewInheritanceScript(b *testing.B) {
	ownerPubKey, inheritorPubKey := createTestPubKeys()
	timelockDays := int64(365)
	chainParams := &chaincfg.TestNet3Params

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewInheritanceScript(ownerPubKey, inheritorPubKey, timelockDays, chainParams)
		if err != nil {
			b.Fatalf("NewInheritanceScript failed: %v", err)
		}
	}
}

func BenchmarkNewInheritanceScript_WithValidation(b *testing.B) {
	ownerPubKey, inheritorPubKey := createTestPubKeys()
	timelockDays := int64(365)
	chainParams := &chaincfg.TestNet3Params

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		script, err := NewInheritanceScript(ownerPubKey, inheritorPubKey, timelockDays, chainParams)
		if err != nil {
			b.Fatalf("NewInheritanceScript failed: %v", err)
		}
		if err := script.ValidateScript(); err != nil {
			b.Fatalf("Script validation failed: %v", err)
		}
	}
}

func BenchmarkNewInheritanceScript_WithAddressGeneration(b *testing.B) {
	ownerPubKey, inheritorPubKey := createTestPubKeys()
	timelockDays := int64(365)
	chainParams := &chaincfg.TestNet3Params

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		script, err := NewInheritanceScript(ownerPubKey, inheritorPubKey, timelockDays, chainParams)
		if err != nil {
			b.Fatalf("NewInheritanceScript failed: %v", err)
		}
		_, err = script.GetP2WSHAddress()
		if err != nil {
			b.Fatalf("GetP2WSHAddress failed: %v", err)
		}
	}
}
