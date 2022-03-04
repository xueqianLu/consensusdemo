package handler

import (
	"github.com/hashrs/consensusdemo/core"
	"github.com/hashrs/consensusdemo/types"
	"math/big"
)

type ChainReader interface {
	GetBlock(*big.Int) *core.Block
	GetTransaction(hash types.Hash) *types.FurtherTransaction
	GetReceipt(hash types.Hash) *types.Receipt
}

type GlobalReader interface {
	GetBalance(addr core.Account) *big.Int
}
