package objectpool

import (
	"github.com/hashrs/consensusdemo/types"
	"math/big"
	"sync"
)

var transactionPool = sync.Pool{
	New: func() interface{} {
		return new(types.FurtherTransaction)
	},
}

func GetTransactionObject() *types.FurtherTransaction {
	return transactionPool.Get().(*types.FurtherTransaction)
}

func PutTransactionObject(tx *types.FurtherTransaction) {
	transactionPool.Put(tx)
}

var receiptPool = sync.Pool{
	New: func() interface{} {
		return new(types.Receipt)
	},
}

func GetReceiptObject() *types.Receipt {
	return receiptPool.Get().(*types.Receipt)
}

func PutReceiptObject(tx *types.Receipt) {
	receiptPool.Put(tx)
}

var bigintPool = sync.Pool{
	New: func() interface{} {
		return new(big.Int)
	},
}

func GetBigIntObject() *big.Int {
	return bigintPool.Get().(*big.Int)
}

func PutBigIntObject(d *big.Int) {
	bigintPool.Put(d)
}
