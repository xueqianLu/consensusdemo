package chain

import (
	"bytes"
	"context"
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
	client, e := ethclient.Dial(url)
	if e != nil {
		log.Error("create chain reader failed, dial url err :", e)
		return nil
	}
	logger := log.WithField("chain", name)
	return &ChainReader{client: client, name: name, logger: logger}
}

func (c *ChainReader) SubscribeTransaction(addr common.Address, stop chan struct{}, report chan *types.Md5tx) error {

	newHead := make(chan *ethtypes.Header, 100)
	sub, err := c.client.SubscribeNewHead(context.Background(), newHead)
	if err != nil {
		log.Error("subscribe new head failed, err:", err)
		return err
	}
	defer sub.Unsubscribe()
	for {
		select {
		case <-stop:
			break
		case e := <-sub.Err():
			log.Errorf("chain reader on %s sunscribe error %s\n", c.name, e.Error())
			break
		case header, ok := <-newHead:
			if !ok {
				continue
			}
			block, e := c.client.BlockByNumber(context.Background(), header.Number)
			if e != nil {
				log.Error("get block by number failed", "err", e)
			} else {
				for _, tx := range block.Transactions() {
					if bytes.Compare(tx.To().Bytes(), addr.Bytes()) == 0 && len(tx.Data()) > 0 {
						t := &types.Md5tx{}
						t.Blockhash = block.Hash().String()
						t.Time = block.Time()
						t.Tx = tx
						report <- t
					}
				}
			}
		}

	}
}
