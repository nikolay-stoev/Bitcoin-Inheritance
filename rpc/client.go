package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/btcsuite/btcd/wire"
	"github.com/nikolay.stoev/bitcoin-inheritance/config"
)

// RPCClient provides Bitcoin RPC functionality
type RPCClient struct {
	config *config.RPCConfig
	client *http.Client
}

// RPCRequest represents a Bitcoin RPC request
type RPCRequest struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	ID     int           `json:"id"`
}

// RPCResponse represents a Bitcoin RPC response
type RPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
	ID     int             `json:"id"`
}

// RPCError represents an RPC error
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewRPCClient creates a new RPC client
func NewRPCClient(cfg *config.RPCConfig) *RPCClient {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &RPCClient{
		config: cfg,
		client: client,
	}
}

// BroadcastTransaction broadcasts a transaction to the Bitcoin network
func (r *RPCClient) BroadcastTransaction(tx *wire.MsgTx) (string, error) {
	// Serialize transaction to hex
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return "", fmt.Errorf("failed to serialize transaction: %w", err)
	}

	txHex := fmt.Sprintf("%x", buf.Bytes())

	// Call sendrawtransaction RPC method
	result, err := r.call("sendrawtransaction", []interface{}{txHex})
	if err != nil {
		return "", fmt.Errorf("failed to broadcast transaction: %w", err)
	}

	var txid string
	if err := json.Unmarshal(result, &txid); err != nil {
		return "", fmt.Errorf("failed to parse transaction ID: %w", err)
	}

	return txid, nil
}

// GetBlockCount returns the current block count
func (r *RPCClient) GetBlockCount() (int64, error) {
	result, err := r.call("getblockcount", []interface{}{})
	if err != nil {
		return 0, fmt.Errorf("failed to get block count: %w", err)
	}

	var blockCount int64
	if err := json.Unmarshal(result, &blockCount); err != nil {
		return 0, fmt.Errorf("failed to parse block count: %w", err)
	}

	return blockCount, nil
}

// TestConnection tests the RPC connection
func (r *RPCClient) TestConnection() error {
	_, err := r.GetBlockCount()
	return err
}

// UTXO represents an unspent transaction output from RPC
type UTXO struct {
	TxID          string  `json:"txid"`
	Vout          uint32  `json:"vout"`
	Address       string  `json:"address"`
	Amount        float64 `json:"amount"`
	Confirmations int64   `json:"confirmations"`
	ScriptPubKey  string  `json:"scriptPubKey"`
}

// ListUnspent returns unspent outputs for a given address
func (r *RPCClient) ListUnspent(address string) ([]*UTXO, error) {
	result, err := r.call("listunspent", []interface{}{0, 9999999, []string{address}})
	if err != nil {
		return nil, fmt.Errorf("failed to list unspent outputs: %w", err)
	}

	var utxos []*UTXO
	if err := json.Unmarshal(result, &utxos); err != nil {
		return nil, fmt.Errorf("failed to parse unspent outputs: %w", err)
	}

	return utxos, nil
}

// GetTransaction gets detailed information about a transaction
func (r *RPCClient) GetTransaction(txid string) (json.RawMessage, error) {
	result, err := r.call("getrawtransaction", []interface{}{txid, true})
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	return result, nil
}

// call makes an RPC call to the Bitcoin node
func (r *RPCClient) call(method string, params []interface{}) (json.RawMessage, error) {
	// Create RPC request
	request := RPCRequest{
		Method: method,
		Params: params,
		ID:     1,
	}

	// Marshal request to JSON
	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("http://%s", r.config.Host)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers and authentication
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(r.config.User, r.config.Pass)

	// Make the request
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Parse RPC response
	var rpcResp RPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to parse RPC response: %w", err)
	}

	// Check for RPC error
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}
