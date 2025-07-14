package keys

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

// KeyPair represents a Bitcoin key pair with convenience methods
type KeyPair struct {
	PrivateKey  *btcec.PrivateKey
	PublicKey   *btcec.PublicKey
	WIF         *btcutil.WIF
	ChainParams *chaincfg.Params
}

// InheritanceKeys holds the key pairs for owner and inheritor
type InheritanceKeys struct {
	Owner     *KeyPair
	Inheritor *KeyPair
}

// NewKeyPair generates a new cryptographically secure key pair
func NewKeyPair(chainParams *chaincfg.Params) (*KeyPair, error) {
	// Generate new private key
	privKey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Derive public key
	pubKey := privKey.PubKey()

	// Encode private key to WIF format (compressed)
	wif, err := btcutil.NewWIF(privKey, chainParams, true)
	if err != nil {
		return nil, fmt.Errorf("failed to encode WIF: %w", err)
	}

	return &KeyPair{
		PrivateKey:  privKey,
		PublicKey:   pubKey,
		WIF:         wif,
		ChainParams: chainParams,
	}, nil
}

// GetCompressedPubKeyBytes returns the compressed public key as bytes
func (kp *KeyPair) GetCompressedPubKeyBytes() []byte {
	return kp.PublicKey.SerializeCompressed()
}

// GetP2WPKHAddress returns a P2WPKH address for this key pair
func (kp *KeyPair) GetP2WPKHAddress() (btcutil.Address, error) {
	pubKeyBytes := kp.GetCompressedPubKeyBytes()
	pubKeyHash := btcutil.Hash160(pubKeyBytes)

	addr, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, kp.ChainParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create P2WPKH address: %w", err)
	}

	return addr, nil
}

// GenerateInheritanceKeys performs the "key ceremony" to generate keys for both parties
func GenerateInheritanceKeys(chainParams *chaincfg.Params) (*InheritanceKeys, error) {
	// Generate owner's key pair
	ownerKeys, err := NewKeyPair(chainParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate owner keys: %w", err)
	}

	// Generate inheritor's key pair
	inheritorKeys, err := NewKeyPair(chainParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate inheritor keys: %w", err)
	}

	log.Printf("Generated owner keys - WIF: %s", ownerKeys.WIF.String())
	log.Printf("Generated inheritor keys - WIF: %s", inheritorKeys.WIF.String())

	// Generate withdrawal addresses for logging
	ownerAddr, err := ownerKeys.GetP2WPKHAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to generate owner address: %w", err)
	}

	inheritorAddr, err := inheritorKeys.GetP2WPKHAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to generate inheritor address: %w", err)
	}

	log.Printf("Owner withdrawal address: %s", ownerAddr.EncodeAddress())
	log.Printf("Inheritor withdrawal address: %s", inheritorAddr.EncodeAddress())

	return &InheritanceKeys{
		Owner:     ownerKeys,
		Inheritor: inheritorKeys,
	}, nil
}

// KeyPairFromWIF creates a KeyPair from a WIF string
func KeyPairFromWIF(wifStr string, chainParams *chaincfg.Params) (*KeyPair, error) {
	wif, err := btcutil.DecodeWIF(wifStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode WIF: %w", err)
	}

	return &KeyPair{
		PrivateKey:  wif.PrivKey,
		PublicKey:   wif.PrivKey.PubKey(),
		WIF:         wif,
		ChainParams: chainParams,
	}, nil
}
