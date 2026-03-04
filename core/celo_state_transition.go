package core

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/exchange"
	"github.com/ethereum/go-ethereum/contracts"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/uint256"
)

// canPayFee checks whether accountOwner's balance can cover transaction fee.
func (st *stateTransition) canPayFee(checkAmountForGas *big.Int) error {
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
			BlockNumber: st.evm.Context.BlockNumber,
			Time:        st.evm.Context.Time,
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

func (st *stateTransition) subFees(effectiveFee *big.Int) (err error) {
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

// calculateBaseFee returns the correct base fee to use during fee calculations
// This is the base fee from the header if no fee currency is used, but the
// base fee converted to fee currency when a fee currency is used.
func (st *stateTransition) calculateBaseFee() *big.Int {
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
