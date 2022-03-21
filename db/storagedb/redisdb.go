package storagedb

import (
	"errors"
	"github.com/hashrs/consensusdemo/config"
	"github.com/hashrs/consensusdemo/lib/redispool"
	"log"
)

type redisdb struct {
	db *redispool.RedisPool
}

func (s *redisdb) Get(key interface{}) ([]byte, bool) {
	ks, ok := key.(string)
	if !ok {
		panic("storage not support key non string type")
	}
	vs := s.db.Get(ks)
	if len(vs) == 0 {
		return []byte{}, false
	} else {
		return []byte(vs), true
	}
}

func (s *redisdb) Del(key interface{}) error {
	ks, ok := key.(string)
	if !ok {
		return errors.New("storage not support key non string type")
	}
	s.db.Del(ks)
	return nil
}

func (s *redisdb) Set(key interface{}, value []byte) error {
	ks, ok := key.(string)
	if !ok {
		panic("storage not support key non string type")
	}

	s.db.Set(ks, string(value))
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
