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

func (c *Config) RedisConn() (conn string, dbNum string, passwd string) {
	conn = c.cfg.Section("redis").Key("conn").String()
	dbNum = c.cfg.Section("redis").Key("dbNum").String()
	passwd = c.cfg.Section("redis").Key("password").String()
	return
}

func (c *Config) Chains() map[string]string {
	keys := c.cfg.Section("chainreader").Keys()
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
