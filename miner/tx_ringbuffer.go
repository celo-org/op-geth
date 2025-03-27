package miner

import "github.com/ethereum/go-ethereum/common"

// TxRingBuffer is a ring buffer of transaction hashes, it has one method to add
// transaction hashes and another method to check if a transaction hash is in the buffer.
type TxRingBuffer struct {
	ring  []common.Hash
	m     map[common.Hash]struct{}
	index int
}

// NewTxRingBuffer creates a new transaction ring buffer with the given size.
func NewTxRingBuffer(size int) *TxRingBuffer {
	return &TxRingBuffer{
		ring: make([]common.Hash, size),
		m:    make(map[common.Hash]struct{}),
	}
}

// Add adds a new transaction hash to the ring buffer, if the buffer is
// at capacity, the oldest transaction hash is overwritten.
func (b *TxRingBuffer) Add(tx common.Hash) {
	// Clean up old transaction hash from map
	oldHash := b.ring[b.index]
	delete(b.m, oldHash)

	// Add new transaction hash to ring and map
	b.ring[b.index] = tx
	b.m[tx] = struct{}{}

	// Increment the index
	b.index++
	if b.index == len(b.ring) {
		b.index = 0
	}
}

func (b *TxRingBuffer) Has(tx common.Hash) bool {
	_, exists := b.m[tx]
	return exists
}
