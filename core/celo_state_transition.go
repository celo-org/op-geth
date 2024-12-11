package core

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/exchange"
	"github.com/ethereum/go-ethereum/contracts"
	"github.com/ethereum/go-ethereum/contracts/addresses"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

// IsFeeCurrencyDenominated returns whether the gas-price related
// fields are denominated in a given fee currency or in the native token.
// This effectively is only true for CIP-64 transactions.
func (msg *Message) IsFeeCurrencyDenominated() bool {
	return msg.FeeCurrency != nil && msg.MaxFeeInFeeCurrency == nil
}

// canPayFee checks whether accountOwner's balance can cover transaction fee.
func (st *StateTransition) canPayFee(checkAmountForGas *big.Int) error {
	var checkAmountInCelo, checkAmountInAlternativeCurrency *big.Int
	if st.msg.FeeCurrency == nil {
		checkAmountInCelo = new(big.Int).Add(checkAmountForGas, st.msg.Value)
		checkAmountInAlternativeCurrency = common.Big0
	} else {
		checkAmountInCelo = st.msg.Value
		checkAmountInAlternativeCurrency = checkAmountForGas
	}

	if checkAmountInCelo.Cmp(common.Big0) > 0 {
		balanceInCeloU256, overflow := uint256.FromBig(checkAmountInCelo)
		if overflow {
			return fmt.Errorf("%w: address %v required balance exceeds 256 bits", ErrInsufficientFunds, st.msg.From.Hex())
		}

		balance := st.state.GetBalance(st.msg.From)

		if balance.Cmp(balanceInCeloU256) < 0 {
			return fmt.Errorf("%w: address %v have %v want %v", ErrInsufficientFunds, st.msg.From.Hex(), balance, checkAmountInCelo)
		}
	}
	if checkAmountInAlternativeCurrency.Cmp(common.Big0) > 0 {
		_, overflow := uint256.FromBig(checkAmountInAlternativeCurrency)
		if overflow {
			return fmt.Errorf("%w: address %v required balance exceeds 256 bits", ErrInsufficientFunds, st.msg.From.Hex())
		}
		backend := &contracts.CeloBackend{
			ChainConfig: st.evm.ChainConfig(),
			State:       st.state,
		}
		balance, err := contracts.GetBalanceERC20(backend, st.msg.From, *st.msg.FeeCurrency)
		if err != nil {
			return err
		}

		if balance.Cmp(checkAmountInAlternativeCurrency) < 0 {
			return fmt.Errorf("%w: address %v have %v want %v, fee currency: %v", ErrInsufficientFunds, st.msg.From.Hex(), balance, checkAmountInAlternativeCurrency, st.msg.FeeCurrency.Hex())
		}
	}
	return nil
}

func (st *StateTransition) subFees(effectiveFee *big.Int) (err error) {
	log.Trace("Debiting fee", "from", st.msg.From, "amount", effectiveFee, "feeCurrency", st.msg.FeeCurrency)

	// native currency
	if st.msg.FeeCurrency == nil {
		effectiveFeeU256, _ := uint256.FromBig(effectiveFee)
		st.state.SubBalance(st.msg.From, effectiveFeeU256, tracing.BalanceDecreaseGasBuy)
		return nil
	} else {
		gasUsedDebit, err := contracts.DebitFees(st.evm, st.msg.FeeCurrency, st.msg.From, effectiveFee)
		st.feeCurrencyGasUsed += gasUsedDebit
		return err
	}
}

// distributeTxFees calculates the amounts and recipients of transaction fees and credits the accounts.
func (st *StateTransition) distributeTxFees() error {
	if st.evm.Config.NoBaseFee && st.msg.GasFeeCap.Sign() == 0 && st.msg.GasTipCap.Sign() == 0 {
		// Skip fee payment when NoBaseFee is set and the fee fields
		// are 0. This avoids a negative effectiveTip being applied to
		// the coinbase when simulating calls.
		return nil
	}

	// Determine the refund and transaction fee to be distributed.
	refund := new(big.Int).Mul(new(big.Int).SetUint64(st.gasRemaining), st.msg.GasPrice)
	gasUsed := new(big.Int).SetUint64(st.gasUsed())
	totalTxFee := new(big.Int).Mul(gasUsed, st.msg.GasPrice)
	from := st.msg.From

	// Divide the transaction into a base (the minimum transaction fee) and tip (any extra, or min(max tip, feecap - GPM) if espresso).
	baseTxFee := new(big.Int).Mul(gasUsed, st.calculateBaseFee())
	// No need to do effectiveTip calculation, because st.gasPrice == effectiveGasPrice, and effectiveTip = effectiveGasPrice - baseTxFee
	tipTxFee := new(big.Int).Sub(totalTxFee, baseTxFee)

	feeCurrency := st.msg.FeeCurrency
	feeHandlerAddress := addresses.GetAddresses(st.evm.ChainConfig().ChainID).FeeHandler

	log.Trace("distributeTxFees", "from", from, "refund", refund, "feeCurrency", feeCurrency,
		"coinbaseFeeRecipient", st.evm.Context.Coinbase, "coinbaseFee", tipTxFee,
		"feeHandler", feeHandlerAddress, "communityFundFee", baseTxFee)

	rules := st.evm.ChainConfig().Rules(st.evm.Context.BlockNumber, st.evm.Context.Random != nil, st.evm.Context.Time)
	var l1Cost *big.Int
	// Check that we are post bedrock to enable op-geth to be able to create pseudo pre-bedrock blocks (these are pre-bedrock, but don't follow l2 geth rules)
	// Note optimismConfig will not be nil if rules.IsOptimismBedrock is true
	if optimismConfig := st.evm.ChainConfig().Optimism; optimismConfig != nil &&
		rules.IsOptimismBedrock && !st.msg.IsDepositTx {
		l1Cost = st.evm.Context.L1CostFunc(st.msg.RollupCostData, st.evm.Context.Time)
	}

	if feeCurrency == nil {
		tipTxFeeU256, overflow := uint256.FromBig(tipTxFee)
		if overflow {
			return fmt.Errorf("celo tip overflows U256: %d", tipTxFee)
		}
		st.state.AddBalance(st.evm.Context.Coinbase, tipTxFeeU256, tracing.BalanceIncreaseRewardTransactionFee)

		refundU256, overflow := uint256.FromBig(refund)
		if overflow {
			return fmt.Errorf("celo refund overflows U256: %d", refund)
		}
		st.state.AddBalance(from, refundU256, tracing.BalanceIncreaseGasReturn)

		baseTxFeeU256, overflow := uint256.FromBig(baseTxFee)
		if overflow {
			return fmt.Errorf("celo base fee overflows U256: %d", baseTxFee)
		}
		if rules.IsCel2 {
			st.state.AddBalance(feeHandlerAddress, baseTxFeeU256, tracing.BalanceIncreaseRewardTransactionFee)
		} else if st.evm.ChainConfig().Optimism != nil {
			st.state.AddBalance(params.OptimismBaseFeeRecipient, baseTxFeeU256, tracing.BalanceIncreaseRewardTransactionFee)
		}

		l1CostU256, overflow := uint256.FromBig(l1Cost)
		if overflow {
			return fmt.Errorf("optimism l1 cost overflows U256: %d", l1Cost)
		}
		if l1Cost != nil {
			st.state.AddBalance(params.OptimismL1FeeRecipient, l1CostU256, tracing.BalanceIncreaseRewardTransactionFee)
		}
	} else {
		if l1Cost != nil {
			l1Cost, _ = exchange.ConvertCeloToCurrency(st.evm.Context.FeeCurrencyContext.ExchangeRates, feeCurrency, l1Cost)
		}
		if err := contracts.CreditFees(
			st.evm,
			feeCurrency,
			from,
			st.evm.Context.Coinbase,
			feeHandlerAddress,
			params.OptimismL1FeeRecipient,
			refund,
			tipTxFee,
			baseTxFee,
			l1Cost,
			st.feeCurrencyGasUsed,
		); err != nil {
			log.Error("Error crediting", "from", from, "coinbase", st.evm.Context.Coinbase, "feeHandler", feeHandlerAddress, "err", err)
			return err
		}
	}

	if st.evm.Config.Tracer != nil && st.evm.Config.Tracer.OnGasChange != nil && st.gasRemaining > 0 {
		st.evm.Config.Tracer.OnGasChange(st.gasRemaining, 0, tracing.GasChangeTxLeftOverReturned)
	}
	return nil
}

// calculateBaseFee returns the correct base fee to use during fee calculations
// This is the base fee from the header if no fee currency is used, but the
// base fee converted to fee currency when a fee currency is used.
func (st *StateTransition) calculateBaseFee() *big.Int {
	baseFee := st.evm.Context.BaseFee
	if baseFee == nil {
		// This can happen in pre EIP-1559 environments
		baseFee = big.NewInt(0)
	}

	if st.msg.FeeCurrency != nil {
		// Existence of the fee currency has been checked in `preCheck`
		baseFee, _ = exchange.ConvertCeloToCurrency(st.evm.Context.FeeCurrencyContext.ExchangeRates, st.msg.FeeCurrency, baseFee)
	}

	return baseFee
}
