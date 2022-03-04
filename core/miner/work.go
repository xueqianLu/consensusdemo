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
	round     chan interface{}
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
		round:     make(chan interface{}, 10000),
		control:   make(chan struct{}),
		clock:     globalclock.NewClock(conf),
		txp:       txp,
	}
	return m
}

func (m *Miner) roundloop() {
	var maxPackTxs = 100000
	for {
		select {
		case <-m.control:
			m.lentry.Info("worker round loop exited.")
			return
		case newround, ok := <-m.round:
			if !ok {
				return
			}
			// new round to gen block.
			roundInfo := newround.(types.RoundInfo)
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
		}
	}
}

func (m *Miner) genBlock() {
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
			m.lentry.Info("mined new block ", " number ", block.Header.Number.Uint64(), " txs ", len(block.Body.Txs))
		}
	}
}

func (m *Miner) Start() {
	m.wg = sync.WaitGroup{}
	m.wg.Add(2)
	close(m.control)
	m.control = make(chan struct{})
	go m.roundloop()
	go m.genBlock()
}

func (m *Miner) Stop() {
	close(m.control)
	m.wg.Wait()
}
