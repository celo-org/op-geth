package rawdb

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
)

// TestPreGingerbreadAdditionalFields tests ReadPreGingerbreadAdditionalFields function
func TestPreGingerbreadAdditionalFields(t *testing.T) {
	db := NewMemoryDatabase()

	mockData, err := rlp.EncodeToBytes(&PreGingerbreadFields{
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
		expectedRes  *PreGingerbreadFields
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
			expectedRes: &PreGingerbreadFields{
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

// TestWritePreGingerbreadAdditionalFields tests WritePreGingerbreadAdditionalFields function
func TestWritePreGingerbreadAdditionalFields(t *testing.T) {
	db := NewMemoryDatabase()

	hash := common.HexToHash("0x1")
	data := []*PreGingerbreadFields{
		{
			BaseFee:  big.NewInt(0),
			GasLimit: big.NewInt(2000),
		},
		{
			BaseFee:  big.NewInt(3000),
			GasLimit: big.NewInt(0),
		},
		{
			BaseFee:  big.NewInt(5000),
			GasLimit: big.NewInt(6000),
		},
	}

	// Make sure that the data is not found
	record0, err := ReadPreGingerbreadAdditionalFields(db, hash)
	assert.NoError(t, err)
	assert.Nil(t, record0)

	for _, d := range data {
		// Write data
		err := WritePreGingerbreadAdditionalFields(db, hash, d)
		assert.NoError(t, err)

		// Read data
		record, err := ReadPreGingerbreadAdditionalFields(db, hash)
		assert.NoError(t, err)
		assert.Equal(t, d, record)
	}
}
