package ethapi

import (
	"errors"
	"math/big"
	"testing"

	"github.com/status-im/keycard-go/hexutils"
	"github.com/stretchr/testify/assert"
)

func Test_parseGasPriceMinimumUpdated(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		result *big.Int
		err    error
	}{
		{
			name:   "should parse gas price successfully",
			data:   hexutils.HexToBytes("00000000000000000000000000000000000000000000000000000000000f4240"),
			result: big.NewInt(1_000_000),
			err:    nil,
		},
		{
			name:   "should return error if data is not in the expected format",
			data:   hexutils.HexToBytes("123456"),
			result: nil,
			err:    errors.New("abi: cannot marshal in to go type: length insufficient 3 require 32"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parseGasPriceMinimumUpdated(test.data)
			assert.Equal(t, test.result, result)

			if test.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.err.Error())
			}
		})
	}
}
