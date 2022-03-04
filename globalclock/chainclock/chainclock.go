package chainclock

import (
	"bytes"
	"context"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashrs/consensusdemo/types"
	log "github.com/sirupsen/logrus"
)

type ChainClock struct {
	logger   *log.Entry
	name     string
	reporter common.Address
	client   *ethclient.Client
	watcher  chan types.RoundInfo
	closed   chan struct{}
}

func NewChainClock(url string, name string, reporter string) *ChainClock {
	logger := log.WithField("chain", name)
	client, e := ethclient.Dial(url)
	if e != nil {
		logger.Error("create chainclock reader failed, dial url err :", e)
		return nil
	}

	return &ChainClock{
		client:   client,
		name:     name,
		reporter: common.HexToAddress(reporter),
		logger:   logger,
		closed:   make(chan struct{}),
		watcher:  nil,
	}
}

func (c *ChainClock) ChainName() string {
	return c.name
}

func (c *ChainClock) Start() error {
	newHead := make(chan *ethtypes.Header, 100)
	sub, err := c.client.SubscribeNewHead(context.Background(), newHead)
	if err != nil {
		c.logger.Error("subscribe new head failed, err:", err)
		return err
	}
	defer sub.Unsubscribe()
	c.logger.Info("subcribe new head succeed")
	for {
		select {
		case <-c.closed:
			break
		case e := <-sub.Err():
			c.logger.Errorf("chainclock reader on %s sunscribe error %s\n", c.name, e.Error())
			break
		case header, ok := <-newHead:
			if !ok {
				continue
			}
			round := types.RoundInfo{
				Timestamp: header.Time,
				Txsinfo:   make([]*types.TxPackage, 0),
			}

			block, e := c.client.BlockByNumber(context.Background(), header.Number)
			if e != nil {
				c.logger.Error("get block by number failed", "err", e)
			} else {
				c.logger.Debug("get new block ", "number ", header.Number)
				for idx, tx := range block.Transactions() {
					from, err := c.client.TransactionSender(context.Background(), tx, block.Hash(), uint(idx))
					if err != nil {
						c.logger.Error("get transaction sender failed", "err", err)
						continue
					}
					if bytes.Compare(from.Bytes(), c.reporter.Bytes()) == 0 && len(tx.Data()) > 0 {
						//if bytes.Compare(tx.To().Bytes(), addr.Bytes()) == 0 && len(tx.Data()) > 0 {
						//c.logger.Debug("get new tx from monitor ", "data is ", hex.EncodeToString(tx.Data()))
						t := &types.TxPackage{}
						t.Blockhash = block.Hash().String()
						t.Time = block.Time()
						t.Tx = tx
						round.Txsinfo = append(round.Txsinfo, t)
					}
				}
			}
			if c.watcher != nil {
				// send round info to watcher.
				c.watcher <- round
			}
		}
	}
}

func (c *ChainClock) Stop() error {
	close(c.closed)
	return nil
}

func (c *ChainClock) WatchClock() chan types.RoundInfo {
	if c.watcher == nil {
		c.watcher = make(chan types.RoundInfo, 100)
	}
	return c.watcher
}
