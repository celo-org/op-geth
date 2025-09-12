package vm

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/contracts/addresses"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

var (
	TransferPrecompileAddress = common.HexToAddress("0x00000000000000000000000000000000000000fd")
)

type CeloPrecompiledContract interface {
	RequiredGas(input []byte) uint64                              // RequiredGas calculates the contract gas use
	Run(input []byte, ctx *celoPrecompileContext) ([]byte, error) // Run runs the precompiled contract
}

type wrap struct {
	PrecompiledContract
}

func (pw *wrap) Run(input []byte, ctx *celoPrecompileContext) ([]byte, error) {
	return pw.PrecompiledContract.Run(input)
}

type celoPrecompileContext struct {
	*BlockContext
	*params.Rules

	caller common.Address
	evm    *EVM
}

func NewContext(caller common.Address, evm *EVM) *celoPrecompileContext {
	return &celoPrecompileContext{
		BlockContext: &evm.Context,
		Rules:        &evm.chainRules,
		caller:       caller,
		evm:          evm,
	}
}

func (ctx *celoPrecompileContext) IsCallerCeloToken() (bool, error) {
	tokenAddress := addresses.GetAddressesOrDefault(ctx.evm.ChainConfig().ChainID, addresses.MainnetAddresses).CeloToken

	return tokenAddress == ctx.caller, nil
}

// Native transfer contract to make CELO ERC20 compatible.
type transfer struct{}

func (c *transfer) RequiredGas(input []byte) uint64 {
	return params.CallValueTransferGas
}

func (c *transfer) Run(input []byte, ctx *celoPrecompileContext) ([]byte, error) {
	// STATICCALL protection: prevent state changes in read-only context
	if ctx.evm.readOnly {
		return nil, ErrWriteProtection
	}

	if isCeloToken, err := ctx.IsCallerCeloToken(); err != nil {
		return nil, err
	} else if !isCeloToken {
		return nil, fmt.Errorf("unable to call transfer from unpermissioned address")
	}

	// input is comprised of 3 arguments:
	//   from:  32 bytes representing the address of the sender
	//   to:    32 bytes representing the address of the recipient
	//   value: 32 bytes, a 256 bit integer representing the amount of CELO to transfer
	// 3 arguments x 32 bytes each = 96 bytes total input
	if len(input) != 96 {
		return nil, ErrInputLength
	}

	// The from parameter should always be set to the caller of the
	// function calling the precompile (tx.sender in Solidity). Reasons why
	// we have to pass that into the precompile from outside:
	// * We can't use ctx.caller because that is always the CELO token
	// * We can't use ctx.evm.Origin because that would limit usage to EOA accounts
	// * The real value we could use is the caller's caller, which is not readily available
	from := common.BytesToAddress(input[0:32])
	to := common.BytesToAddress(input[32:64])

	var parsed bool
	value, parsed := math.ParseBig256(hexutil.Encode(input[64:96]))
	if !parsed {
		return nil, fmt.Errorf("Error parsing transfer: unable to parse value from %s", hexutil.Encode(input[64:96]))
	}
	valueU256, overflow := uint256.FromBig(value)
	if overflow {
		return nil, fmt.Errorf("Error parsing transfer: value overflow")
	}

	// The mockEVM doesn't have a StateDB set up, so ctx.evm.StateDB is nil
	// in the tests. We could add a mockStateDB to it to avoid this check.
	if ctx.evm.StateDB != nil && ctx.IsOptimismJovian {
		// Warm the from and to addresses by adding them to the access list. We didn't
		// do this when Celo launched, but it is more consistent with Ethereum, where
		// all accessed addresses are always warmed.
		// See also https://github.com/celo-org/celo-blockchain-planning/issues/1226
		ctx.evm.StateDB.AddAddressToAccessList(from)
		ctx.evm.StateDB.AddAddressToAccessList(to)
	}

	// Fail if we're trying to transfer more than the available balance
	if !ctx.CanTransfer(ctx.evm.StateDB, from, valueU256) {
		return nil, ErrInsufficientBalance
	}

	ctx.Transfer(ctx.evm.StateDB, from, to, valueU256)

	return nil, nil
}
