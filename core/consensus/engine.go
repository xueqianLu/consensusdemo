package consensus

import (
	"github.com/hashrs/consensusdemo/core"
	"github.com/hashrs/consensusdemo/core/chaindb"
	"github.com/hashrs/consensusdemo/core/globaldb"
	"github.com/hashrs/consensusdemo/core/objectpool"

	//"github.com/hashrs/consensusdemo/lib"
	"github.com/hashrs/consensusdemo/types"
	log "github.com/sirupsen/logrus"
	"math/big"
	"time"
)

var (
	big0 = big.NewInt(0)
	big1 = big.NewInt(1)
)

type Engine interface {
	CheckMiner() bool
	MakeBlock(header *core.BlockHeader, txs []*types.FurtherTransaction) (*core.Block, []*types.Receipt)
	ExecBlock(block *core.Block) []*types.Receipt
	Commit() error
}

func NewEngine(globaldb globaldb.GlobalDB, chaindb chaindb.ChainDB) Engine {

	return &dummyEngine{
		chaindb:  chaindb,
		globaldb: globaldb,
		lentry:   log.WithField("consensus", "dummy"),
	}
}

type dummyEngine struct {
	chaindb  chaindb.ChainDB
	globaldb globaldb.GlobalDB
	lentry   *log.Entry
}

func (c *dummyEngine) CheckMiner() bool {
	return true
}

func (c *dummyEngine) MakeBlock(header *core.BlockHeader, txs []*types.FurtherTransaction) (*core.Block, []*types.Receipt) {
	cur := c.chaindb.CurrentHeight()
	t1 := time.Now()
	last := c.chaindb.GetBlockHeader(cur)
	t2 := time.Now()
	log.Info("worker make block get block", "cost ", t2.Sub(t1).Milliseconds())
	parent := types.Hash{}

	if last != nil {
		parent = last.BlockHash
	}

	header.Number = new(big.Int).Add(cur, big1)
	header.Parent = parent

	block := &core.Block{
		Header: header,
		Body: &core.BlockBody{
			Txs: txs,
		},
	}
	block.Header.BlockHash = block.Hash()

	t3 := time.Now()
	log.Info("worker make block make header", "cost ", t3.Sub(t2).Milliseconds())
	receipts := c.exec(block)
	t4 := time.Now()
	log.Info("worker make block exec tx", "cost ", t4.Sub(t3).Milliseconds())
	block.Header.ReceiptRoot = types.Hash{} //lib.HashSlices(types.Receipts(receipts))
	block.Header.TxRoot = types.Hash{}      //lib.HashSlices(types.FurtherTransactions(txs))

	return block, receipts
}

func (c *dummyEngine) exec(block *core.Block) []*types.Receipt {
	var receipts = make([]*types.Receipt, len(block.Body.Txs))
	for i := 0; i < len(block.Body.Txs); i++ {
		tx := block.Body.Txs[i]
		pr := objectpool.GetReceiptObject()
		pr.Txhash = tx.Hash()
		pr.From = tx.From
		pr.To = *tx.To()
		pr.Index = int64(i)
		pr.PackedTime = int64(block.Header.Timestamp)
		pr.ExecTime = time.Now().Unix()
		pr.Value = new(big.Int).Set(tx.Value())
		pr.BlockNumber = new(big.Int).Set(block.Header.Number)
		if c.CanTransfer(core.Account{tx.From}, core.Account{*tx.To()}, tx.Value()) {
			c.Transfer(core.Account{tx.From}, core.Account{*tx.To()}, tx.Value())
			pr.Status = 1
		} else {
			pr.Status = 0
		}
		receipts[i] = pr
	}
	return receipts
}

func (c *dummyEngine) ExecBlock(block *core.Block) []*types.Receipt {
	return c.exec(block)
}

func (c *dummyEngine) CanTransfer(from core.Account, to core.Account, value *big.Int) bool {
	balance := c.globaldb.GetBalance(from)
	return value.Cmp(big0) == 0 || balance.Cmp(value) >= 0
}

func (c *dummyEngine) Transfer(from core.Account, to core.Account, value *big.Int) error {
	c.globaldb.SubBalance(from, value)
	c.globaldb.AddBalance(to, value)
	return nil
}

func (c *dummyEngine) Commit() error {
	return c.globaldb.Commit()
}
