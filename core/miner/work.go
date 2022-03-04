package miner

import (
	"github.com/hashrs/consensusdemo/config"
	"github.com/hashrs/consensusdemo/core"
	"github.com/hashrs/consensusdemo/core/chaindb"
	"github.com/hashrs/consensusdemo/core/consensus"
	"github.com/hashrs/consensusdemo/core/globaldb"
	"github.com/hashrs/consensusdemo/globalclock"
	"github.com/hashrs/consensusdemo/txpool"
	"github.com/hashrs/consensusdemo/types"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type newPrepareBlock struct {
	timestamp uint64
	packedtxs []*types.FurtherTransaction
}

type Miner struct {
	lentry    *log.Entry
	engine    consensus.Engine
	chain     chaindb.ChainDB
	packageCh chan newPrepareBlock
	control   chan struct{}
	round     chan types.RoundInfo
	clock     globalclock.Clock
	wg        sync.WaitGroup
	txp       *txpool.TxPool
}

func NewMiner(globaldb globaldb.GlobalDB, chaindb chaindb.ChainDB, txp *txpool.TxPool, conf *config.Config) *Miner {
	m := &Miner{
		lentry:    log.WithField("miner", 1),
		engine:    consensus.NewEngine(globaldb, chaindb),
		chain:     chaindb,
		packageCh: make(chan newPrepareBlock, 10000),
		control:   make(chan struct{}),
		clock:     globalclock.NewClock(conf),
		txp:       txp,
	}

	return m
}

func (m *Miner) roundloop() {
	var maxPackTxs = 100000
	m.lentry.Info("worker round loop started")
	for {
		select {
		case <-m.control:
			m.lentry.Info("worker round loop exited.")
			return
		case roundInfo, ok := <-m.round:
			if !ok {
				return
			}
			m.lentry.Info("worker got new round,", " time=", roundInfo.Timestamp, " tx count=", len(roundInfo.Txsinfo))
			// new round to gen block.
			committxs := make([]*types.FurtherTransaction, 0, maxPackTxs)
			packedtxs := roundInfo.Txsinfo
			for _, packed := range packedtxs {
				hashes := packed.Hashs()
				for _, hash := range hashes {
					txs := m.txp.GetTxs(hash)
					for len(txs) == 0 {
						// if tx is not get in txpool, wait forever.
						// todo: implement timeout.
						time.Sleep(time.Millisecond * 10)
						txs = m.txp.GetTxs(hash)
					}

					committxs = append(committxs, txs...)
					if len(committxs) >= maxPackTxs {
						m.packageCh <- newPrepareBlock{roundInfo.Timestamp, committxs}
						committxs = make([]*types.FurtherTransaction, 0, maxPackTxs)
					}
				}
			}
			m.packageCh <- newPrepareBlock{roundInfo.Timestamp, committxs}
		}
	}
}

func (m *Miner) genBlock() {
	m.lentry.Info("worker gen block loop started")
	for {
		select {
		case <-m.control:
			m.lentry.Info("worker gen block loop exited")
			break
		case prepared, ok := <-m.packageCh:
			if !ok {
				return
			}
			header := &core.BlockHeader{
				Timestamp: prepared.timestamp,
			}
			block := m.engine.MakeBlock(header, prepared.packedtxs)
			m.chain.SaveBlock(block)

			txs := block.Body.Txs
			for _, tx := range txs {
				m.chain.SaveTransaction(tx)
			}
			receipts := m.engine.ExecBlock(block)
			for _, r := range receipts {
				m.chain.SaveReceipt(r)
			}

			m.lentry.Info("mined new block ", " number ", block.Header.Number.Uint64(), " txs ", len(block.Body.Txs))
		}
	}
}

func (m *Miner) Start() {
	m.round = m.clock.WatchClock()
	go m.clock.Start()
	m.wg = sync.WaitGroup{}
	m.wg.Add(2)
	close(m.control)
	m.control = make(chan struct{})
	go m.roundloop()
	go m.genBlock()
	m.lentry.Info("miner started")
}

func (m *Miner) Stop() {
	m.clock.Stop()
	close(m.control)
	m.wg.Wait()
}
