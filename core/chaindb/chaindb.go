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
	"time"
)

type ChainDB interface {
	CurrentHeight() *big.Int
	SaveReceipts(r []*types.Receipt)
	SaveTransactions(tx []*types.FurtherTransaction)
	GetBlock(*big.Int) *core.Block
	GetBlockHeader(*big.Int) *core.BlockHeader
	GetTransaction(hash types.Hash) *types.FurtherTransaction
	GetReceipt(hash types.Hash) *types.Receipt
	SaveBlock(*core.Block) error
}

var (
	big10 = big.NewInt(10)
	big1  = big.NewInt(1)
)

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
	startHeight    *big.Int
}

func (m *memChaindb) saveHeight(n *big.Int) {
	if m.startHeight == nil {
		m.startHeight = new(big.Int).Set(n)
	}
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
				batch := m.database.NewBatch()
				delk := make([]string, 0, len(rs))
				for _, r := range rs {
					k := receiptKey(r.Txhash)
					d, e := r.Encode()
					if e != nil {
						log.Trace("receipt encode error failed to store", "txhash ", r.Txhash)
						continue
					}
					//log.Debug("save receipt to database", "hash ", k)
					batch.Set(k, d)
					delk = append(delk, k)
				}
				batch.Write()
				for _, k := range delk {
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
				batch := m.database.NewBatch()
				delk := make([]string, 0, len(txs))
				for _, tx := range txs {
					k := transactionKey(tx.Hash())
					d, e := tx.Encode()
					if e != nil {
						log.Trace("transaction encode error failed to store", "txhash ", tx.Hash())
						continue
					}
					batch.Set(k, d)
					delk = append(delk, k)
				}
				batch.Write()
				for _, k := range txs {
					if pr, exist := m.cache.Get(k); exist {
						objectpool.PutTransactionObject(pr.(*types.FurtherTransaction))
						m.cache.Del(k)
					}
				}
			}
		}
	}()

	go func() {
		duration := time.Second * 10
		tm := time.NewTicker(duration)
		defer tm.Stop()

		for {
			select {
			case block := <-m.tosaveBlock:
				{
					hk := blockHeaderKey(block.Header.Number)
					dh, e := block.Header.Encode()
					if e != nil {
						log.Trace("block header encode error failed to store", "number ", block.Header.Number)
						continue
					}
					m.database.Set(hk, dh)
				}
				{
					bodyk := blockBodyKey(block.Header.Number)
					dbody, e := block.Body.Encode()
					if e != nil {
						log.Trace("block body encode error failed to store", "number ", block.Header.Number)
						continue
					}
					m.database.Set(bodyk, dbody)
					m.cache.Del(bodyk)
				}
			case <-tm.C:
				if m.startHeight != nil {
					h := m.CurrentHeight()
					for h.Cmp(m.startHeight.Add(m.startHeight, big10)) > 0 {
						hk := blockHeaderKey(m.startHeight)
						m.cache.Del(hk)
						m.startHeight = new(big.Int).Add(m.startHeight, big1)
					}
				}
				tm.Reset(duration)
			}
		}
	}()
}

func (m *memChaindb) toStoreReceipts(receipts []*types.Receipt) {
	m.tosaveReceipts <- receipts
}

func (m *memChaindb) toStoreTransactions(txs []*types.FurtherTransaction) {
	m.tosaveTxs <- txs
}

func (m *memChaindb) toStoreBlocks(block *core.Block) {
	m.tosaveBlock <- block
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
	h := m.GetBlockHeader(num)
	if h == nil {
		return nil
	}
	b := m.GetBlockBody(num)
	if b == nil {
		return nil
	}
	return &core.Block{Header: h, Body: b}
}

func (m *memChaindb) GetBlockBody(num *big.Int) *core.BlockBody {
	k := blockBodyKey(num)
	if body, exist := m.cache.Get(k); exist {
		return body.(*core.BlockBody)
	} else {
		d, exist := m.database.Get(k)
		if !exist {
			return nil
		}
		var body core.BlockBody
		if err := json.Unmarshal(d, &body); err != nil {
			return nil
		}
		return &body
	}
}

func (m *memChaindb) GetBlockHeader(num *big.Int) *core.BlockHeader {
	k := blockHeaderKey(num)
	if header, exist := m.cache.Get(k); exist {
		return header.(*core.BlockHeader)
	} else {
		d, exist := m.database.Get(k)
		if !exist {
			return nil
		}
		var header core.BlockHeader
		if err := json.Unmarshal(d, &header); err != nil {
			return nil
		}
		return &header
	}
}

func (m *memChaindb) cacheBlockHeader(block *core.Block) error {
	k := blockHeaderKey(block.Header.Number)
	m.cache.Set(k, block.Header)
	return nil
}

func (m *memChaindb) cacheBlockBody(block *core.Block) error {
	k := blockBodyKey(block.Header.Number)
	m.cache.Set(k, block.Body)
	return nil
}

func (m *memChaindb) SaveBlock(block *core.Block) error {
	m.cacheBlockHeader(block)
	m.cacheBlockBody(block)
	m.toStoreBlocks(block)
	m.saveHeight(block.Header.Number)
	return nil
}

func (m *memChaindb) SaveTransactions(txs []*types.FurtherTransaction) {
	log.Debug("in save transactions ", " txs number ", len(txs))
	length := len(txs)
	for i := 0; i < length; i++ {
		k := transactionKey(txs[i].Hash())
		m.cache.Set(k, txs[i])
	}
	m.toStoreTransactions(txs)
}

func (m *memChaindb) SaveReceipts(rs []*types.Receipt) {
	log.Debug("in save receipts ", " txs number ", len(rs))
	for i := 0; i < len(rs); i++ {
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
