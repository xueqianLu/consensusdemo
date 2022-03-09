package chaindb

import (
	"encoding/json"
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
	h := n.Text(10)
	k := chainHeightKey()
	m.database.Set(k, h) // db save string base 10

	m.height.Store(new(big.Int).Set(n)) // cache save *big.Int
}

func (m *memChaindb) CurrentHeight() *big.Int {
	var height = big.NewInt(0)
	if v := m.height.Load(); v != nil {
		height = new(big.Int).Set(v.(*big.Int))
	} else {
		h, exist := m.database.Get(chainHeightKey())
		if exist {
			height, _ = new(big.Int).SetString(h.(string), 10)
		}
		m.height.Store(new(big.Int).Set(height))
	}
	return height
}

func (m *memChaindb) GetBlock(num *big.Int) *core.Block {
	k := blockKey(num)
	if data, exist := m.database.Get(k); exist {
		var block core.Block
		if err := json.Unmarshal([]byte(data.(string)), &block); err != nil {
			return nil
		}

		return &block
	} else {
		return nil
	}
}

func (m *memChaindb) SaveBlock(block *core.Block) error {
	k := blockKey(block.Header.Number)
	v, _ := json.Marshal(block)
	m.database.Set(k, string(v))
	m.saveHeight(block.Header.Number)
	return nil
}

func (m *memChaindb) SaveTransaction(tx *types.FurtherTransaction) {
	k := transactionKey(tx.Hash())
	v, _ := json.Marshal(tx)
	m.database.Set(k, string(v))
}

func (m *memChaindb) SaveReceipt(r *types.Receipt) {
	k := receiptKey(r.Txhash)
	v, _ := json.Marshal(r)
	m.database.Set(k, string(v))
}

func (m *memChaindb) GetTransaction(hash types.Hash) *types.FurtherTransaction {
	k := transactionKey(hash)
	if data, exist := m.database.Get(k); exist {
		var tx types.FurtherTransaction
		if err := json.Unmarshal([]byte(data.(string)), &tx); err != nil {
			return nil
		}
		return &tx
	} else {
		return nil
	}

}
func (m *memChaindb) GetReceipt(hash types.Hash) *types.Receipt {
	k := receiptKey(hash)
	if data, exist := m.database.Get(k); exist {
		var r types.Receipt
		if err := json.Unmarshal([]byte(data.(string)), &r); err != nil {
			return nil
		}

		return &r
	} else {
		return nil
	}
}
