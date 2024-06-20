// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abigen

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// CeloTokenMetaData contains all meta data concerning the CeloToken contract.
var CeloTokenMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"test\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"allowance\",\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"approve\",\"inputs\":[{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"balanceOf\",\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"burn\",\"inputs\":[{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"circulatingSupply\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"decimals\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"decreaseAllowance\",\"inputs\":[{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"depositAmount\",\"inputs\":[{\"name\":\"_depositAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getBurnedAmount\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getVersionNumber\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"goldTokenMintingSchedule\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIMintGoldSchedule\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"increaseAllowance\",\"inputs\":[{\"name\":\"spender\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"increaseSupply\",\"inputs\":[{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"registryAddress\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initialized\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isL2\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isOwner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"l2ToL1MessagePasser\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"mint\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"name\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registry\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIRegistry\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setGoldTokenMintingScheduleAddress\",\"inputs\":[{\"name\":\"goldTokenMintingScheduleAddress\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setL2ToL1MessagePasser\",\"inputs\":[{\"name\":\"_l2ToL1MessagePasser\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setRegistry\",\"inputs\":[{\"name\":\"registryAddress\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"symbol\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"totalSupply\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transfer\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferFrom\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferWithComment\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"comment\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"withdrawAmount\",\"inputs\":[{\"name\":\"_withdrawAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"withdrawn\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"Approval\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"spender\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"RegistrySet\",\"inputs\":[{\"name\":\"registryAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"SetGoldTokenMintingScheduleAddress\",\"inputs\":[{\"name\":\"newScheduleAddress\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Transfer\",\"inputs\":[{\"name\":\"from\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"to\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"value\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TransferComment\",\"inputs\":[{\"name\":\"comment\",\"type\":\"string\",\"indexed\":false,\"internalType\":\"string\"}],\"anonymous\":false}]",
}

// CeloTokenABI is the input ABI used to generate the binding from.
// Deprecated: Use CeloTokenMetaData.ABI instead.
var CeloTokenABI = CeloTokenMetaData.ABI

// CeloToken is an auto generated Go binding around an Ethereum contract.
type CeloToken struct {
	CeloTokenCaller     // Read-only binding to the contract
	CeloTokenTransactor // Write-only binding to the contract
	CeloTokenFilterer   // Log filterer for contract events
}

// CeloTokenCaller is an auto generated read-only Go binding around an Ethereum contract.
type CeloTokenCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CeloTokenTransactor is an auto generated write-only Go binding around an Ethereum contract.
type CeloTokenTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CeloTokenFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type CeloTokenFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// CeloTokenSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type CeloTokenSession struct {
	Contract     *CeloToken        // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// CeloTokenCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type CeloTokenCallerSession struct {
	Contract *CeloTokenCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts    // Call options to use throughout this session
}

// CeloTokenTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type CeloTokenTransactorSession struct {
	Contract     *CeloTokenTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// CeloTokenRaw is an auto generated low-level Go binding around an Ethereum contract.
type CeloTokenRaw struct {
	Contract *CeloToken // Generic contract binding to access the raw methods on
}

// CeloTokenCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type CeloTokenCallerRaw struct {
	Contract *CeloTokenCaller // Generic read-only contract binding to access the raw methods on
}

// CeloTokenTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type CeloTokenTransactorRaw struct {
	Contract *CeloTokenTransactor // Generic write-only contract binding to access the raw methods on
}

// NewCeloToken creates a new instance of CeloToken, bound to a specific deployed contract.
func NewCeloToken(address common.Address, backend bind.ContractBackend) (*CeloToken, error) {
	contract, err := bindCeloToken(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &CeloToken{CeloTokenCaller: CeloTokenCaller{contract: contract}, CeloTokenTransactor: CeloTokenTransactor{contract: contract}, CeloTokenFilterer: CeloTokenFilterer{contract: contract}}, nil
}

// NewCeloTokenCaller creates a new read-only instance of CeloToken, bound to a specific deployed contract.
func NewCeloTokenCaller(address common.Address, caller bind.ContractCaller) (*CeloTokenCaller, error) {
	contract, err := bindCeloToken(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CeloTokenCaller{contract: contract}, nil
}

// NewCeloTokenTransactor creates a new write-only instance of CeloToken, bound to a specific deployed contract.
func NewCeloTokenTransactor(address common.Address, transactor bind.ContractTransactor) (*CeloTokenTransactor, error) {
	contract, err := bindCeloToken(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CeloTokenTransactor{contract: contract}, nil
}

// NewCeloTokenFilterer creates a new log filterer instance of CeloToken, bound to a specific deployed contract.
func NewCeloTokenFilterer(address common.Address, filterer bind.ContractFilterer) (*CeloTokenFilterer, error) {
	contract, err := bindCeloToken(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CeloTokenFilterer{contract: contract}, nil
}

// bindCeloToken binds a generic wrapper to an already deployed contract.
func bindCeloToken(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := CeloTokenMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CeloToken *CeloTokenRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CeloToken.Contract.CeloTokenCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CeloToken *CeloTokenRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CeloToken.Contract.CeloTokenTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CeloToken *CeloTokenRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CeloToken.Contract.CeloTokenTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CeloToken *CeloTokenCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _CeloToken.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_CeloToken *CeloTokenTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CeloToken.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_CeloToken *CeloTokenTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _CeloToken.Contract.contract.Transact(opts, method, params...)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address spender) view returns(uint256)
func (_CeloToken *CeloTokenCaller) Allowance(opts *bind.CallOpts, _owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "allowance", _owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address spender) view returns(uint256)
func (_CeloToken *CeloTokenSession) Allowance(_owner common.Address, spender common.Address) (*big.Int, error) {
	return _CeloToken.Contract.Allowance(&_CeloToken.CallOpts, _owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address _owner, address spender) view returns(uint256)
func (_CeloToken *CeloTokenCallerSession) Allowance(_owner common.Address, spender common.Address) (*big.Int, error) {
	return _CeloToken.Contract.Allowance(&_CeloToken.CallOpts, _owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256)
func (_CeloToken *CeloTokenCaller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "balanceOf", _owner)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256)
func (_CeloToken *CeloTokenSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _CeloToken.Contract.BalanceOf(&_CeloToken.CallOpts, _owner)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _owner) view returns(uint256)
func (_CeloToken *CeloTokenCallerSession) BalanceOf(_owner common.Address) (*big.Int, error) {
	return _CeloToken.Contract.BalanceOf(&_CeloToken.CallOpts, _owner)
}

// CirculatingSupply is a free data retrieval call binding the contract method 0x9358928b.
//
// Solidity: function circulatingSupply() view returns(uint256)
func (_CeloToken *CeloTokenCaller) CirculatingSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "circulatingSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CirculatingSupply is a free data retrieval call binding the contract method 0x9358928b.
//
// Solidity: function circulatingSupply() view returns(uint256)
func (_CeloToken *CeloTokenSession) CirculatingSupply() (*big.Int, error) {
	return _CeloToken.Contract.CirculatingSupply(&_CeloToken.CallOpts)
}

// CirculatingSupply is a free data retrieval call binding the contract method 0x9358928b.
//
// Solidity: function circulatingSupply() view returns(uint256)
func (_CeloToken *CeloTokenCallerSession) CirculatingSupply() (*big.Int, error) {
	return _CeloToken.Contract.CirculatingSupply(&_CeloToken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_CeloToken *CeloTokenCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_CeloToken *CeloTokenSession) Decimals() (uint8, error) {
	return _CeloToken.Contract.Decimals(&_CeloToken.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_CeloToken *CeloTokenCallerSession) Decimals() (uint8, error) {
	return _CeloToken.Contract.Decimals(&_CeloToken.CallOpts)
}

// GetBurnedAmount is a free data retrieval call binding the contract method 0x265126bd.
//
// Solidity: function getBurnedAmount() view returns(uint256)
func (_CeloToken *CeloTokenCaller) GetBurnedAmount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "getBurnedAmount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetBurnedAmount is a free data retrieval call binding the contract method 0x265126bd.
//
// Solidity: function getBurnedAmount() view returns(uint256)
func (_CeloToken *CeloTokenSession) GetBurnedAmount() (*big.Int, error) {
	return _CeloToken.Contract.GetBurnedAmount(&_CeloToken.CallOpts)
}

// GetBurnedAmount is a free data retrieval call binding the contract method 0x265126bd.
//
// Solidity: function getBurnedAmount() view returns(uint256)
func (_CeloToken *CeloTokenCallerSession) GetBurnedAmount() (*big.Int, error) {
	return _CeloToken.Contract.GetBurnedAmount(&_CeloToken.CallOpts)
}

// GetVersionNumber is a free data retrieval call binding the contract method 0x54255be0.
//
// Solidity: function getVersionNumber() pure returns(uint256, uint256, uint256, uint256)
func (_CeloToken *CeloTokenCaller) GetVersionNumber(opts *bind.CallOpts) (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "getVersionNumber")

	if err != nil {
		return *new(*big.Int), *new(*big.Int), *new(*big.Int), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	out2 := *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	out3 := *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return out0, out1, out2, out3, err

}

// GetVersionNumber is a free data retrieval call binding the contract method 0x54255be0.
//
// Solidity: function getVersionNumber() pure returns(uint256, uint256, uint256, uint256)
func (_CeloToken *CeloTokenSession) GetVersionNumber() (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	return _CeloToken.Contract.GetVersionNumber(&_CeloToken.CallOpts)
}

// GetVersionNumber is a free data retrieval call binding the contract method 0x54255be0.
//
// Solidity: function getVersionNumber() pure returns(uint256, uint256, uint256, uint256)
func (_CeloToken *CeloTokenCallerSession) GetVersionNumber() (*big.Int, *big.Int, *big.Int, *big.Int, error) {
	return _CeloToken.Contract.GetVersionNumber(&_CeloToken.CallOpts)
}

// GoldTokenMintingSchedule is a free data retrieval call binding the contract method 0xf5120aaa.
//
// Solidity: function goldTokenMintingSchedule() view returns(address)
func (_CeloToken *CeloTokenCaller) GoldTokenMintingSchedule(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "goldTokenMintingSchedule")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GoldTokenMintingSchedule is a free data retrieval call binding the contract method 0xf5120aaa.
//
// Solidity: function goldTokenMintingSchedule() view returns(address)
func (_CeloToken *CeloTokenSession) GoldTokenMintingSchedule() (common.Address, error) {
	return _CeloToken.Contract.GoldTokenMintingSchedule(&_CeloToken.CallOpts)
}

// GoldTokenMintingSchedule is a free data retrieval call binding the contract method 0xf5120aaa.
//
// Solidity: function goldTokenMintingSchedule() view returns(address)
func (_CeloToken *CeloTokenCallerSession) GoldTokenMintingSchedule() (common.Address, error) {
	return _CeloToken.Contract.GoldTokenMintingSchedule(&_CeloToken.CallOpts)
}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_CeloToken *CeloTokenCaller) Initialized(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "initialized")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_CeloToken *CeloTokenSession) Initialized() (bool, error) {
	return _CeloToken.Contract.Initialized(&_CeloToken.CallOpts)
}

// Initialized is a free data retrieval call binding the contract method 0x158ef93e.
//
// Solidity: function initialized() view returns(bool)
func (_CeloToken *CeloTokenCallerSession) Initialized() (bool, error) {
	return _CeloToken.Contract.Initialized(&_CeloToken.CallOpts)
}

// IsL2 is a free data retrieval call binding the contract method 0x76348f71.
//
// Solidity: function isL2() view returns(bool)
func (_CeloToken *CeloTokenCaller) IsL2(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "isL2")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsL2 is a free data retrieval call binding the contract method 0x76348f71.
//
// Solidity: function isL2() view returns(bool)
func (_CeloToken *CeloTokenSession) IsL2() (bool, error) {
	return _CeloToken.Contract.IsL2(&_CeloToken.CallOpts)
}

// IsL2 is a free data retrieval call binding the contract method 0x76348f71.
//
// Solidity: function isL2() view returns(bool)
func (_CeloToken *CeloTokenCallerSession) IsL2() (bool, error) {
	return _CeloToken.Contract.IsL2(&_CeloToken.CallOpts)
}

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() view returns(bool)
func (_CeloToken *CeloTokenCaller) IsOwner(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "isOwner")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() view returns(bool)
func (_CeloToken *CeloTokenSession) IsOwner() (bool, error) {
	return _CeloToken.Contract.IsOwner(&_CeloToken.CallOpts)
}

// IsOwner is a free data retrieval call binding the contract method 0x8f32d59b.
//
// Solidity: function isOwner() view returns(bool)
func (_CeloToken *CeloTokenCallerSession) IsOwner() (bool, error) {
	return _CeloToken.Contract.IsOwner(&_CeloToken.CallOpts)
}

// L2ToL1MessagePasser is a free data retrieval call binding the contract method 0x85e06ce2.
//
// Solidity: function l2ToL1MessagePasser() view returns(address)
func (_CeloToken *CeloTokenCaller) L2ToL1MessagePasser(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "l2ToL1MessagePasser")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L2ToL1MessagePasser is a free data retrieval call binding the contract method 0x85e06ce2.
//
// Solidity: function l2ToL1MessagePasser() view returns(address)
func (_CeloToken *CeloTokenSession) L2ToL1MessagePasser() (common.Address, error) {
	return _CeloToken.Contract.L2ToL1MessagePasser(&_CeloToken.CallOpts)
}

// L2ToL1MessagePasser is a free data retrieval call binding the contract method 0x85e06ce2.
//
// Solidity: function l2ToL1MessagePasser() view returns(address)
func (_CeloToken *CeloTokenCallerSession) L2ToL1MessagePasser() (common.Address, error) {
	return _CeloToken.Contract.L2ToL1MessagePasser(&_CeloToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_CeloToken *CeloTokenCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_CeloToken *CeloTokenSession) Name() (string, error) {
	return _CeloToken.Contract.Name(&_CeloToken.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_CeloToken *CeloTokenCallerSession) Name() (string, error) {
	return _CeloToken.Contract.Name(&_CeloToken.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CeloToken *CeloTokenCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CeloToken *CeloTokenSession) Owner() (common.Address, error) {
	return _CeloToken.Contract.Owner(&_CeloToken.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_CeloToken *CeloTokenCallerSession) Owner() (common.Address, error) {
	return _CeloToken.Contract.Owner(&_CeloToken.CallOpts)
}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_CeloToken *CeloTokenCaller) Registry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "registry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_CeloToken *CeloTokenSession) Registry() (common.Address, error) {
	return _CeloToken.Contract.Registry(&_CeloToken.CallOpts)
}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_CeloToken *CeloTokenCallerSession) Registry() (common.Address, error) {
	return _CeloToken.Contract.Registry(&_CeloToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_CeloToken *CeloTokenCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_CeloToken *CeloTokenSession) Symbol() (string, error) {
	return _CeloToken.Contract.Symbol(&_CeloToken.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_CeloToken *CeloTokenCallerSession) Symbol() (string, error) {
	return _CeloToken.Contract.Symbol(&_CeloToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_CeloToken *CeloTokenCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_CeloToken *CeloTokenSession) TotalSupply() (*big.Int, error) {
	return _CeloToken.Contract.TotalSupply(&_CeloToken.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_CeloToken *CeloTokenCallerSession) TotalSupply() (*big.Int, error) {
	return _CeloToken.Contract.TotalSupply(&_CeloToken.CallOpts)
}

// Withdrawn is a free data retrieval call binding the contract method 0xc80ec522.
//
// Solidity: function withdrawn() view returns(uint256)
func (_CeloToken *CeloTokenCaller) Withdrawn(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _CeloToken.contract.Call(opts, &out, "withdrawn")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Withdrawn is a free data retrieval call binding the contract method 0xc80ec522.
//
// Solidity: function withdrawn() view returns(uint256)
func (_CeloToken *CeloTokenSession) Withdrawn() (*big.Int, error) {
	return _CeloToken.Contract.Withdrawn(&_CeloToken.CallOpts)
}

// Withdrawn is a free data retrieval call binding the contract method 0xc80ec522.
//
// Solidity: function withdrawn() view returns(uint256)
func (_CeloToken *CeloTokenCallerSession) Withdrawn() (*big.Int, error) {
	return _CeloToken.Contract.Withdrawn(&_CeloToken.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_CeloToken *CeloTokenTransactor) Approve(opts *bind.TransactOpts, spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "approve", spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_CeloToken *CeloTokenSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.Approve(&_CeloToken.TransactOpts, spender, value)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address spender, uint256 value) returns(bool)
func (_CeloToken *CeloTokenTransactorSession) Approve(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.Approve(&_CeloToken.TransactOpts, spender, value)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 value) returns(bool)
func (_CeloToken *CeloTokenTransactor) Burn(opts *bind.TransactOpts, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "burn", value)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 value) returns(bool)
func (_CeloToken *CeloTokenSession) Burn(value *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.Burn(&_CeloToken.TransactOpts, value)
}

// Burn is a paid mutator transaction binding the contract method 0x42966c68.
//
// Solidity: function burn(uint256 value) returns(bool)
func (_CeloToken *CeloTokenTransactorSession) Burn(value *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.Burn(&_CeloToken.TransactOpts, value)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 value) returns(bool)
func (_CeloToken *CeloTokenTransactor) DecreaseAllowance(opts *bind.TransactOpts, spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "decreaseAllowance", spender, value)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 value) returns(bool)
func (_CeloToken *CeloTokenSession) DecreaseAllowance(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.DecreaseAllowance(&_CeloToken.TransactOpts, spender, value)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address spender, uint256 value) returns(bool)
func (_CeloToken *CeloTokenTransactorSession) DecreaseAllowance(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.DecreaseAllowance(&_CeloToken.TransactOpts, spender, value)
}

// DepositAmount is a paid mutator transaction binding the contract method 0x87f8ab26.
//
// Solidity: function depositAmount(uint256 _depositAmount) returns()
func (_CeloToken *CeloTokenTransactor) DepositAmount(opts *bind.TransactOpts, _depositAmount *big.Int) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "depositAmount", _depositAmount)
}

// DepositAmount is a paid mutator transaction binding the contract method 0x87f8ab26.
//
// Solidity: function depositAmount(uint256 _depositAmount) returns()
func (_CeloToken *CeloTokenSession) DepositAmount(_depositAmount *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.DepositAmount(&_CeloToken.TransactOpts, _depositAmount)
}

// DepositAmount is a paid mutator transaction binding the contract method 0x87f8ab26.
//
// Solidity: function depositAmount(uint256 _depositAmount) returns()
func (_CeloToken *CeloTokenTransactorSession) DepositAmount(_depositAmount *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.DepositAmount(&_CeloToken.TransactOpts, _depositAmount)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 value) returns(bool)
func (_CeloToken *CeloTokenTransactor) IncreaseAllowance(opts *bind.TransactOpts, spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "increaseAllowance", spender, value)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 value) returns(bool)
func (_CeloToken *CeloTokenSession) IncreaseAllowance(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.IncreaseAllowance(&_CeloToken.TransactOpts, spender, value)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address spender, uint256 value) returns(bool)
func (_CeloToken *CeloTokenTransactorSession) IncreaseAllowance(spender common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.IncreaseAllowance(&_CeloToken.TransactOpts, spender, value)
}

// IncreaseSupply is a paid mutator transaction binding the contract method 0xb921e163.
//
// Solidity: function increaseSupply(uint256 amount) returns()
func (_CeloToken *CeloTokenTransactor) IncreaseSupply(opts *bind.TransactOpts, amount *big.Int) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "increaseSupply", amount)
}

// IncreaseSupply is a paid mutator transaction binding the contract method 0xb921e163.
//
// Solidity: function increaseSupply(uint256 amount) returns()
func (_CeloToken *CeloTokenSession) IncreaseSupply(amount *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.IncreaseSupply(&_CeloToken.TransactOpts, amount)
}

// IncreaseSupply is a paid mutator transaction binding the contract method 0xb921e163.
//
// Solidity: function increaseSupply(uint256 amount) returns()
func (_CeloToken *CeloTokenTransactorSession) IncreaseSupply(amount *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.IncreaseSupply(&_CeloToken.TransactOpts, amount)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address registryAddress) returns()
func (_CeloToken *CeloTokenTransactor) Initialize(opts *bind.TransactOpts, registryAddress common.Address) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "initialize", registryAddress)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address registryAddress) returns()
func (_CeloToken *CeloTokenSession) Initialize(registryAddress common.Address) (*types.Transaction, error) {
	return _CeloToken.Contract.Initialize(&_CeloToken.TransactOpts, registryAddress)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address registryAddress) returns()
func (_CeloToken *CeloTokenTransactorSession) Initialize(registryAddress common.Address) (*types.Transaction, error) {
	return _CeloToken.Contract.Initialize(&_CeloToken.TransactOpts, registryAddress)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 value) returns(bool)
func (_CeloToken *CeloTokenTransactor) Mint(opts *bind.TransactOpts, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "mint", to, value)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 value) returns(bool)
func (_CeloToken *CeloTokenSession) Mint(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.Mint(&_CeloToken.TransactOpts, to, value)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address to, uint256 value) returns(bool)
func (_CeloToken *CeloTokenTransactorSession) Mint(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.Mint(&_CeloToken.TransactOpts, to, value)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CeloToken *CeloTokenTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CeloToken *CeloTokenSession) RenounceOwnership() (*types.Transaction, error) {
	return _CeloToken.Contract.RenounceOwnership(&_CeloToken.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_CeloToken *CeloTokenTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _CeloToken.Contract.RenounceOwnership(&_CeloToken.TransactOpts)
}

// SetGoldTokenMintingScheduleAddress is a paid mutator transaction binding the contract method 0xe584d7f6.
//
// Solidity: function setGoldTokenMintingScheduleAddress(address goldTokenMintingScheduleAddress) returns()
func (_CeloToken *CeloTokenTransactor) SetGoldTokenMintingScheduleAddress(opts *bind.TransactOpts, goldTokenMintingScheduleAddress common.Address) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "setGoldTokenMintingScheduleAddress", goldTokenMintingScheduleAddress)
}

// SetGoldTokenMintingScheduleAddress is a paid mutator transaction binding the contract method 0xe584d7f6.
//
// Solidity: function setGoldTokenMintingScheduleAddress(address goldTokenMintingScheduleAddress) returns()
func (_CeloToken *CeloTokenSession) SetGoldTokenMintingScheduleAddress(goldTokenMintingScheduleAddress common.Address) (*types.Transaction, error) {
	return _CeloToken.Contract.SetGoldTokenMintingScheduleAddress(&_CeloToken.TransactOpts, goldTokenMintingScheduleAddress)
}

// SetGoldTokenMintingScheduleAddress is a paid mutator transaction binding the contract method 0xe584d7f6.
//
// Solidity: function setGoldTokenMintingScheduleAddress(address goldTokenMintingScheduleAddress) returns()
func (_CeloToken *CeloTokenTransactorSession) SetGoldTokenMintingScheduleAddress(goldTokenMintingScheduleAddress common.Address) (*types.Transaction, error) {
	return _CeloToken.Contract.SetGoldTokenMintingScheduleAddress(&_CeloToken.TransactOpts, goldTokenMintingScheduleAddress)
}

// SetL2ToL1MessagePasser is a paid mutator transaction binding the contract method 0x98b8c242.
//
// Solidity: function setL2ToL1MessagePasser(address _l2ToL1MessagePasser) returns()
func (_CeloToken *CeloTokenTransactor) SetL2ToL1MessagePasser(opts *bind.TransactOpts, _l2ToL1MessagePasser common.Address) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "setL2ToL1MessagePasser", _l2ToL1MessagePasser)
}

// SetL2ToL1MessagePasser is a paid mutator transaction binding the contract method 0x98b8c242.
//
// Solidity: function setL2ToL1MessagePasser(address _l2ToL1MessagePasser) returns()
func (_CeloToken *CeloTokenSession) SetL2ToL1MessagePasser(_l2ToL1MessagePasser common.Address) (*types.Transaction, error) {
	return _CeloToken.Contract.SetL2ToL1MessagePasser(&_CeloToken.TransactOpts, _l2ToL1MessagePasser)
}

// SetL2ToL1MessagePasser is a paid mutator transaction binding the contract method 0x98b8c242.
//
// Solidity: function setL2ToL1MessagePasser(address _l2ToL1MessagePasser) returns()
func (_CeloToken *CeloTokenTransactorSession) SetL2ToL1MessagePasser(_l2ToL1MessagePasser common.Address) (*types.Transaction, error) {
	return _CeloToken.Contract.SetL2ToL1MessagePasser(&_CeloToken.TransactOpts, _l2ToL1MessagePasser)
}

// SetRegistry is a paid mutator transaction binding the contract method 0xa91ee0dc.
//
// Solidity: function setRegistry(address registryAddress) returns()
func (_CeloToken *CeloTokenTransactor) SetRegistry(opts *bind.TransactOpts, registryAddress common.Address) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "setRegistry", registryAddress)
}

// SetRegistry is a paid mutator transaction binding the contract method 0xa91ee0dc.
//
// Solidity: function setRegistry(address registryAddress) returns()
func (_CeloToken *CeloTokenSession) SetRegistry(registryAddress common.Address) (*types.Transaction, error) {
	return _CeloToken.Contract.SetRegistry(&_CeloToken.TransactOpts, registryAddress)
}

// SetRegistry is a paid mutator transaction binding the contract method 0xa91ee0dc.
//
// Solidity: function setRegistry(address registryAddress) returns()
func (_CeloToken *CeloTokenTransactorSession) SetRegistry(registryAddress common.Address) (*types.Transaction, error) {
	return _CeloToken.Contract.SetRegistry(&_CeloToken.TransactOpts, registryAddress)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_CeloToken *CeloTokenTransactor) Transfer(opts *bind.TransactOpts, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "transfer", to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_CeloToken *CeloTokenSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.Transfer(&_CeloToken.TransactOpts, to, value)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address to, uint256 value) returns(bool)
func (_CeloToken *CeloTokenTransactorSession) Transfer(to common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.Transfer(&_CeloToken.TransactOpts, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_CeloToken *CeloTokenTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "transferFrom", from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_CeloToken *CeloTokenSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.TransferFrom(&_CeloToken.TransactOpts, from, to, value)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 value) returns(bool)
func (_CeloToken *CeloTokenTransactorSession) TransferFrom(from common.Address, to common.Address, value *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.TransferFrom(&_CeloToken.TransactOpts, from, to, value)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CeloToken *CeloTokenTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CeloToken *CeloTokenSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _CeloToken.Contract.TransferOwnership(&_CeloToken.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_CeloToken *CeloTokenTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _CeloToken.Contract.TransferOwnership(&_CeloToken.TransactOpts, newOwner)
}

// TransferWithComment is a paid mutator transaction binding the contract method 0xe1d6aceb.
//
// Solidity: function transferWithComment(address to, uint256 value, string comment) returns(bool)
func (_CeloToken *CeloTokenTransactor) TransferWithComment(opts *bind.TransactOpts, to common.Address, value *big.Int, comment string) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "transferWithComment", to, value, comment)
}

// TransferWithComment is a paid mutator transaction binding the contract method 0xe1d6aceb.
//
// Solidity: function transferWithComment(address to, uint256 value, string comment) returns(bool)
func (_CeloToken *CeloTokenSession) TransferWithComment(to common.Address, value *big.Int, comment string) (*types.Transaction, error) {
	return _CeloToken.Contract.TransferWithComment(&_CeloToken.TransactOpts, to, value, comment)
}

// TransferWithComment is a paid mutator transaction binding the contract method 0xe1d6aceb.
//
// Solidity: function transferWithComment(address to, uint256 value, string comment) returns(bool)
func (_CeloToken *CeloTokenTransactorSession) TransferWithComment(to common.Address, value *big.Int, comment string) (*types.Transaction, error) {
	return _CeloToken.Contract.TransferWithComment(&_CeloToken.TransactOpts, to, value, comment)
}

// WithdrawAmount is a paid mutator transaction binding the contract method 0x0562b9f7.
//
// Solidity: function withdrawAmount(uint256 _withdrawAmount) returns()
func (_CeloToken *CeloTokenTransactor) WithdrawAmount(opts *bind.TransactOpts, _withdrawAmount *big.Int) (*types.Transaction, error) {
	return _CeloToken.contract.Transact(opts, "withdrawAmount", _withdrawAmount)
}

// WithdrawAmount is a paid mutator transaction binding the contract method 0x0562b9f7.
//
// Solidity: function withdrawAmount(uint256 _withdrawAmount) returns()
func (_CeloToken *CeloTokenSession) WithdrawAmount(_withdrawAmount *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.WithdrawAmount(&_CeloToken.TransactOpts, _withdrawAmount)
}

// WithdrawAmount is a paid mutator transaction binding the contract method 0x0562b9f7.
//
// Solidity: function withdrawAmount(uint256 _withdrawAmount) returns()
func (_CeloToken *CeloTokenTransactorSession) WithdrawAmount(_withdrawAmount *big.Int) (*types.Transaction, error) {
	return _CeloToken.Contract.WithdrawAmount(&_CeloToken.TransactOpts, _withdrawAmount)
}

// CeloTokenApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the CeloToken contract.
type CeloTokenApprovalIterator struct {
	Event *CeloTokenApproval // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CeloTokenApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CeloTokenApproval)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CeloTokenApproval)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CeloTokenApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CeloTokenApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CeloTokenApproval represents a Approval event raised by the CeloToken contract.
type CeloTokenApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_CeloToken *CeloTokenFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*CeloTokenApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _CeloToken.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &CeloTokenApprovalIterator{contract: _CeloToken.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_CeloToken *CeloTokenFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *CeloTokenApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _CeloToken.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CeloTokenApproval)
				if err := _CeloToken.contract.UnpackLog(event, "Approval", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_CeloToken *CeloTokenFilterer) ParseApproval(log types.Log) (*CeloTokenApproval, error) {
	event := new(CeloTokenApproval)
	if err := _CeloToken.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CeloTokenOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the CeloToken contract.
type CeloTokenOwnershipTransferredIterator struct {
	Event *CeloTokenOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CeloTokenOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CeloTokenOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CeloTokenOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CeloTokenOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CeloTokenOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CeloTokenOwnershipTransferred represents a OwnershipTransferred event raised by the CeloToken contract.
type CeloTokenOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_CeloToken *CeloTokenFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*CeloTokenOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _CeloToken.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &CeloTokenOwnershipTransferredIterator{contract: _CeloToken.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_CeloToken *CeloTokenFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *CeloTokenOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _CeloToken.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CeloTokenOwnershipTransferred)
				if err := _CeloToken.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_CeloToken *CeloTokenFilterer) ParseOwnershipTransferred(log types.Log) (*CeloTokenOwnershipTransferred, error) {
	event := new(CeloTokenOwnershipTransferred)
	if err := _CeloToken.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CeloTokenRegistrySetIterator is returned from FilterRegistrySet and is used to iterate over the raw logs and unpacked data for RegistrySet events raised by the CeloToken contract.
type CeloTokenRegistrySetIterator struct {
	Event *CeloTokenRegistrySet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CeloTokenRegistrySetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CeloTokenRegistrySet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CeloTokenRegistrySet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CeloTokenRegistrySetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CeloTokenRegistrySetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CeloTokenRegistrySet represents a RegistrySet event raised by the CeloToken contract.
type CeloTokenRegistrySet struct {
	RegistryAddress common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterRegistrySet is a free log retrieval operation binding the contract event 0x27fe5f0c1c3b1ed427cc63d0f05759ffdecf9aec9e18d31ef366fc8a6cb5dc3b.
//
// Solidity: event RegistrySet(address indexed registryAddress)
func (_CeloToken *CeloTokenFilterer) FilterRegistrySet(opts *bind.FilterOpts, registryAddress []common.Address) (*CeloTokenRegistrySetIterator, error) {

	var registryAddressRule []interface{}
	for _, registryAddressItem := range registryAddress {
		registryAddressRule = append(registryAddressRule, registryAddressItem)
	}

	logs, sub, err := _CeloToken.contract.FilterLogs(opts, "RegistrySet", registryAddressRule)
	if err != nil {
		return nil, err
	}
	return &CeloTokenRegistrySetIterator{contract: _CeloToken.contract, event: "RegistrySet", logs: logs, sub: sub}, nil
}

// WatchRegistrySet is a free log subscription operation binding the contract event 0x27fe5f0c1c3b1ed427cc63d0f05759ffdecf9aec9e18d31ef366fc8a6cb5dc3b.
//
// Solidity: event RegistrySet(address indexed registryAddress)
func (_CeloToken *CeloTokenFilterer) WatchRegistrySet(opts *bind.WatchOpts, sink chan<- *CeloTokenRegistrySet, registryAddress []common.Address) (event.Subscription, error) {

	var registryAddressRule []interface{}
	for _, registryAddressItem := range registryAddress {
		registryAddressRule = append(registryAddressRule, registryAddressItem)
	}

	logs, sub, err := _CeloToken.contract.WatchLogs(opts, "RegistrySet", registryAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CeloTokenRegistrySet)
				if err := _CeloToken.contract.UnpackLog(event, "RegistrySet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRegistrySet is a log parse operation binding the contract event 0x27fe5f0c1c3b1ed427cc63d0f05759ffdecf9aec9e18d31ef366fc8a6cb5dc3b.
//
// Solidity: event RegistrySet(address indexed registryAddress)
func (_CeloToken *CeloTokenFilterer) ParseRegistrySet(log types.Log) (*CeloTokenRegistrySet, error) {
	event := new(CeloTokenRegistrySet)
	if err := _CeloToken.contract.UnpackLog(event, "RegistrySet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CeloTokenSetGoldTokenMintingScheduleAddressIterator is returned from FilterSetGoldTokenMintingScheduleAddress and is used to iterate over the raw logs and unpacked data for SetGoldTokenMintingScheduleAddress events raised by the CeloToken contract.
type CeloTokenSetGoldTokenMintingScheduleAddressIterator struct {
	Event *CeloTokenSetGoldTokenMintingScheduleAddress // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CeloTokenSetGoldTokenMintingScheduleAddressIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CeloTokenSetGoldTokenMintingScheduleAddress)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CeloTokenSetGoldTokenMintingScheduleAddress)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CeloTokenSetGoldTokenMintingScheduleAddressIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CeloTokenSetGoldTokenMintingScheduleAddressIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CeloTokenSetGoldTokenMintingScheduleAddress represents a SetGoldTokenMintingScheduleAddress event raised by the CeloToken contract.
type CeloTokenSetGoldTokenMintingScheduleAddress struct {
	NewScheduleAddress common.Address
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterSetGoldTokenMintingScheduleAddress is a free log retrieval operation binding the contract event 0x63b9a42c1de0047cec2b83aff9244d020bc3ccdb5e1a569a12e85a9e32895792.
//
// Solidity: event SetGoldTokenMintingScheduleAddress(address indexed newScheduleAddress)
func (_CeloToken *CeloTokenFilterer) FilterSetGoldTokenMintingScheduleAddress(opts *bind.FilterOpts, newScheduleAddress []common.Address) (*CeloTokenSetGoldTokenMintingScheduleAddressIterator, error) {

	var newScheduleAddressRule []interface{}
	for _, newScheduleAddressItem := range newScheduleAddress {
		newScheduleAddressRule = append(newScheduleAddressRule, newScheduleAddressItem)
	}

	logs, sub, err := _CeloToken.contract.FilterLogs(opts, "SetGoldTokenMintingScheduleAddress", newScheduleAddressRule)
	if err != nil {
		return nil, err
	}
	return &CeloTokenSetGoldTokenMintingScheduleAddressIterator{contract: _CeloToken.contract, event: "SetGoldTokenMintingScheduleAddress", logs: logs, sub: sub}, nil
}

// WatchSetGoldTokenMintingScheduleAddress is a free log subscription operation binding the contract event 0x63b9a42c1de0047cec2b83aff9244d020bc3ccdb5e1a569a12e85a9e32895792.
//
// Solidity: event SetGoldTokenMintingScheduleAddress(address indexed newScheduleAddress)
func (_CeloToken *CeloTokenFilterer) WatchSetGoldTokenMintingScheduleAddress(opts *bind.WatchOpts, sink chan<- *CeloTokenSetGoldTokenMintingScheduleAddress, newScheduleAddress []common.Address) (event.Subscription, error) {

	var newScheduleAddressRule []interface{}
	for _, newScheduleAddressItem := range newScheduleAddress {
		newScheduleAddressRule = append(newScheduleAddressRule, newScheduleAddressItem)
	}

	logs, sub, err := _CeloToken.contract.WatchLogs(opts, "SetGoldTokenMintingScheduleAddress", newScheduleAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CeloTokenSetGoldTokenMintingScheduleAddress)
				if err := _CeloToken.contract.UnpackLog(event, "SetGoldTokenMintingScheduleAddress", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSetGoldTokenMintingScheduleAddress is a log parse operation binding the contract event 0x63b9a42c1de0047cec2b83aff9244d020bc3ccdb5e1a569a12e85a9e32895792.
//
// Solidity: event SetGoldTokenMintingScheduleAddress(address indexed newScheduleAddress)
func (_CeloToken *CeloTokenFilterer) ParseSetGoldTokenMintingScheduleAddress(log types.Log) (*CeloTokenSetGoldTokenMintingScheduleAddress, error) {
	event := new(CeloTokenSetGoldTokenMintingScheduleAddress)
	if err := _CeloToken.contract.UnpackLog(event, "SetGoldTokenMintingScheduleAddress", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CeloTokenTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the CeloToken contract.
type CeloTokenTransferIterator struct {
	Event *CeloTokenTransfer // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CeloTokenTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CeloTokenTransfer)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CeloTokenTransfer)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CeloTokenTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CeloTokenTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CeloTokenTransfer represents a Transfer event raised by the CeloToken contract.
type CeloTokenTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_CeloToken *CeloTokenFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*CeloTokenTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _CeloToken.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &CeloTokenTransferIterator{contract: _CeloToken.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_CeloToken *CeloTokenFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *CeloTokenTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _CeloToken.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CeloTokenTransfer)
				if err := _CeloToken.contract.UnpackLog(event, "Transfer", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_CeloToken *CeloTokenFilterer) ParseTransfer(log types.Log) (*CeloTokenTransfer, error) {
	event := new(CeloTokenTransfer)
	if err := _CeloToken.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// CeloTokenTransferCommentIterator is returned from FilterTransferComment and is used to iterate over the raw logs and unpacked data for TransferComment events raised by the CeloToken contract.
type CeloTokenTransferCommentIterator struct {
	Event *CeloTokenTransferComment // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *CeloTokenTransferCommentIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(CeloTokenTransferComment)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(CeloTokenTransferComment)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *CeloTokenTransferCommentIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *CeloTokenTransferCommentIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// CeloTokenTransferComment represents a TransferComment event raised by the CeloToken contract.
type CeloTokenTransferComment struct {
	Comment string
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransferComment is a free log retrieval operation binding the contract event 0xe5d4e30fb8364e57bc4d662a07d0cf36f4c34552004c4c3624620a2c1d1c03dc.
//
// Solidity: event TransferComment(string comment)
func (_CeloToken *CeloTokenFilterer) FilterTransferComment(opts *bind.FilterOpts) (*CeloTokenTransferCommentIterator, error) {

	logs, sub, err := _CeloToken.contract.FilterLogs(opts, "TransferComment")
	if err != nil {
		return nil, err
	}
	return &CeloTokenTransferCommentIterator{contract: _CeloToken.contract, event: "TransferComment", logs: logs, sub: sub}, nil
}

// WatchTransferComment is a free log subscription operation binding the contract event 0xe5d4e30fb8364e57bc4d662a07d0cf36f4c34552004c4c3624620a2c1d1c03dc.
//
// Solidity: event TransferComment(string comment)
func (_CeloToken *CeloTokenFilterer) WatchTransferComment(opts *bind.WatchOpts, sink chan<- *CeloTokenTransferComment) (event.Subscription, error) {

	logs, sub, err := _CeloToken.contract.WatchLogs(opts, "TransferComment")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(CeloTokenTransferComment)
				if err := _CeloToken.contract.UnpackLog(event, "TransferComment", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransferComment is a log parse operation binding the contract event 0xe5d4e30fb8364e57bc4d662a07d0cf36f4c34552004c4c3624620a2c1d1c03dc.
//
// Solidity: event TransferComment(string comment)
func (_CeloToken *CeloTokenFilterer) ParseTransferComment(log types.Log) (*CeloTokenTransferComment, error) {
	event := new(CeloTokenTransferComment)
	if err := _CeloToken.contract.UnpackLog(event, "TransferComment", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
