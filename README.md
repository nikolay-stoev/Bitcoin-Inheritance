# Bitcoin Inheritance Protocol

A Bitcoin inheritance application that creates time-locked contracts using OP_IF/OP_ELSE conditional logic and OP_CHECKSEQUENCEVERIFY (CSV) timelocks.

## Overview

This project implements a Bitcoin inheritance protocol that allows:
- **Owner**: Can spend funds at any time using the IF path
- **Inheritor**: Can spend funds after a time delay using the ELSE path with OP_CHECKSEQUENCEVERIFY

## Architecture

The system uses a Bitcoin Script with the following structure:

```
OP_IF
    # Path 1: Owner's Immediate Spend
    <Owner_PublicKey> OP_CHECKSIG
OP_ELSE
    # Path 2: Inheritor's Time-Delayed Spend
    <Relative_Timelock_Value> OP_CHECKSEQUENCEVERIFY OP_DROP
    <Inheritor_PublicKey> OP_CHECKSIG
OP_ENDIF
```

## Project Structure

```
├── config/          # Configuration management
│   └── config.go    # Network and contract settings
├── keys/            # Cryptographic key management
│   └── keys.go      # Key generation and WIF handling
├── script/          # Bitcoin script construction
│   └── script.go    # Inheritance script building
├── transaction/     # Transaction building and signing
│   └── transaction.go # TX construction and validation
└── main.go          # CLI application entry point
```

## Dependencies

- `github.com/btcsuite/btcd` - Bitcoin protocol implementation
- `github.com/btcsuite/btcd/btcec/v2` - Cryptographic functions
- `github.com/btcsuite/btcd/btcutil` - Bitcoin utilities
- `github.com/btcsuite/btcd/chaincfg` - Network parameters
- `github.com/btcsuite/btcd/txscript` - Script building and signing
- `github.com/spf13/cobra` - CLI framework

## Requirements

- Go 1.24 or later
- Bitcoin testnet node (btcd) for full functionality

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd bitcoin-inheritance
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o bitcoin-inheritance
```

## Usage

### Generate a New Contract

```bash
./bitcoin-inheritance generate --testnet --timelock-days 180
```

This will:
1. Generate new key pairs for owner and inheritor. You will need to modify the script to use existing keys.
2. Create the inheritance script
3. Derive a P2WSH funding address
4. Display the contract details

### Owner Withdrawal (Placeholder)

```bash
./bitcoin-inheritance owner-withdraw --testnet
```

### Inheritor Withdrawal (Placeholder)

```bash
./bitcoin-inheritance inheritor-withdraw --testnet
```

## Configuration

The application supports both testnet and mainnet:

- **Testnet** (default): Safe for development with worthless coins
- **Mainnet**: Connect to Bitcoin mainnet by modifying the config (use with caution!)

Configuration includes:
- Network parameters (testnet/mainnet)
- RPC connection settings for btcd
- Default timelock duration (180 days)
- Default transaction fee (2000 satoshis)

## Bitcoin Testnet Setup

For development, you'll need:

1. **btcd testnet node** running with RPC enabled
2. **Testnet Bitcoin** from faucets:
   - https://testnet-faucet.mempool.co/
   - https://tbtc.bitaps.com/
   - https://bitcoinfaucet.uo1.net/

## Security Considerations

- **Testnet Only**: Current implementation is for testnet development
- **Key Management**: Private keys are generated fresh each time
- **Fee Management**: Uses static fees (should be dynamic in production)
- **Script Validation**: Basic validation is implemented
- **Timelock Encoding**: Properly implements BIP 68 relative timelock encoding

## Future Enhancements

- **Taproot Support**: Implement Taproot-based contracts for better privacy
- **Dynamic Fee Estimation**: Connect to fee estimation services
- **Key Persistence**: Save/load keys from secure storage
- **Transaction Broadcasting**: Implement actual RPC client for broadcasting
- **Script Execution**: Add off-chain script validation
- **Monitoring**: Add transaction confirmation monitoring

## License

This project is for educational and development purposes. Use at your own risk.
