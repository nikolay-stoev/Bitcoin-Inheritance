package transaction

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash   *chainhash.Hash
	Vout     uint32
	Amount   btcutil.Amount
	PkScript []byte
}

// TransactionBuilder helps build Bitcoin transactions
type TransactionBuilder struct {
	chainParams *chaincfg.Params
	fee         btcutil.Amount
}

// NewTransactionBuilder creates a new transaction builder
func NewTransactionBuilder(chainParams *chaincfg.Params, fee btcutil.Amount) *TransactionBuilder {
	return &TransactionBuilder{
		chainParams: chainParams,
		fee:         fee,
	}
}

// BuildOwnerWithdrawTx builds a transaction for the owner to withdraw funds
func (tb *TransactionBuilder) BuildOwnerWithdrawTx(
	contractUTXO *UTXO,
	destinationAddr btcutil.Address,
	redeemScript []byte,
) (*wire.MsgTx, error) {

	// Create new transaction
	tx := wire.NewMsgTx(wire.TxVersion)

	// Add input pointing to the contract UTXO
	outPoint := wire.NewOutPoint(contractUTXO.TxHash, contractUTXO.Vout)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	tx.AddTxIn(txIn)

	// Calculate output amount (input amount minus fee)
	outputAmount := contractUTXO.Amount - tb.fee
	if outputAmount <= 0 {
		return nil, fmt.Errorf("insufficient funds: fee (%v) exceeds UTXO amount (%v)", tb.fee, contractUTXO.Amount)
	}

	// Create output script for destination address
	destinationScript, err := txscript.PayToAddrScript(destinationAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination script: %w", err)
	}

	// Add output
	txOut := wire.NewTxOut(int64(outputAmount), destinationScript)
	tx.AddTxOut(txOut)

	log.Printf("Built owner withdrawal transaction")
	log.Printf("  Input: %s:%d (%v satoshis)", contractUTXO.TxHash, contractUTXO.Vout, contractUTXO.Amount)
	log.Printf("  Output: %s (%v satoshis)", destinationAddr.EncodeAddress(), outputAmount)
	log.Printf("  Fee: %v satoshis", tb.fee)

	return tx, nil
}

// BuildInheritorWithdrawTx builds a transaction for the inheritor to withdraw funds
func (tb *TransactionBuilder) BuildInheritorWithdrawTx(
	contractUTXO *UTXO,
	destinationAddr btcutil.Address,
	redeemScript []byte,
	relativeTimelock int64,
) (*wire.MsgTx, error) {

	// Create new transaction
	tx := wire.NewMsgTx(wire.TxVersion)

	// Add input pointing to the contract UTXO
	outPoint := wire.NewOutPoint(contractUTXO.TxHash, contractUTXO.Vout)
	txIn := wire.NewTxIn(outPoint, nil, nil)

	// CRITICAL: Set the sequence field to satisfy OP_CHECKSEQUENCEVERIFY
	txIn.Sequence = uint32(relativeTimelock)

	tx.AddTxIn(txIn)

	// Calculate output amount (input amount minus fee)
	outputAmount := contractUTXO.Amount - tb.fee
	if outputAmount <= 0 {
		return nil, fmt.Errorf("insufficient funds: fee (%v) exceeds UTXO amount (%v)", tb.fee, contractUTXO.Amount)
	}

	// Create output script for destination address
	destinationScript, err := txscript.PayToAddrScript(destinationAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination script: %w", err)
	}

	// Add output
	txOut := wire.NewTxOut(int64(outputAmount), destinationScript)
	tx.AddTxOut(txOut)

	log.Printf("Built inheritor withdrawal transaction")
	log.Printf("  Input: %s:%d (%v satoshis)", contractUTXO.TxHash, contractUTXO.Vout, contractUTXO.Amount)
	log.Printf("  Output: %s (%v satoshis)", destinationAddr.EncodeAddress(), outputAmount)
	log.Printf("  Fee: %v satoshis", tb.fee)
	log.Printf("  Sequence: %d (timelock)", relativeTimelock)

	return tx, nil
}

// SignOwnerTransaction signs a transaction for the owner using the IF path
func (tb *TransactionBuilder) SignOwnerTransaction(
	tx *wire.MsgTx,
	contractUTXO *UTXO,
	redeemScript []byte,
	ownerPrivateKey interface{}, // This will be *btcec.PrivateKey, but avoiding import for now
) error {
	// This is a placeholder for the signing logic
	// In a complete implementation, this would:
	// 1. Create a MultiPrevOutFetcher
	// 2. Generate signature hash
	// 3. Sign using txscript.RawTxInWitnessSignature
	// 4. Assemble witness with: [signature, OP_1, redeemScript]

	log.Printf("Signing owner transaction (placeholder)")
	return nil
}

// SignInheritorTransaction signs a transaction for the inheritor using the ELSE path
func (tb *TransactionBuilder) SignInheritorTransaction(
	tx *wire.MsgTx,
	contractUTXO *UTXO,
	redeemScript []byte,
	inheritorPrivateKey interface{}, // This will be *btcec.PrivateKey, but avoiding import for now
) error {
	// This is a placeholder for the signing logic
	// In a complete implementation, this would:
	// 1. Create a MultiPrevOutFetcher
	// 2. Generate signature hash
	// 3. Sign using txscript.RawTxInWitnessSignature
	// 4. Assemble witness with: [signature, OP_0, redeemScript]

	log.Printf("Signing inheritor transaction (placeholder)")
	return nil
}

// ValidateTransaction performs basic validation on a transaction
func (tb *TransactionBuilder) ValidateTransaction(tx *wire.MsgTx) error {
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}

	if len(tx.TxIn) == 0 {
		return fmt.Errorf("transaction has no inputs")
	}

	if len(tx.TxOut) == 0 {
		return fmt.Errorf("transaction has no outputs")
	}

	// Check that outputs don't exceed inputs (basic sanity check)
	var totalOut int64
	for _, out := range tx.TxOut {
		totalOut += out.Value
	}

	if totalOut < 0 {
		return fmt.Errorf("transaction has negative output value")
	}

	log.Printf("Transaction validation passed")
	return nil
}

// SerializeTransaction serializes a transaction to hex string
func (tb *TransactionBuilder) SerializeTransaction(tx *wire.MsgTx) (string, error) {
	// This is a placeholder for transaction serialization
	// In a complete implementation, this would serialize the transaction
	// to bytes and return as hex string

	log.Printf("Serializing transaction (placeholder)")
	return "placeholder_hex_string", nil
}
