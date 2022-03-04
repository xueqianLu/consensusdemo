package chaindb

import (
	"github.com/hashrs/consensusdemo/core"
	"github.com/hashrs/consensusdemo/db"
	"github.com/hashrs/consensusdemo/types"
	"math/big"
	"sync/atomic"
)

type ChainDB interface {
	CurrentHeight() *big.Int
	SaveReceipt(r *types.Receipt)
	SaveTransaction(tx *types.FurtherTransaction)
	GetBlock(*big.Int) *core.Block
	GetTransaction(hash types.Hash) *types.FurtherTransaction
	GetReceipt(hash types.Hash) *types.Receipt
	SaveBlock(*core.Block) error
}

func NewChainDB(database db.Database) ChainDB {
	return &memChaindb{
		database: database,
	}
}

type memChaindb struct {
	database db.Database
	height   atomic.Value
}

func (m *memChaindb) saveHeight(n *big.Int) {
	h := n.Int64()
	k := chainHeightKey()
	m.database.Set(k, h)                // db save int64
	m.height.Store(new(big.Int).Set(n)) // cache save *big.Int
}

func (m *memChaindb) CurrentHeight() *big.Int {
	var height = big.NewInt(0)
	if v := m.height.Load(); v != nil {
		height = new(big.Int).Set(v.(*big.Int))
	} else {
		h, exist := m.database.Get(chainHeightKey())
		if exist {
			height = big.NewInt(h.(int64))
		}
		m.height.Store(new(big.Int).Set(height))
	}
	return height
}

func (m *memChaindb) GetBlock(num *big.Int) *core.Block {
	k := blockKey(num)
	if block, exist := m.database.Get(k); exist {
		return block.(*core.Block)
	} else {
		return nil
	}
}

func (m *memChaindb) SaveBlock(block *core.Block) error {
	k := blockKey(block.Header.Number)
	m.database.Set(k, block)
	m.saveHeight(block.Header.Number)
	return nil
}

func (m *memChaindb) SaveTransaction(tx *types.FurtherTransaction) {
	k := transactionKey(tx.Hash())
	m.database.Set(k, tx)
}

func (m *memChaindb) SaveReceipt(r *types.Receipt) {
	k := receiptKey(r.Txhash)
	m.database.Set(k, r)
}

func (m *memChaindb) GetTransaction(hash types.Hash) *types.FurtherTransaction {
	k := transactionKey(hash)
	if tx, exist := m.database.Get(k); exist {
		return tx.(*types.FurtherTransaction)
	} else {
		return nil
	}

}
func (m *memChaindb) GetReceipt(hash types.Hash) *types.Receipt {
	k := receiptKey(hash)
	if r, exist := m.database.Get(k); exist {
		return r.(*types.Receipt)
	} else {
		return nil
	}
}
