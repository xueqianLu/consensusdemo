package objectpool

import (
	"github.com/hashrs/consensusdemo/types"
	//log "github.com/sirupsen///logrus"
	"math/big"
	"sync"
)

var transactionPool = sync.Pool{
	New: func() interface{} {
		return new(types.FurtherTransaction)
	},
}

func GetTransactionObject() *types.FurtherTransaction {
	return new(types.FurtherTransaction)
	ptr := transactionPool.Get().(*types.FurtherTransaction)
	//log.Debugf("get tx object ptr %p \n", ptr)
	return ptr
}

func PutTransactionObject(tx *types.FurtherTransaction) {
	return
	//log.Debugf("put tx object ptr %p \n", tx)
	transactionPool.Put(tx)
}

var receiptPool = sync.Pool{
	New: func() interface{} {
		return new(types.Receipt)
	},
}

func GetReceiptObject() *types.Receipt {
	ptr := receiptPool.Get().(*types.Receipt)
	//log.Debugf("get receipt object ptr %p \n", ptr)
	return ptr

}

func PutReceiptObject(tx *types.Receipt) {
	//log.Debugf("put receipt object ptr %p \n", tx)
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
