package storagedb

import (
	"github.com/hashrs/consensusdemo/config"
	"sync"
)

type syncmapdb struct {
	db sync.Map
}

func (s *syncmapdb) Get(key interface{}) (interface{}, bool) {
	v, ok := s.db.Load(key)
	if !ok {
		return nil, false
	} else {
		return v, true
	}
}

func (s *syncmapdb) Set(key, value interface{}) error {
	s.db.Store(key, value)
	return nil
}

func newSyncMapDB(conf *config.Config) *syncmapdb {
	return &syncmapdb{}
}
