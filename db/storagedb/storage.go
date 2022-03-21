package storagedb

import (
	"github.com/hashrs/consensusdemo/config"
	"github.com/hashrs/consensusdemo/db"
)

func NewStorageDB(conf *config.Config) db.Database {
	return newRedisDB(conf)
	//return newSyncMapDB(conf)
}
