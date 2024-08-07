package contracts

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// ReadOnlyStateDB wraps a StateDB to remove access list tracking.
// Used together with an EVMInterpreter in readOnly mode (e.g. StaticCall), all
// changes to the StateDB are prevented.
//
// Gas calculations based on ReadOnlyStateDB will be wrong because the accessed
// storage slots and addresses are not tracked.
type ReadOnlyStateDB struct {
	vm.StateDB
}

func (r *ReadOnlyStateDB) AddressInAccessList(addr common.Address) bool {
	return false
}

func (r *ReadOnlyStateDB) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	return false, false
}

func (r *ReadOnlyStateDB) AddAddressToAccessList(addr common.Address) {
}

func (r *ReadOnlyStateDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {
}
