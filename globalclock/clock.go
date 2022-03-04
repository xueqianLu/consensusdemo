package globalclock

import (
	"github.com/hashrs/consensusdemo/config"
	"github.com/hashrs/consensusdemo/globalclock/chainclock"
	"github.com/hashrs/consensusdemo/types"
	log "github.com/sirupsen/logrus"
)

type Clock interface {
	WatchClock() chan types.RoundInfo
}

func NewClock(conf *config.Config) Clock {
	reporter := conf.Address()
	chains := conf.Chains()
	for name, rpc := range chains {
		c := chainclock.NewChainClock(rpc, name, reporter)
		if c != nil {
			return c
		}
	}
	log.Error("have no available clock")
	return nil
}
