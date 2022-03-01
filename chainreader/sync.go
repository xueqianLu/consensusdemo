package chainreader

import (
	"bytes"
	"context"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashrs/consensusdemo/types"
	log "github.com/sirupsen/logrus"
)

type ChainReader struct {
	logger *log.Entry
	name   string
	client *ethclient.Client
}

func NewChainReader(url string, name string) *ChainReader {
	logger := log.WithField("chain", name)
	client, e := ethclient.Dial(url)
	if e != nil {
		logger.Error("create chainreader reader failed, dial url err :", e)
		return nil
	}

	return &ChainReader{client: client, name: name, logger: logger}
}

func (c *ChainReader) ChainName() string {
	return c.name
}

func (c *ChainReader) SubscribeTransaction(addr common.Address, stop chan struct{}, report chan *types.TxPackage) error {

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
		case <-stop:
			break
		case e := <-sub.Err():
			c.logger.Errorf("chainreader reader on %s sunscribe error %s\n", c.name, e.Error())
			break
		case header, ok := <-newHead:
			if !ok {
				continue
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
					if bytes.Compare(from.Bytes(), addr.Bytes()) == 0 && len(tx.Data()) > 0 {
						//if bytes.Compare(tx.To().Bytes(), addr.Bytes()) == 0 && len(tx.Data()) > 0 {
						c.logger.Debug("get new tx from monitor ", "data is ", hex.EncodeToString(tx.Data()))
						t := &types.TxPackage{}
						t.Blockhash = block.Hash().String()
						t.Time = block.Time()
						t.Tx = tx
						report <- t
					}
					idx = idx
				}
			}
		}

	}
}
