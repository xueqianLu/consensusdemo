package storagedb

import (
	"github.com/hashrs/consensusdemo/config"
	"github.com/hashrs/consensusdemo/db"
	"sync"
)

type syncmapdb struct {
	db sync.Map
}

func (s *syncmapdb) NewBatch() db.Batch {
	// not support now.
	return nil
}

func (s *syncmapdb) Get(key interface{}) ([]byte, bool) {
	v, ok := s.db.Load(key)
	if !ok {
		return nil, false
	} else {
		return v.([]byte), true
	}
}

func (s *syncmapdb) Del(key interface{}) error {
	s.db.Delete(key)
	return nil
}

func (s *syncmapdb) Set(key interface{}, value []byte) error {
	s.db.Store(key, value)
	return nil
}

func newSyncMapDB(conf *config.Config) *syncmapdb {
	return &syncmapdb{}
}
