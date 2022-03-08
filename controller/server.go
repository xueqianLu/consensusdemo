package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/hashrs/consensusdemo/config"
	"github.com/hashrs/consensusdemo/controller/handler"
	"github.com/hashrs/consensusdemo/core/chaindb"
	"github.com/hashrs/consensusdemo/core/globaldb"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func InitRouter(hand handler.ApiHandler, r *gin.Engine) {
	api := r.Group("/api")
	v1 := api.Group("/v1")
	v1.GET("/transaction", hand.GetTransaction)
	v1.GET("/receipt", hand.GetReceipt)
	v1.GET("/balance", hand.GetAccount)
	v1.GET("/block", hand.GetBlock)
	v1.POST("/initaccount", hand.InitAccount)
	v1.POST("/memprofile", hand.MemProfile)
}

type APIServer struct {
	*gin.Engine
	chainReader  handler.ChainReader
	globalReader handler.GlobalReader
	running      bool
}

func NewServer(chain chaindb.ChainDB, global globaldb.GlobalDB) *APIServer {
	r := gin.New()
	//r.Use(Cors()) //默认跨域
	r.Use(gin.Recovery())
	hand := handler.NewHandler(chain, global)
	InitRouter(hand, r)
	return &APIServer{r, chain, global, false}
}

func (s *APIServer) Start() error {
	if s.running {
		return nil
	}

	defer func() { s.running = false }()

	s.running = true
	addr := config.GetConfig().APIServerAddr()
	log.Infof("Start API Server on %s", addr)
	if err := s.Run(addr); err != nil {
		log.WithError(err).Fatalln("fail to run server")
		return err
	} else {
		return nil
	}
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Set("content-type", "application/json")
		}
		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}
