package chaindb

import (
	"encoding/json"
	"github.com/hashrs/consensusdemo/core"
	"github.com/hashrs/consensusdemo/core/objectpool"
	"github.com/hashrs/consensusdemo/db"
	"github.com/hashrs/consensusdemo/db/memdb"
	"github.com/hashrs/consensusdemo/types"
	log "github.com/sirupsen/logrus"
	"math/big"
	"sync/atomic"
)

type ChainDB interface {
	CurrentHeight() *big.Int
	SaveReceipts(r []*types.Receipt)
	SaveTransactions(tx []*types.FurtherTransaction)
	GetBlock(*big.Int) *core.Block
	GetTransaction(hash types.Hash) *types.FurtherTransaction
	GetReceipt(hash types.Hash) *types.Receipt
	SaveBlock(*core.Block) error
}

func NewChainDB(database db.Database) ChainDB {
	chain := &memChaindb{
		database:       database,
		cache:          memdb.NewMemDB(),
		tosaveBlock:    make(chan *core.Block, 1000000),
		tosaveTxs:      make(chan types.FurtherTransactions, 1000000),
		tosaveReceipts: make(chan types.Receipts, 1000000),
	}
	chain.storeTask()
	return chain
}

type memChaindb struct {
	cache          db.CacheKV
	database       db.Database
	height         atomic.Value
	tosaveReceipts chan types.Receipts
	tosaveTxs      chan types.FurtherTransactions
	tosaveBlock    chan *core.Block
}

func (m *memChaindb) saveHeight(n *big.Int) {
	hstr := n.Text(10)
	k := chainHeightKey()
	m.database.Set(k, []byte(hstr))     // db save string
	m.height.Store(new(big.Int).Set(n)) // cache save *big.Int
}

func (m *memChaindb) storeTask() {

	go func() {
		for {
			select {
			case rs := <-m.tosaveReceipts:
				for _, r := range rs {
					k := receiptKey(r.Txhash)
					d, e := r.Encode()
					if e != nil {
						log.Trace("receipt encode error failed to store", "txhash ", r.Txhash)
						continue
					}
					//log.Debug("save receipt to database", "hash ", k)
					m.database.Set(k, d)
					if pr, exist := m.cache.Get(k); exist {
						objectpool.PutReceiptObject(pr.(*types.Receipt))
						m.cache.Del(k)
					}
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case txs := <-m.tosaveTxs:
				for _, tx := range txs {
					k := transactionKey(tx.Hash())
					d, e := tx.Encode()
					if e != nil {
						log.Trace("transaction encode error failed to store", "txhash ", tx.Hash())
						continue
					}
					m.database.Set(k, d)
					if pr, exist := m.cache.Get(k); exist {
						objectpool.PutTransactionObject(pr.(*types.FurtherTransaction))
						m.cache.Del(k)
					}
				}
			}

		}
	}()

	go func() {
		for {
			select {
			case block := <-m.tosaveBlock:
				k := blockKey(block.Header.Number)
				d, e := block.Encode()
				if e != nil {
					log.Trace("block encode error failed to store", "number ", block.Header.Number)
					continue
				}
				m.database.Set(k, d)
				m.cache.Del(k)
			}
		}
	}()
}

func (m *memChaindb) toStoreReceipts(receipts []*types.Receipt) {
	log.Debug("before send to saveReceipts")
	m.tosaveReceipts <- receipts
	log.Debug("after send to saveReceipts")
}

func (m *memChaindb) toStoreTransactions(txs []*types.FurtherTransaction) {
	log.Debug("before send to saveTxs")
	m.tosaveTxs <- txs
	log.Debug("after send to saveTxs")
}

func (m *memChaindb) toStoreBlocks(block *core.Block) {
	log.Debug("before send to saveBlock")
	m.tosaveBlock <- block
	log.Debug("after send to saveBlock")
}

func (m *memChaindb) CurrentHeight() *big.Int {
	var height = big.NewInt(0)
	if v := m.height.Load(); v != nil {
		height = new(big.Int).Set(v.(*big.Int))
	} else {
		h, exist := m.database.Get(chainHeightKey())
		if exist {
			height, _ = new(big.Int).SetString(string(h), 10)
		}
		m.height.Store(new(big.Int).Set(height))
	}
	return height
}

func (m *memChaindb) GetBlock(num *big.Int) *core.Block {
	k := blockKey(num)
	if block, exist := m.cache.Get(k); exist {
		return block.(*core.Block)
	} else {
		d, exist := m.database.Get(k)
		if !exist {
			return nil
		}
		var block core.Block
		if err := json.Unmarshal(d, &block); err != nil {
			return nil
		}
		return &block
	}
}

func (m *memChaindb) SaveBlock(block *core.Block) error {
	k := blockKey(block.Header.Number)
	m.cache.Set(k, block)
	m.toStoreBlocks(block)
	m.saveHeight(block.Header.Number)
	return nil
}

func (m *memChaindb) SaveTransactions(txs []*types.FurtherTransaction) {
	log.Debug("in save transactions ", " txs number ", len(txs))
	for i:=0; i < len(txs); i++ {
		tx := txs[i]
		k := transactionKey(tx.Hash())
		log.Debug("before set tx to cache ", " i ", i)
		m.cache.Set(k, tx)
		log.Debug("after set tx to cache ", " i ", i)
	}
	m.toStoreTransactions(txs)
}

func (m *memChaindb) SaveReceipts(rs []*types.Receipt) {
	log.Debug("in save receipts ", " txs number ", len(rs))
	for i:=0; i < len(rs); i++ {
		r := rs[i]
		k := receiptKey(r.Txhash)
		m.cache.Set(k, r)
	}
	m.toStoreReceipts(rs)

}

func (m *memChaindb) GetTransaction(hash types.Hash) *types.FurtherTransaction {
	k := transactionKey(hash)
	if tx, exist := m.cache.Get(k); exist {
		return tx.(*types.FurtherTransaction)
	} else {
		d, exist := m.database.Get(k)
		if !exist {
			return nil
		}
		var tx types.FurtherTransaction
		if err := json.Unmarshal(d, &tx); err != nil {
			return nil
		}
		return &tx
	}
}

func (m *memChaindb) GetReceipt(txhash types.Hash) *types.Receipt {
	k := receiptKey(txhash)
	if r, exist := m.cache.Get(k); exist {
		//log.Debug("get receipt in cache", "txhash", txhash.String())
		return r.(*types.Receipt)
	} else {
		d, exist := m.database.Get(k)
		if !exist {
			//log.Debug("get receipt from database failed", "txhash", txhash.String())
			return nil
		}
		//log.Debug("get receipt from database succeed", "txhash", txhash.String())
		var r types.Receipt
		if err := json.Unmarshal(d, &r); err != nil {
			return nil
		}
		return &r
	}
}
