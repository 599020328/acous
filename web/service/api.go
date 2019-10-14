package service

import (
	"acous/web/conf"
	"github.com/gin-gonic/gin"
	"log"
)

var configure *conf.Conf
var workChain []string
var nowWorkNum int
var timeLog int64

func StartWebService(c *conf.Conf) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	configure = c

	router.POST("/begin_as_source", BeginHandler)
	router.POST("/begin_update_delay", BeginUpdateDelayHandler)
	router.POST("/get_neighbor_delay", GetNeighborDelayHandler)

	router.POST("/request", RequestHandler)
	router.POST("/response", ResultHandler)
	router.POST("/update", UpdateHandler)

	router.POST("/log_config", LogConfig)

	router.POST("/test", TestSt)

	err := router.Run("0.0.0.0:" + c.ServerPort)
	if err != nil {
		log.Fatal(err)
	}
}
