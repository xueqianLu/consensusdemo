package consensus

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashrs/consensusdemo/core"
	"github.com/hashrs/consensusdemo/core/db"
	"golang.org/x/crypto/sha3"
	"math/big"
)

var (
	big0 = big.NewInt(0)
	big1 = big.NewInt(1)
)

type Engine interface {
	CheckMiner() bool
	MakeBlock(txs []*core.FurtherTransaction) *core.Block
}

func NewEngine() Engine {

	return &dummyEngine{
		chaindb:  db.NewChainDB(),
		globaldb: db.NewGlobalDB(),
	}
}

type dummyEngine struct {
	chaindb  db.ChainDB
	globaldb db.GlobalDB
}

func (c *dummyEngine) CheckMiner() bool {
	return true
}

func (c *dummyEngine) MakeBlock(txs []*core.FurtherTransaction) *core.Block {
	cur := c.chaindb.CurrentHeight()
	last := c.chaindb.GetBlock(cur)
	parent := common.Hash{}
	if last != nil {
		parent = last.Header.Hash
	}

	c.exec(txs)

	block := &core.Block{
		Header: core.BlockHeader{
			Parent: parent,
			Number: new(big.Int).Add(cur, big1),
		},
		Body: core.BlockBody{
			Txs: txs,
		},
	}
	block.Header.Hash = sha3.Sum256(parent.Bytes())
	return block
}

func (c *dummyEngine) exec(txs []*core.FurtherTransaction) error {
	for _, tx := range txs {
		if tx.Value().Cmp(big0) != 0 {
			c.globaldb.SubBalance(core.Account(tx.From), tx.Value())
			c.globaldb.AddBalance(core.Account(*tx.To()), tx.Value())
		}
	}
	return nil
}
