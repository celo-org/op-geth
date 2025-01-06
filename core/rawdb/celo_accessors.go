package rawdb

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	CeloPreGingerbreadFieldsPrefix = []byte("celoPgbFields-") // CeloPreGingerbreadFieldsPrefix + block hash -> PreGingerbreadAdditionalFields
)

type PreGingerbreadFields struct {
	BaseFee  *big.Int
	GasLimit *big.Int
}

// preGingerbreadAdditionalFieldsKey calculates a database key for PreGingerbreadAdditionalFields for the given block hash
func preGingerbreadAdditionalFieldsKey(hash common.Hash) []byte {
	return append(CeloPreGingerbreadFieldsPrefix, hash[:]...)
}

// ReadPreGingerbreadAdditionalFields reads PreGingerbreadAdditionalFields from the given database for the given block hash
func ReadPreGingerbreadAdditionalFields(db ethdb.KeyValueReader, blockHash common.Hash) (*PreGingerbreadFields, error) {
	data, _ := db.Get(preGingerbreadAdditionalFieldsKey(blockHash))
	if len(data) == 0 {
		return nil, nil
	}

	fields := &PreGingerbreadFields{}

	err := rlp.DecodeBytes(data, fields)
	if err != nil {
		return nil, err
	}

	return fields, nil
}

// WritePreGingerbreadAdditionalFields writes PreGingerbreadAdditionalFields to the given database for the given block hash
func WritePreGingerbreadAdditionalFields(db ethdb.KeyValueWriter, blockHash common.Hash, data *PreGingerbreadFields) error {
	rawData, err := rlp.EncodeToBytes(data)
	if err != nil {
		return err
	}

	if err := db.Put(preGingerbreadAdditionalFieldsKey(blockHash), rawData); err != nil {
		return err
	}

	return nil
}
