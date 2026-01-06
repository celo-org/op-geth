# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is the Celo L2 execution client, a fork of Optimism's op-geth, which itself is a fork of go-ethereum (geth). It serves as the execution layer for the Celo L2 blockchain, implementing Celo-specific features like fee currencies while maintaining compatibility with Ethereum and Optimism.

## Development Commands

### Building
- `make geth` - Build the main geth binary
- `make all` - Build all packages and executables

### Code Quality
- `make fmt` - Format Go code with gofmt
- `make lint` - Run linters for code quality checks

### Testing
- `gotestsum` - Run all tests
- `./e2e_test/run_all_tests.sh` - Run comprehensive E2E test suite
- Individual E2E tests in `e2e_test/test_*.sh`

## Core Architecture

### Celo-Specific Features

**Fee Currencies**: The primary Celo innovation allowing transaction fees to be paid in ERC-20 tokens instead of native CELO:
- Implementation: `contracts/fee_currencies.go`, `core/celo_state_transition.go`
- New transaction types: `CeloDynamicFeeTx`, `CeloDynamicFeeTxV2` in `core/types/`
- Multi-gas pools: `core/celo_multi_gaspool.go`

**Token Duality**: CELO functions as both native currency and ERC-20:
- Precompiled transfer contract at address `0xfd`
- Implementation: `core/vm/celo_contracts.go`

**Currency Blocklisting**: Protection against problematic fee currencies:
- Time-based blocking/eviction system
- Implementation: `miner/currency_blocklist.go`

**Transaction Pool Extensions**: Celo-specific validation and handling:
- Fee currency validation: `core/txpool/celo_validation.go`
- Legacy pool extensions: `core/txpool/legacypool/celo_*.go`
- Blob pool extensions: `core/txpool/blobpool/celo_blobpool.go`

**State Database Hooks**: Custom state transition handling:
- Hooked state DB: `core/state/celo_statedb_hooked.go`
- Genesis handling: `core/celo_genesis.go`
- EVM extensions: `core/celo_evm.go`, `core/vm/celo_evm.go`

**Legacy Transaction Support**: Backward compatibility with Celo L1:
- Legacy transaction types: `core/types/celo_tx_legacy.go`
- Block and receipt types: `core/types/celo_block.go`, `core/types/celo_receipt.go`

### Key Directories

**Core Blockchain Logic**:
- `core/` - State transition, blockchain processing, transaction pools
- `core/types/` - Celo transaction types, receipts, blocks
- `core/vm/` - EVM modifications and Celo precompiled contracts

**Celo Integration**:
- `contracts/` - Fee currency logic, contract bindings
- `contracts/celo/` - Generated Go bindings for Celo contracts
- `internal/celoapi/` - Celo-specific RPC methods
- `internal/ethapi/` - Celo extensions to Ethereum API (celo_*.go files)
- `internal/sequencerapi/` - Sequencer API implementations

**Configuration**:
- `params/` - Chain configurations, protocol parameters, fork definitions
- `params/celo_config.go` - Celo-specific chain parameters

**Mining & Transaction Processing**:
- `miner/` - Block building with Celo features (currency_blocklist.go, celo_defaults.go, etc.)
- `eth/gasestimator/` - Gas estimation for Celo transactions
- `eth/catalyst/` - Beacon chain integration with Celo support

### Fork Structure

This codebase maintains three-way compatibility:
1. **Ethereum**: Base functionality and EIPs
2. **Optimism**: Rollup functionality from op-geth
3. **Celo**: Fee currencies and token duality features

### Transaction Flow with Fee Currencies

1. Transaction specifies `FeeCurrency` field (nil = native CELO)
2. Exchange rates fetched from `FeeCurrencyDirectory` contract
3. Fee calculation uses exchange rates to convert from CELO
4. Balance validation ensures sufficient fee currency
5. Debit operation calls fee currency contract
6. Transaction execution proceeds normally
7. Credit operation distributes fees to recipients

### Network Configuration

- **Celo Mainnet**: Chain ID 42220
- **Alfajores Testnet**: Chain ID 44787
- **Baklava Testnet**: Chain ID 62320
- **Cel2Time**: Fork timestamp when Celo L2 features activate

## Testing Strategy

**Unit Tests**: Standard Go testing with `go test`
**Integration Tests**: E2E test suite covering fee currency flows
**Fork Tests**: Compatibility testing across Ethereum/Optimism/Celo features

Key test areas:
- Fee currency debit/credit operations
- Transaction type compatibility (legacy, dynamic fee V1/V2)
- Multi-gas pool management
- Currency blocklisting behavior
- Token duality functionality
- Transaction pool validation with fee currencies
- State database hooks and transitions
- Genesis block handling for Celo chains

## Code Patterns

**Celo Extensions**: Look for `celo_` prefixed files for Celo-specific implementations
- Core layer: `core/celo_*.go` - Genesis, EVM, state transitions, multi-gas pools
- Type extensions: `core/types/celo_*.go` - Transaction types, blocks, receipts
- State layer: `core/state/celo_*.go` - State database hooks
- Transaction pool: `core/txpool/**/celo_*.go` - Validation and pool management
- VM layer: `core/vm/celo_*.go` - Precompiled contracts and EVM extensions
- API layer: `internal/ethapi/celo_*.go` and `internal/celoapi/` - RPC methods
- Mining: `miner/celo_*.go` - Block building and defaults
- Storage: `core/rawdb/celo_*.go` - Database accessors

**Fee Currency Context**: Most fee operations use `FeeCurrencyContext` struct
**Chain Config**: Use `params.CeloChainConfig` for Celo-specific parameters
**Transaction Types**: Handle both legacy Ethereum and new Celo transaction types
