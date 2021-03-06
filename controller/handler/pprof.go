package handler

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
)

func (a ApiHandler) MemProfile(c *gin.Context) {
	f, err := os.Create("memprofile.log")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	runtime.GC()    // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
	ResponseSuccess(c, nil)
}

func (a ApiHandler) DebugStart(c *gin.Context) {

	go func() {
		http.ListenAndServe(":8085", nil)
	}()
	ResponseSuccess(c, nil)
}
