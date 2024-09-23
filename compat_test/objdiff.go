//go:build compat_test

// MIT License
//
// Copyright (c) 2012-2020 Mat Ryer, Tyler Bunnell and contributors.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package compat_tests

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"runtime/debug"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"
)

// EqualBlocks compares two instances of types.Block and returns an error if they are not equal.
func EqualBlocks(block1, block2 *types.Block) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("block %q mismatch: %w", block1.Hash(), err)
		}
	}()
	if block1 == nil || block2 == nil {
		if block1 == block2 {
			return nil
		}
		return errors.New("one of the blocks is nil")
	}

	if block1.NumberU64() != block2.NumberU64() {
		return fmt.Errorf("block numbers do not match: %d != %d", block1.NumberU64(), block2.NumberU64())
	}

	if block1.Hash() != block2.Hash() {
		return fmt.Errorf("block hashes do not match: %s != %s", block1.Hash(), block2.Hash())
	}

	if block1.ParentHash() != block2.ParentHash() {
		return fmt.Errorf("parent hashes do not match: %s != %s", block1.ParentHash(), block2.ParentHash())
	}

	if block1.UncleHash() != block2.UncleHash() {
		return fmt.Errorf("uncle hashes do not match: %s != %s", block1.UncleHash(), block2.UncleHash())
	}

	if block1.Coinbase() != block2.Coinbase() {
		return fmt.Errorf("coinbase addresses do not match: %s != %s", block1.Coinbase(), block2.Coinbase())
	}

	if block1.Root() != block2.Root() {
		return fmt.Errorf("state roots do not match: %s != %s", block1.Root(), block2.Root())
	}

	if block1.TxHash() != block2.TxHash() {
		return fmt.Errorf("transaction roots do not match: %s != %s", block1.TxHash(), block2.TxHash())
	}

	if block1.ReceiptHash() != block2.ReceiptHash() {
		return fmt.Errorf("receipt roots do not match: %s != %s", block1.ReceiptHash(), block2.ReceiptHash())
	}

	if block1.Difficulty().Cmp(block2.Difficulty()) != 0 {
		return fmt.Errorf("difficulties do not match: %s != %s", block1.Difficulty(), block2.Difficulty())
	}

	if block1.GasLimit() != block2.GasLimit() {
		return fmt.Errorf("gas limits do not match: %d != %d", block1.GasLimit(), block2.GasLimit())
	}

	if block1.GasUsed() != block2.GasUsed() {
		return fmt.Errorf("gas used do not match: %d != %d", block1.GasUsed(), block2.GasUsed())
	}

	if block1.Time() != block2.Time() {
		return fmt.Errorf("timestamps do not match: %d != %d", block1.Time(), block2.Time())
	}

	if !bytes.Equal(block1.Extra(), block2.Extra()) {
		return fmt.Errorf("extra data do not match: %s != %s", block1.Extra(), block2.Extra())
	}

	if block1.MixDigest() != block2.MixDigest() {
		return fmt.Errorf("mix digests do not match: %s != %s", block1.MixDigest(), block2.MixDigest())
	}

	if block1.Nonce() != block2.Nonce() {
		return fmt.Errorf("nonces do not match: %d != %d", block1.Nonce(), block2.Nonce())
	}

	if err := EqualTransactionSlices(block1.Transactions(), block2.Transactions()); err != nil {
		return fmt.Errorf("transactions do not match: %w", err)
	}

	return nil
}

// EqualTransactionSlices compares two slices of types.Transaction and returns an error if they are not equal.
func EqualTransactionSlices(txs1, txs2 []*types.Transaction) error {
	if len(txs1) != len(txs2) {
		return errors.New("transaction slices are of different lengths")
	}

	for i := range txs1 {
		if err := EqualTransactions(txs1[i], txs2[i]); err != nil {
			return fmt.Errorf("transaction at index %d mismatch: %w", i, err)
		}
	}

	return nil
}

// EqualTransactions compares two instances of types.Transaction and returns an error if they are not equal.
func EqualTransactions(tx1, tx2 *types.Transaction) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("transaction %q mismatch: %w", tx1.Hash(), err)
		}
	}()
	if tx1 == nil || tx2 == nil {
		if tx1 == tx2 {
			return nil
		}
		return errors.New("one of the transactions is nil")
	}

	if tx1.Type() != tx2.Type() {
		return fmt.Errorf("types do not match: %d != %d", tx1.Type(), tx2.Type())
	}

	if tx1.Nonce() != tx2.Nonce() {
		return fmt.Errorf("nonces do not match: %d != %d", tx1.Nonce(), tx2.Nonce())
	}

	if tx1.GasPrice().Cmp(tx2.GasPrice()) != 0 {
		return fmt.Errorf("gas prices do not match: %s != %s", tx1.GasPrice(), tx2.GasPrice())
	}

	if tx1.GasFeeCap().Cmp(tx2.GasFeeCap()) != 0 {
		return fmt.Errorf("gas fee caps do not match: %s != %s", tx1.GasFeeCap(), tx2.GasFeeCap())
	}

	if tx1.GasTipCap().Cmp(tx2.GasTipCap()) != 0 {
		return fmt.Errorf("gas tip caps do not match: %s != %s", tx1.GasTipCap(), tx2.GasTipCap())
	}

	if tx1.Gas() != tx2.Gas() {
		return fmt.Errorf("gas limits do not match: %d != %d", tx1.Gas(), tx2.Gas())
	}

	if tx1.To() == nil && tx2.To() != nil || tx1.To() != nil && tx2.To() == nil {
		return errors.New("one of the recipient addresses is nil")
	}

	if tx1.To() != nil && tx2.To() != nil && *tx1.To() != *tx2.To() {
		return fmt.Errorf("recipient addresses do not match: %s != %s", tx1.To().Hex(), tx2.To().Hex())
	}

	if tx1.Value().Cmp(tx2.Value()) != 0 {
		return fmt.Errorf("values do not match: %s != %s", tx1.Value(), tx2.Value())
	}

	if !reflect.DeepEqual(tx1.Data(), tx2.Data()) {
		return errors.New("data payloads do not match")
	}

	if !reflect.DeepEqual(tx1.AccessList(), tx2.AccessList()) {
		return errors.New("access lists do not match")
	}

	if tx1.ChainId().Cmp(tx2.ChainId()) != 0 {
		return fmt.Errorf("chain IDs do not match: %s != %s", tx1.ChainId(), tx2.ChainId())
	}

	if tx1.Hash() != tx2.Hash() {
		return fmt.Errorf("hashes do not match: %s != %s", tx1.Hash().Hex(), tx2.Hash().Hex())
	}

	if tx1.Size() != tx2.Size() {
		return fmt.Errorf("sizes do not match: %d != %d", tx1.Size(), tx2.Size())
	}

	if tx1.Protected() != tx2.Protected() {
		return fmt.Errorf("protected flags do not match: %t != %t", tx1.Protected(), tx2.Protected())
	}
	r1, s1, v1 := tx1.RawSignatureValues()
	r2, s2, v2 := tx2.RawSignatureValues()
	if r1.Cmp(r2) != 0 || s1.Cmp(s2) != 0 || v1.Cmp(v2) != 0 {
		return errors.New("transaction signature values do not match")
	}

	return nil
}

// Since blocks can contain lists of transactions we just repeat the pattern for a mismatched atomic pointer.
var atomicPointerTransactionsOrBlockDiff = regexp.MustCompile(`^

Diff:
--- Expected
\+\+\+ Actual
(?:@@\s*-\d+,\d+\s*\+\d+,\d+\s*@@
\s*\},
-\s*v:\s*\(unsafe\.Pointer\)\s*0x[0-9a-f]+
\+\s*v:\s*\(unsafe\.Pointer\)\s*0x[0-9a-f]+
\s*\},
?)+
$`)

var atomicPointerTransactionDiff = regexp.MustCompile(`^

Diff:
--- Expected
\+\+\+\s*Actual
@@\s*-\d+,\d+\s*\+\d+,\d+\s*@@
\s*\},
-\s*v:\s*\(unsafe\.Pointer\)\s*0x[0-9a-f]+
\+\s*v:\s*\(unsafe\.Pointer\)\s*0x[0-9a-f]+
\s*\},
@@\s*-\d+,\d+\s*\+\d+,\d+\s*@@
\s*\},
-\s*v:\s*\(unsafe\.Pointer\)\s*0x[0-9a-f]+
\+\s*v:\s*\(unsafe\.Pointer\)\s*0x[0-9a-f]+
\s*\},
$`)

func EqualObjects(expected, actual interface{}, msgAndArgs ...interface{}) error {
	msg := messageFromMsgAndArgs(msgAndArgs...)
	if err := validateEqualArgs(expected, actual); err != nil {
		return fmt.Errorf("%s: Invalid operation: %#v == %#v (%s)", msg, expected, actual, err)
	}
	// A workaround for the atomic pointers now used to store the block and transaction hashes.
	b1, ok1 := expected.(*types.Block)
	b2, ok2 := actual.(*types.Block)
	if ok1 && ok2 {
		return EqualBlocks(b1, b2)
	}
	t1, ok1 := expected.(*types.Transaction)
	t2, ok2 := actual.(*types.Transaction)
	if ok1 && ok2 {
		return EqualTransactions(t1, t2)
	}
	ts1, ok1 := expected.([]*types.Transaction)
	ts2, ok2 := actual.([]*types.Transaction)
	if ok1 && ok2 {
		return EqualTransactionSlices(ts1, ts2)
	}

	if !assert.ObjectsAreEqual(expected, actual) {
		diff := diff(expected, actual)

		expected, actual = formatUnequalValues(expected, actual)
		return fmt.Errorf("%s: Not equal: \n"+
			"expected: %s\n"+
			"actual  : %s%s\n"+
			"stack:  : %s\n", msg, expected, actual, diff, string(debug.Stack()))
	}

	return nil
}

// validateEqualArgs checks whether provided arguments can be safely used in the
// Equal/NotEqual functions.
func validateEqualArgs(expected, actual interface{}) error {
	if expected == nil && actual == nil {
		return nil
	}

	if isFunction(expected) || isFunction(actual) {
		return errors.New("cannot take func type as argument")
	}
	return nil
}

func isFunction(arg interface{}) bool {
	if arg == nil {
		return false
	}
	return reflect.TypeOf(arg).Kind() == reflect.Func
}

// diff returns a diff of both values as long as both are of the same type and
// are a struct, map, slice, array or string. Otherwise it returns an empty string.
func diff(expected interface{}, actual interface{}) string {
	if expected == nil || actual == nil {
		return ""
	}

	et, ek := typeAndKind(expected)
	at, _ := typeAndKind(actual)

	if et != at {
		return ""
	}

	if ek != reflect.Struct && ek != reflect.Map && ek != reflect.Slice && ek != reflect.Array && ek != reflect.String {
		return ""
	}

	var e, a string

	switch et {
	case reflect.TypeOf(""):
		e = reflect.ValueOf(expected).String()
		a = reflect.ValueOf(actual).String()
	case reflect.TypeOf(time.Time{}):
		e = spewConfigStringerEnabled.Sdump(expected)
		a = spewConfigStringerEnabled.Sdump(actual)
	default:
		e = spewConfig.Sdump(expected)
		a = spewConfig.Sdump(actual)
	}

	diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(e),
		B:        difflib.SplitLines(a),
		FromFile: "Expected",
		FromDate: "",
		ToFile:   "Actual",
		ToDate:   "",
		Context:  1,
	})

	return "\n\nDiff:\n" + diff
}

var spewConfig = spew.ConfigState{
	Indent:                  " ",
	DisablePointerAddresses: true,
	DisableCapacities:       true,
	SortKeys:                true,
	DisableMethods:          true,
	MaxDepth:                10,
}

var spewConfigStringerEnabled = spew.ConfigState{
	Indent:                  " ",
	DisablePointerAddresses: true,
	DisableCapacities:       true,
	SortKeys:                true,
	MaxDepth:                10,
}

func typeAndKind(v interface{}) (reflect.Type, reflect.Kind) {
	t := reflect.TypeOf(v)
	k := t.Kind()

	if k == reflect.Ptr {
		t = t.Elem()
		k = t.Kind()
	}
	return t, k
}

// formatUnequalValues takes two values of arbitrary types and returns string
// representations appropriate to be presented to the user.
//
// If the values are not of like type, the returned strings will be prefixed
// with the type name, and the value will be enclosed in parenthesis similar
// to a type conversion in the Go grammar.
func formatUnequalValues(expected, actual interface{}) (e string, a string) {
	if reflect.TypeOf(expected) != reflect.TypeOf(actual) {
		return fmt.Sprintf("%T(%s)", expected, truncatingFormat(expected)),
			fmt.Sprintf("%T(%s)", actual, truncatingFormat(actual))
	}
	switch expected.(type) {
	case time.Duration:
		return fmt.Sprintf("%v", expected), fmt.Sprintf("%v", actual)
	}
	return truncatingFormat(expected), truncatingFormat(actual)
}

// truncatingFormat formats the data and truncates it if it's too long.
//
// This helps keep formatted error messages lines from exceeding the
// bufio.MaxScanTokenSize max line length that the go testing framework imposes.
func truncatingFormat(data interface{}) string {
	value := fmt.Sprintf("%#v", data)
	max := bufio.MaxScanTokenSize - 100 // Give us some space the type info too if needed.
	if len(value) > max {
		value = value[0:max] + "<... truncated>"
	}
	return value
}

func messageFromMsgAndArgs(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 || msgAndArgs == nil {
		return ""
	}
	if len(msgAndArgs) == 1 {
		msg := msgAndArgs[0]
		if msgAsStr, ok := msg.(string); ok {
			return msgAsStr
		}
		return fmt.Sprintf("%+v", msg)
	}
	if len(msgAndArgs) > 1 {
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	return ""
}
