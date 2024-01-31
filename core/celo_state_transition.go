package core

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/exchange"
	"github.com/ethereum/go-ethereum/contracts"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/uint256"
)

// canPayFee checks whether accountOwner's balance can cover transaction fee.
func (st *stateTransition) canPayFee(checkAmount *uint256.Int) error {
	if st.msg.FeeCurrency == nil {
		balance := st.state.GetBalance(st.msg.From)

		if balance.Cmp(checkAmount) < 0 {
			return fmt.Errorf("%w: address %v have %v want %v", ErrInsufficientFunds, st.msg.From.Hex(), balance, checkAmount)
		}
	} else {
		backend := &contracts.CeloBackend{
			ChainConfig: st.evm.ChainConfig(),
			State:       st.state,
		}
		balance, err := contracts.GetBalanceERC20(backend, st.msg.From, *st.msg.FeeCurrency)
		if err != nil {
			return err
		}

		// Token amount can't be bigger than 256 bit
		balanceU256, _ := uint256.FromBig(balance)
		if balanceU256.Cmp(checkAmount) < 0 {
			return fmt.Errorf("%w: address %v have %v want %v, fee currency: %v", ErrInsufficientFunds, st.msg.From.Hex(), balance, checkAmount, st.msg.FeeCurrency.Hex())
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
		return contracts.DebitFees(st.evm, st.msg.FeeCurrency, st.msg.From, effectiveFee)
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
		baseFee, _ = exchange.ConvertCeloToCurrency(st.evm.Context.ExchangeRates, st.msg.FeeCurrency, baseFee)
	}

	return baseFee
}
