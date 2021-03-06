package config

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

type Config struct {
	cfg *ini.File
}

var gConfig = &Config{}

func init() {
	confpath := "app.conf"
	cfg, err := ini.Load(confpath)
	if err != nil {
		log.Println("文件读取错误", err)
		os.Exit(1)
	}
	gConfig.cfg = cfg
}

func GetConfig() *Config {
	return gConfig
}

func (c *Config) TxPoolRoutines() uint {
	threads, err := c.cfg.Section("txpool").Key("routines").Uint()
	if err != nil {
		return 10
	} else {
		return threads
	}
}

func (c *Config) RedisConn() (conn string, dbNum string, passwd string) {
	conn = c.cfg.Section("redis").Key("conn").String()
	dbNum = c.cfg.Section("redis").Key("dbNum").String()
	passwd = c.cfg.Section("redis").Key("password").String()
	return
}

func (c *Config) StorageConn() (conn string, dbNum string, passwd string) {
	conn = c.cfg.Section("storage").Key("conn").String()
	dbNum = c.cfg.Section("storage").Key("dbNum").String()
	passwd = c.cfg.Section("storage").Key("password").String()
	return
}

func (c *Config) LevelDBConf() (file string, cache int, handles int) {
	file = c.cfg.Section("leveldb").Key("file").String()
	cache, _ = c.cfg.Section("leveldb").Key("cache").Int()
	handles, _ = c.cfg.Section("leveldb").Key("handle").Int()
	return
}

func (c *Config) Chains() map[string]string {
	keys := c.cfg.Section("chain").Keys()
	var chain = make(map[string]string)
	for _, key := range keys {
		chain[key.Name()] = key.Value()
	}
	return chain
}

func (c *Config) Address() string {
	return c.cfg.Section("monitor").Key("account").String()
}

func (c *Config) LogLevel() string {
	level := c.cfg.Section("log").Key("level").String()
	if len(level) == 0 {
		return "info"
	}
	return level
}

func (c *Config) LogFile() string {
	name := c.cfg.Section("log").Key("filename").String()
	if len(name) == 0 {
		name = "consensus.log"
	}
	filename := filepath.Join("logs", name)
	return filename
}

func (c *Config) APIServerAddr() string {
	addr := c.cfg.Section("apiserver").Key("addr").String()
	if len(addr) == 0 {
		addr = "127.0.0.1:9800"
	}
	return addr
}
