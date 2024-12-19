package rawdb

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	preGingerbreadFieldsPrefix = []byte("preGingerbreadFields-")
)

type PreGingerbreadAdditionalFields struct {
	BaseFee  *big.Int `rlp:"nil"`
	GasLimit *big.Int `rlp:"nil"`
}

func preGingerbreadAdditionalFieldsKey(hash common.Hash) []byte {
	return append(preGingerbreadFieldsPrefix, hash.Bytes()...)
}

func ReadPreGingerbreadAdditionalFields(db ethdb.KeyValueReader, blockHash common.Hash) (*PreGingerbreadAdditionalFields, error) {
	data, _ := db.Get(preGingerbreadAdditionalFieldsKey(blockHash))
	if len(data) == 0 {
		return nil, nil
	}

	fields := &PreGingerbreadAdditionalFields{}

	err := rlp.DecodeBytes(data, fields)
	if err != nil {
		return nil, err
	}

	return fields, nil
}

func WritePreGingerbreadAdditionalFields(db ethdb.KeyValueWriter, blockHash common.Hash, data *PreGingerbreadAdditionalFields) error {
	rawData, err := rlp.EncodeToBytes(data)
	if err != nil {
		return err
	}

	if err := db.Put(preGingerbreadAdditionalFieldsKey(blockHash), rawData); err != nil {
		return err
	}

	return nil
}
