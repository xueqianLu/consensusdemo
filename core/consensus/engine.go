package consensus

import (
	"github.com/hashrs/consensusdemo/core"
	"github.com/hashrs/consensusdemo/core/chaindb"
	"github.com/hashrs/consensusdemo/core/globaldb"
	"github.com/hashrs/consensusdemo/lib"
	"github.com/hashrs/consensusdemo/types"
	"math/big"
)

var (
	big0 = big.NewInt(0)
	big1 = big.NewInt(1)
)

type Engine interface {
	CheckMiner() bool
	MakeBlock(header *core.BlockHeader, txs []*types.FurtherTransaction) *core.Block
	ExecBlock(block *core.Block) []*types.Receipt
}

func NewEngine(globaldb globaldb.GlobalDB, chaindb chaindb.ChainDB) Engine {

	return &dummyEngine{
		chaindb:  chaindb,
		globaldb: globaldb,
	}
}

type dummyEngine struct {
	chaindb  chaindb.ChainDB
	globaldb globaldb.GlobalDB
}

func (c *dummyEngine) CheckMiner() bool {
	return true
}

func (c *dummyEngine) MakeBlock(header *core.BlockHeader, txs []*types.FurtherTransaction) *core.Block {
	cur := c.chaindb.CurrentHeight()
	last := c.chaindb.GetBlock(cur)
	parent := types.Hash{}

	if last != nil {
		parent = last.Hash()
	}
	header.Number = new(big.Int).Add(cur, big1)
	header.Parent = parent

	block := &core.Block{
		Header: *header,
		Body: core.BlockBody{
			Txs: txs,
		},
	}
	receipts := c.exec(block)

	block.Header.ReceiptRoot = lib.HashSlices(types.Receipts(receipts))
	block.Header.TxRoot = lib.HashSlices(types.FurtherTransactions(txs))

	return block
}

func (c *dummyEngine) exec(block *core.Block) []*types.Receipt {
	var receipts = make([]*types.Receipt, 0)
	for _, tx := range block.Body.Txs {
		r := &types.Receipt{
			Txhash:      tx.Hash(),
			From:        tx.From,
			To:          *tx.To(),
			Timestamp:   int64(block.Header.Timestamp),
			Value:       new(big.Int).Set(tx.Value()),
			BlockNumber: new(big.Int).Set(block.Header.Number),
		}
		if c.CanTransfer(core.Account{tx.From}, core.Account{*tx.To()}, tx.Value()) {
			c.Transfer(core.Account{tx.From}, core.Account{*tx.To()}, tx.Value())
			r.Status = 1
		} else {
			r.Status = 0
		}
		receipts = append(receipts, r)
	}
	return receipts
}

func (c *dummyEngine) ExecBlock(block *core.Block) []*types.Receipt {
	return c.exec(block)
}

func (c *dummyEngine) CanTransfer(from core.Account, to core.Account, value *big.Int) bool {
	balance := c.globaldb.GetBalance(from)
	return value.Cmp(big0) == 0 || balance.Cmp(value) > 0
}

func (c *dummyEngine) Transfer(from core.Account, to core.Account, value *big.Int) error {
	c.globaldb.SubBalance(from, value)
	c.globaldb.AddBalance(to, value)
	return nil
}
