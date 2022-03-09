package storagedb

import (
	"github.com/hashrs/consensusdemo/config"
	"github.com/hashrs/consensusdemo/lib/redispool"
	"log"
)

type redisdb struct {
	db *redispool.RedisPool
}

func (s *redisdb) Get(key interface{}) (interface{}, bool) {
	ks, ok := key.(string)
	if !ok {
		panic("storage not support key non string type")
	}
	vs := s.db.Get(ks)
	if len(vs) == 0 {
		return nil, false
	} else {
		return vs, true
	}
}

func (s *redisdb) Set(key, value interface{}) error {
	ks, ok := key.(string)
	if !ok {
		panic("storage not support key non string type")
	}
	vs, ok := value.(string)
	if !ok {
		panic("storage not support key non string type")
	}
	s.db.Set(ks, vs)
	return nil
}

func newRedisDB(conf *config.Config) *redisdb {
	hosts, dbNum, passwd := conf.StorageConn()
	r, err := redispool.NewRedisPool(hosts, dbNum, passwd)
	if err != nil {
		log.Fatal("storage db redis connect failed")
	}

	return &redisdb{db: r}
}
