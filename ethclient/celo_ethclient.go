package ethclient

import (
	"github.com/ethereum/go-ethereum"
)

// Converts FilterQuery into format compatible with eth rpc
func ToFilterArg(q ethereum.FilterQuery) (interface{}, error) {
	return toFilterArg(q)
}
