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

func loop(rp *redispool.RedisPool, stop chan struct{}, allTx sync.Map) {
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
				tx, _ := rp.LPop("Transactions")
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
						allTx.Store(pair.GetMd5(), pair.GetTransaction())
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
	logfile, err := os.OpenFile(conf.LogFile(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	//logfile, err := os.OpenFile(conf.LogFile(), os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Error("can't open log file", conf.LogFile())
	}
	log.SetFormatter(&nested.Formatter{
		TimestampFormat: time.RFC3339,
	})

	log.SetOutput(io.MultiWriter(logfile))
}

func readChain(reader *chain.ChainReader, stop chan struct{}, allTx sync.Map) {
	newTx := make(chan *types.Md5tx, 1000)
	addr := common.HexToAddress(config.GetConfig().Address())
	go reader.SubscribeTransaction(addr, stop, newTx)
	for {
		select {
		case tx := <-newTx:
			for _, m := range tx.Md5s() {
				if origialTx, ok := allTx.Load(m); !ok {

				}
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

	allTxMap := sync.Map{}
	stop := make(chan struct{})
	chains := conf.Chains()
	go loop(db, stop, allTxMap)
	for name, rpc := range chains {
		reader := chain.NewChainReader(rpc, name)
		go readChain(reader, stop, allTxMap)
	}

	wg.Wait()

}
