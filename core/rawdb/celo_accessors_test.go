package rawdb

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadAndWritePreGingerbreadBlockBaseFee tests reading and writing pre-gingerbread block base fee to database
func TestReadAndWritePreGingerbreadBlockBaseFee(t *testing.T) {
	db := NewMemoryDatabase()

	hash := common.HexToHash("0x1")
	value := big.NewInt(1234)

	// Make sure that the data is not found
	record0, err := ReadPreGingerbreadBlockBaseFee(db, hash)
	require.NoError(t, err)
	require.Nil(t, record0)

	// Write data
	err = WritePreGingerbreadBlockBaseFee(db, hash, value)
	require.NoError(t, err)

	// Read data
	record, err := ReadPreGingerbreadBlockBaseFee(db, hash)
	require.NoError(t, err)
	assert.Equal(t, value, record)
}
