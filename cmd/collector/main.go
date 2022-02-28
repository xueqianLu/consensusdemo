package main

import (
	"encoding/json"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashrs/consensusdemo/chain"
	"github.com/hashrs/consensusdemo/cmd/collector/config"
	"github.com/hashrs/consensusdemo/redispool"
	"github.com/hashrs/consensusdemo/types"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"sync"
	"time"
)

var (
	redisMQName = "list"
)

func loop(rp *redispool.RedisPool, stop chan struct{}, allTx *sync.Map) {
	newtx := make(chan string, 1000000)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				tx, err := rp.RPop(redisMQName)
				if err != nil {
					log.Error("redis rpop", "err", err)
				}
				if len(tx) == 0 {
					time.Sleep(time.Microsecond * 50)
				} else {
					newtx <- tx
				}
			}
		}
	}()
	wg.Add(1)
	go func() {
		wg.Done()
		for {
			select {
			case tx := <-newtx:
				go func(cachetx string) {
					// get tx pair from redis and save to map.
					var pair types.TxPair
					if err := json.Unmarshal([]byte(tx), &pair); err == nil {
						log.Debug("got txpair from redis ", "hash ", pair.GetHash())
						allTx.Store(pair.GetHash(), &pair)
					} else {
						log.Error("unmarshal tx pair failed", "err", err)
					}
				}(tx)
			case <-stop:
				return
			}
		}
	}()

	wg.Wait()

}

func InitLog(conf *config.Config) {
	level := conf.LogLevel()
	switch level {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
	logfile, err := os.OpenFile(conf.LogFile(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend|755)
	//logfile, err := os.OpenFile(conf.LogFile(), os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Error("can't open log file", conf.LogFile())
	}
	log.SetFormatter(&nested.Formatter{
		TimestampFormat: time.RFC3339,
	})

	log.SetOutput(io.MultiWriter(logfile))
}

type txInfo struct {
	pair      *types.TxPair
	onchainTx *types.TxPackage
}

func readChain(reader *chain.ChainReader, stop chan struct{}, allTx *sync.Map, sortch chan *txInfo) {
	newTx := make(chan *types.TxPackage, 1000)
	addr := common.HexToAddress(config.GetConfig().Address())
	go reader.SubscribeTransaction(addr, stop, newTx)
	for {
		select {
		case tx := <-newTx:
			log.Debug("got new md5tx from chain ", reader.ChainName())
			hashs := tx.Hashs()
			for _, m := range hashs {
				if txpair, ok := allTx.Load(m); !ok {
					log.Error("not found original tx with md5 ", m)
					continue
				} else {
					//log.Info("original packaged at chain", reader.ChainName(), "block time ", tx.Time)
					info := &txInfo{
						onchainTx: tx,
						pair:      txpair.(*types.TxPair),
					}
					sortch <- info
				}
			}
		}
	}

}

func sortTx(receive chan *txInfo) {
	var maxPackTxs = 100
	var packtxs = make([]*txInfo, 0, maxPackTxs)
	for {
		select {
		case info, ok := <-receive:
			if !ok {
				return
			}
			log.Debug("in sortTx ", "got tx info", "hash ", info.pair.Hash)
			txs := info.pair.GetTransactions()
			maxPackTxs = maxPackTxs + len(txs) - len(txs)

			packtxs = append(packtxs, info)
			if len(packtxs) == maxPackTxs {
				log.Info("package tx finished.")
				// todo: packtxs sorting with timestamp and md5.

				packtxs = make([]*txInfo, 0, maxPackTxs)
			}
		}
	}

}
func main() {
	conf := config.GetConfig()
	InitLog(conf)
	hosts, dbNum, passwd := conf.RedisConn()
	db, err := redispool.NewRedisPool(hosts, dbNum, passwd)
	if err != nil {
		log.Fatal("redis connect failed")
	}
	log.Info("redis connect succeed.")
	defer db.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)

	allTxMap := &sync.Map{}
	stop := make(chan struct{})
	chains := conf.Chains()

	var fullTxCh = make(chan *txInfo, 10000000)
	go sortTx(fullTxCh)
	go loop(db, stop, allTxMap)
	for name, rpc := range chains {
		reader := chain.NewChainReader(rpc, name)
		go readChain(reader, stop, allTxMap, fullTxCh)
	}

	wg.Wait()

}
