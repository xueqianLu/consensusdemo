package txpool

import (
	"encoding/json"
	"github.com/hashrs/consensusdemo/config"
	"github.com/hashrs/consensusdemo/core/objectpool"
	"github.com/hashrs/consensusdemo/lib/redispool"
	"github.com/hashrs/consensusdemo/types"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

var (
	redisMQName = "list"
)

type TxPool struct {
	routines uint
	rp       *redispool.RedisPool
	closed   chan struct{}
	allTx    sync.Map
	wg       sync.WaitGroup
}

func NewTxpool(conf *config.Config) *TxPool {
	hosts, dbNum, passwd := conf.RedisConn()
	r, err := redispool.NewRedisPool(hosts, dbNum, passwd)
	if err != nil {
		log.Fatal("txpool redis connect failed")
	}
	log.Info("txpool redis connect succeed.")

	return &TxPool{
		rp:       r,
		routines: conf.TxPoolRoutines(),
		closed:   make(chan struct{}),
	}
}

func (t *TxPool) Start() {
	//log.Debug("txpool start ", "routines ", t.routines)
	for i := uint(0); i < t.routines; i++ {
		t.wg.Add(1)
		go t.loop(i)
	}
}

func (t *TxPool) Stop() {
	close(t.closed)
	t.wg.Wait()
}

func (t *TxPool) GetTxs(packedHash string) []*types.FurtherTransaction {
	if v, exist := t.allTx.Load(packedHash); !exist {
		return []*types.FurtherTransaction{}
	} else {
		txs := v.([]*types.FurtherTransaction)
		return txs
	}
}

func (t *TxPool) ResetTxs(packedhashs []string) {
	for _, hash := range packedhashs {
		t.allTx.Delete(hash)
	}
}

func (t *TxPool) loop(idx uint) {
	defer t.wg.Done()

	newtx := make(chan string, 1000000)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		l := log.WithField("routine", "redis").WithField("index", idx)
		for {
			select {
			case <-t.closed:
				return
			default:
				tx, err := t.rp.RPop(redisMQName)
				if err != nil {
					l.Error("redis rpop", "err", err)
				}
				if len(tx) == 0 {
					//l.Debug("got 0 tx from redis")
					time.Sleep(time.Millisecond * 100)
				} else {
					//l.Debug("got tx from redis, ", " hash ",)
					newtx <- tx
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
						//l.Debug("save redis tx to map")
						txs := GetTransactions(pair)
						t.allTx.Store(pair.GetHash(), txs)
						l.Debug("got tx from redis, ", " hash ", pair.GetHash())
					} else {
						l.Error("unmarshal tx pair failed", "err", err)
					}
				}(tx)
			case <-t.closed:
				return
			}
		}
	}()
	wg.Wait()
}

func GetTransactions(t types.TxPair) []*types.FurtherTransaction {
	var txs = make([]*types.FurtherTransaction, len(t.Txs))
	for i := 0; i < len(txs); i++ {
		tx := t.Txs[i]
		ptx := objectpool.GetTransactionObject()
		err := ptx.UnmarshalBinary(tx.TxBytes)
		if err != nil {
			log.Error("decode rlp tx ", "err", err)
			continue
		}
		ptx.From.SetBytes(tx.From)
		txs[i] = ptx
	}
	return txs
}
