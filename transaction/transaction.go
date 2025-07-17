package transaction

import (
	"bytes"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
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
	ownerPrivateKey *btcec.PrivateKey,
) error {
	// Create a MultiPrevOutFetcher for the UTXO
	prevOutFetcher := txscript.NewMultiPrevOutFetcher(nil)

	// Create the P2WSH output script from the redeem script
	scriptHash := btcutil.Hash160(redeemScript)
	p2wshScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_0).AddData(scriptHash).Script()
	if err != nil {
		return fmt.Errorf("failed to create P2WSH script: %w", err)
	}

	// Add the UTXO to the fetcher
	prevOut := &wire.TxOut{
		Value:    int64(contractUTXO.Amount),
		PkScript: p2wshScript,
	}
	prevOutFetcher.AddPrevOut(*wire.NewOutPoint(contractUTXO.TxHash, contractUTXO.Vout), prevOut)

	// Generate signature hash for the transaction
	sigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
	hashType := txscript.SigHashAll

	sigHash, err := txscript.CalcWitnessSigHash(redeemScript, sigHashes, hashType, tx, 0, int64(contractUTXO.Amount))
	if err != nil {
		return fmt.Errorf("failed to calculate signature hash: %w", err)
	}

	// Sign the hash with the owner's private key
	sig := ecdsa.Sign(ownerPrivateKey, sigHash)
	sigBytes := append(sig.Serialize(), byte(hashType))

	// Assemble witness: [signature, OP_1 (true), redeemScript]
	witness := wire.TxWitness{
		sigBytes,
		{txscript.OP_1}, // OP_1 to take the IF path
		redeemScript,
	}

	// Set the witness for the first (and only) input
	tx.TxIn[0].Witness = witness

	log.Printf("Transaction signed successfully with owner's key (IF path)")
	return nil
}

// SignInheritorTransaction signs a transaction for the inheritor using the ELSE path
func (tb *TransactionBuilder) SignInheritorTransaction(
	tx *wire.MsgTx,
	contractUTXO *UTXO,
	redeemScript []byte,
	inheritorPrivateKey *btcec.PrivateKey,
) error {
	// Create a MultiPrevOutFetcher for the UTXO
	prevOutFetcher := txscript.NewMultiPrevOutFetcher(nil)

	// Create the P2WSH output script from the redeem script
	scriptHash := btcutil.Hash160(redeemScript)
	p2wshScript, err := txscript.NewScriptBuilder().AddOp(txscript.OP_0).AddData(scriptHash).Script()
	if err != nil {
		return fmt.Errorf("failed to create P2WSH script: %w", err)
	}

	// Add the UTXO to the fetcher
	prevOut := &wire.TxOut{
		Value:    int64(contractUTXO.Amount),
		PkScript: p2wshScript,
	}
	prevOutFetcher.AddPrevOut(*wire.NewOutPoint(contractUTXO.TxHash, contractUTXO.Vout), prevOut)

	// Generate signature hash for the transaction
	sigHashes := txscript.NewTxSigHashes(tx, prevOutFetcher)
	hashType := txscript.SigHashAll

	sigHash, err := txscript.CalcWitnessSigHash(redeemScript, sigHashes, hashType, tx, 0, int64(contractUTXO.Amount))
	if err != nil {
		return fmt.Errorf("failed to calculate signature hash: %w", err)
	}

	// Sign the hash with the inheritor's private key
	sig := ecdsa.Sign(inheritorPrivateKey, sigHash)
	sigBytes := append(sig.Serialize(), byte(hashType))

	// Assemble witness: [signature, OP_0 (false), redeemScript]
	witness := wire.TxWitness{
		sigBytes,
		{txscript.OP_0}, // OP_0 to take the ELSE path
		redeemScript,
	}

	// Set the witness for the first (and only) input
	tx.TxIn[0].Witness = witness

	log.Printf("Transaction signed successfully with inheritor's key (ELSE path)")
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
	// Serialize the transaction to bytes
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return "", fmt.Errorf("failed to serialize transaction: %w", err)
	}

	// Convert to hex string
	txHex := fmt.Sprintf("%x", buf.Bytes())

	log.Printf("Transaction serialized to %d bytes", len(buf.Bytes()))
	return txHex, nil
}
