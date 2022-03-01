package db

import (
	"github.com/hashrs/consensusdemo/core"
	log "github.com/sirupsen/logrus"
	"math/big"
	"sync"
)

type ChainDB interface {
	CurrentHeight() *big.Int
	GetBlock(*big.Int) *core.Block
	SaveBlock(*core.Block) error
}

func NewChainDB() ChainDB {
	return &memChaindb{
		height: big.NewInt(0),
	}
}

type memChaindb struct {
	state  sync.Map
	height *big.Int
}

func (m *memChaindb) CurrentHeight() *big.Int {
	return m.height
}

func (m *memChaindb) GetBlock(num *big.Int) *core.Block {
	if block, exist := m.state.Load(num.Uint64()); exist {
		return block.(*core.Block)
	} else {
		return nil
	}
}
func (m *memChaindb) SaveBlock(block *core.Block) error {
	m.state.Store(block.Header.Number.Uint64(), block)
	m.height = new(big.Int).Set(block.Header.Number)
	log.Println("save block set height to ", "h ", m.height)
	return nil
}
