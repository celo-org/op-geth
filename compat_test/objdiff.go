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

	if !assert.ObjectsAreEqual(expected, actual) {
		diff := diff(expected, actual)

		// A workaround for the atomic pointers now used to store the block and transaction hashes.
		b1, ok1 := expected.(*types.Block)
		b2, ok2 := actual.(*types.Block)
		if ok1 && ok2 {
			// So if the atomic pointers do not match but the hashes do we consider the blocks equal.
			if b1.Hash() == b2.Hash() && atomicPointerTransactionsOrBlockDiff.MatchString(diff) {
				return nil
			}
			// fmt.Printf("blockdiff matched (%v): %q\n", atomicPointerBlockDiff.MatchString(diff), diff)
		}
		t1, ok1 := expected.(*types.Transaction)
		t2, ok2 := actual.(*types.Transaction)
		if ok1 && ok2 {
			// So if the atomic pointers do not match but the hashes do we consider the blocks equal.
			if t1.Hash() == t2.Hash() && atomicPointerTransactionDiff.MatchString(diff) {
				return nil
			}
			// fmt.Printf("transaction diff matched (%v): %q\n", atom.MatchString(diff), diff)
		}
		ts1, ok1 := expected.([]*types.Transaction)
		ts2, ok2 := actual.([]*types.Transaction)
		if ok1 && ok2 {
			// Compare hashes of all transactions in the slice.
			if len(ts1) == len(ts2) {
				equal := true
				for i := range ts1 {
					if ts1[i].Hash() != ts2[i].Hash() {
						equal = false
						break
					}
				}
				if equal && atomicPointerTransactionsOrBlockDiff.MatchString(diff) {
					// So if the atomic pointers do not match but the hashes do we consider the blocks equal.
					return nil
				}
				// fmt.Printf("transactionsdiff matched (%v): %q\n", atomicPointerBlockDiff.MatchString(diff), diff)
			}
		}

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
