package vm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/types"
)

// ChainContext supports retrieving headers and consensus parameters from the
// current blockchain to be used during transaction processing.
type ChainContext interface {
	// Engine retrieves the chain's consensus engine.
	Engine() consensus.Engine

	// GetHeader returns the hash corresponding to their hash.
	GetHeader(common.Hash, uint64) *types.Header
}

// NewEVMBlockContext creates a new context for use in the EVM.
func NewEVMBlockContext(number uint64, timestamp uint64, db StateDB) BlockContext {
	// If we don't have an explicit author (i.e. not mining), extract from the header
	// var beneficiary common.Address
	// if author == nil {
	// 	beneficiary, _ = chain.Engine().Author(header) // Ignore error, we're past header validation
	// } else {
	// 	beneficiary = *author
	// }
	return BlockContext{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     GetHashFn(db),
		Coinbase:    common.HexToAddress("0x0000000000000000000000000000000000000000"),
		BlockNumber: new(big.Int).SetUint64(number),
		Time:        new(big.Int).SetUint64(timestamp),
		Difficulty:  big.NewInt(0x2000),
		GasLimit:    0x2fefd8,
	}
}

// NewEVMTxContext creates a new transaction context for a single transaction.
func NewEVMTxContext(msg types.Message) TxContext {
	return TxContext{
		Origin:   msg.From(),
		GasPrice: new(big.Int).Set(msg.GasPrice()),
	}
}

// GetHashFn returns a GetHashFunc which retrieves header hashes by number
func GetHashFn(db StateDB) func(n uint64) common.Hash {
	return func(n uint64) common.Hash {
		return db.GetBlockEVMHash(n)
	}
}

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db StateDB, addr common.Address, amount *big.Int) bool {
	return db.GetEVMBalance(addr).Cmp(amount) >= 0
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db StateDB, sender, recipient common.Address, amount *big.Int) {
	db.SubEVMBalance(sender, amount)
	db.AddEVMBalance(recipient, amount)
}
