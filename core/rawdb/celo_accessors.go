package rawdb

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
)

var (
	CeloPreGingerbreadBaseFeePrefix = []byte("celoPgbBlockBaseFee-") // CeloPreGingerbreadBlockBaseFeePrefix + block hash -> BaseFee
)

type PreGingerbreadFields struct {
	BaseFee  *big.Int
	GasLimit *big.Int
}

// preGingerbreadBlockBaseFeeKey calculates a database key of pre-Gingerbread block BaseFee for the given block hash
func preGingerbreadBlockBaseFeeKey(hash common.Hash) []byte {
	return append(CeloPreGingerbreadBaseFeePrefix, hash[:]...)
}

// ReadPreGingerbreadBlockBaseFee reads BaseFee of pre-Gingerbread block from the given database for the given block hash
func ReadPreGingerbreadBlockBaseFee(db ethdb.KeyValueReader, blockHash common.Hash) (*big.Int, error) {
	data, _ := db.Get(preGingerbreadBlockBaseFeeKey(blockHash))
	if len(data) == 0 {
		return nil, nil
	}

	return new(big.Int).SetBytes(data), nil
}

// WritePreGingerbreadBlockBaseFee writes BaseFee of pre-Gingerbread block to the given database at the given block hash
func WritePreGingerbreadBlockBaseFee(db ethdb.KeyValueWriter, blockHash common.Hash, baseFee *big.Int) error {
	return db.Put(preGingerbreadBlockBaseFeeKey(blockHash), baseFee.Bytes())
}
