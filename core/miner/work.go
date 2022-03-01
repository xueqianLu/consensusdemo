package miner

import (
	"github.com/hashrs/consensusdemo/core"
	"github.com/hashrs/consensusdemo/core/consensus"
	"github.com/hashrs/consensusdemo/core/db"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Miner struct {
	lentry    *log.Entry
	engine    consensus.Engine
	chaindb   db.ChainDB
	txpoolCh  chan []*core.FurtherTransaction
	packageCh chan []*core.FurtherTransaction
	control   chan struct{}
	wg        sync.WaitGroup
}

func NewMiner(globaldb db.GlobalDB, chaindb db.ChainDB) *Miner {
	return &Miner{
		lentry:    log.WithField("miner", 1),
		engine:    consensus.NewEngine(globaldb, chaindb),
		chaindb:   chaindb,
		txpoolCh:  make(chan []*core.FurtherTransaction, 1000000),
		packageCh: make(chan []*core.FurtherTransaction, 10000),
		control:   make(chan struct{}),
	}
}

func (m *Miner) txPoolLoop() {
	var maxPackTxs = 100000
	var packtxs = make([]*core.FurtherTransaction, 0, maxPackTxs)
	tm := time.NewTicker(time.Second)
	defer tm.Stop()

	for {
		select {
		case <-m.control:
			m.lentry.Info("txpool loop exited")
			break
		case txs, ok := <-m.txpoolCh:
			if !ok {
				return
			}
			packtxs = append(packtxs, txs...)
			if len(packtxs) >= maxPackTxs {
				m.lentry.Info("package tx finished.", "total txs ", len(packtxs))
				m.packageCh <- packtxs
				packtxs = make([]*core.FurtherTransaction, 0, maxPackTxs)
			}
		case <-tm.C:
			m.lentry.Info("package tx with timeout.", "total txs ", len(packtxs))
			m.packageCh <- packtxs
			packtxs = make([]*core.FurtherTransaction, 0, maxPackTxs)
		}
	}
}

func (m *Miner) genBlock() {
	for {
		select {
		case <-m.control:
			m.lentry.Info("genblock loop exited")
			break
		case packtx, ok := <-m.packageCh:
			if !ok {
				return
			}
			block := m.engine.MakeBlock(packtx)
			m.chaindb.SaveBlock(block)
			m.lentry.Info("mined new block ", " number ", block.Header.Number.Uint64(), " txs ", len(block.Body.Txs))
		}
	}
}

func (m *Miner) TxPool() chan []*core.FurtherTransaction {
	return m.txpoolCh
}

func (m *Miner) Start() {
	m.wg = sync.WaitGroup{}
	m.wg.Add(2)
	close(m.control)
	m.control = make(chan struct{})
	go m.txPoolLoop()
	go m.genBlock()
}

func (m *Miner) Stop() {
	close(m.control)
	m.wg.Wait()
}
