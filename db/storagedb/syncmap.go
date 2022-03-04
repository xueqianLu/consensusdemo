package storagedb

import (
	"github.com/hashrs/consensusdemo/db"
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

func NewStorageDB() db.Database {
	return &syncmapdb{}
}
