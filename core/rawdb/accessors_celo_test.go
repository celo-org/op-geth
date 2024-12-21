package rawdb

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
)

func TestPreGingerbreadAdditionalFields(t *testing.T) {
	db := NewMemoryDatabase()

	mockData, err := rlp.EncodeToBytes(&PreGingerbreadAdditionalFields{
		BaseFee:  big.NewInt(1000),
		GasLimit: big.NewInt(2000),
	})
	assert.NoError(t, err)

	type SeedData struct {
		hash common.Hash
		data []byte
	}

	tests := []struct {
		name         string
		seedData     []SeedData
		hash         common.Hash
		expectedRes  *PreGingerbreadAdditionalFields
		returnsError bool
	}{
		{
			name:         "should return nil if data is not found",
			seedData:     []SeedData{},
			hash:         common.HexToHash("0x1"),
			expectedRes:  nil,
			returnsError: false,
		},
		{
			name: "should return data",
			seedData: []SeedData{
				{
					hash: common.HexToHash("0x2"),
					data: mockData,
				},
			},
			hash: common.HexToHash("0x2"),
			expectedRes: &PreGingerbreadAdditionalFields{
				BaseFee:  big.NewInt(1000),
				GasLimit: big.NewInt(2000),
			},
			returnsError: false,
		},
		{
			name: "should return error if data is broken",
			seedData: []SeedData{
				{
					hash: common.HexToHash("0x3"),
					data: []byte{0x1, 0x2, 0x3},
				},
			},
			hash:         common.HexToHash("0x3"),
			expectedRes:  nil,
			returnsError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, seed := range test.seedData {
				err := db.Put(preGingerbreadAdditionalFieldsKey(seed.hash), seed.data)
				assert.NoError(t, err)
			}

			t.Cleanup(func() {
				for _, seed := range test.seedData {
					err := db.Delete(preGingerbreadAdditionalFieldsKey(seed.hash))
					assert.NoError(t, err)
				}
			})

			res, err := ReadPreGingerbreadAdditionalFields(db, test.hash)
			assert.Equal(t, test.expectedRes, res)

			if test.returnsError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
