package main

import (
	"encoding/json"
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

func loop(rp *redispool.RedisPool, stop chan struct{}, allTx *sync.Map, idx int) {

	newtx := make(chan string, 1000000)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		l := log.WithField("routine", "redis").WithField("index", idx)
		for {
			select {
			case <-stop:
				return
			default:
				//l.Debug("go to pop from redis")
				tx, err := rp.RPop(redisMQName)
				if err != nil {
					l.Error("redis rpop", "err", err)
				}
				if len(tx) == 0 {
					//l.Debug("got 0 tx from redis")
					time.Sleep(time.Millisecond * 10)
				} else {
					//l.Debug("send to store tx")
					go func() { newtx <- tx }()
					//l.Debug("after send to store tx")
				}
			}
		}
	}()
	wg.Add(1)
	go func() {
		wg.Done()
		l := log.WithField("routine", "store").WithField("index", idx)
		for {
			select {
			case tx := <-newtx:
				go func(cachetx string) {
					// get tx pair from redis and save to map.
					var pair types.TxPair
					if err := json.Unmarshal([]byte(tx), &pair); err == nil {
						l.Debug("got txpair from redis ", "hash ", pair.GetHash(), " with tx count ", len(pair.GetTransactions()))
						//txs := pair.GetTransactions()
						//
						//for _,t := range txs {
						//	l.Debug("origin tx ", "from is ", t.From, "tx info is ", t.Transaction)
						//}
						allTx.Store(pair.GetHash(), &pair)
					} else {
						l.Error("unmarshal tx pair failed", "err", err)
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
	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "2006-01-02T15:04:05.999999999Z07:00"
	Formatter.FullTimestamp = true
	Formatter.ForceColors = true
	log.SetFormatter(Formatter)

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
			log.Debug("got new package tx from chain ", reader.ChainName())
			hashs := tx.Hashs()
			for _, m := range hashs {
				if txpair, ok := allTx.Load(m); !ok {
					log.Error("not found original tx with hash ", m)
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
	var maxPackTxs = 10000
	var packtxs = make([]*types.FurtherTransaction, 0, maxPackTxs)
	for {
		select {
		case info, ok := <-receive:
			if !ok {
				return
			}
			txs := info.pair.GetTransactions()
			log.Debug("in sortTx ", "got tx info", "hash ", info.pair.GetHash(), " original tx count ", len(txs))

			packtxs = append(packtxs, txs...)
			if len(packtxs) > maxPackTxs {
				log.Info("package tx finished.", "total txs ", len(packtxs))
				// todo: packtxs sorting with timestamp and md5.

				packtxs = make([]*types.FurtherTransaction, 0, maxPackTxs)
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
	for i := 0; i < 10; i++ {
		go loop(db, stop, allTxMap, i)
	}
	for name, rpc := range chains {
		reader := chain.NewChainReader(rpc, name)
		go readChain(reader, stop, allTxMap, fullTxCh)
	}

	wg.Wait()

}
