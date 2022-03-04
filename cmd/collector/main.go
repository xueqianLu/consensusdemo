package main

import (
	"github.com/hashrs/consensusdemo/config"
	"github.com/hashrs/consensusdemo/node"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"sync"
)

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
	logfile, err := os.OpenFile(conf.LogFile(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend|0755)
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

func main() {
	conf := config.GetConfig()
	InitLog(conf)

	cnode := node.NewNode(conf)
	cnode.Start()

	log.Info("go to start miner ")
	worker := cnode.Miner()
	worker.Start()

	wg := sync.WaitGroup{}
	wg.Add(1)

	wg.Wait()
}
