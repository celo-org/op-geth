package rawdb

import (
	"fmt"
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

	// Make sure it returns an error for a nonexistent record
	record0, err := ReadPreGingerbreadBlockBaseFee(db, hash)
	assert.ErrorContains(t, err, fmt.Sprintf("error retrieving pre gingerbread base fee for block: %s, error: not found", hash.String()))
	require.Nil(t, record0)

	// Write data
	err = WritePreGingerbreadBlockBaseFee(db, hash, value)
	require.NoError(t, err)

	// Read data
	record, err := ReadPreGingerbreadBlockBaseFee(db, hash)
	require.NoError(t, err)
	assert.Equal(t, value, record)
}
