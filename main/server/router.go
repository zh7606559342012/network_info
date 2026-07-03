package server

import (
	"github.com/gin-gonic/gin"
	"network_info/main/controllers"
)

func NewRouter() *gin.Engine {
	router := gin.New()

	omsRouter := router.Group("/monitor_agent/v1")
	gnbRouter := router.Group("/monitor_agent/v1/common/eap_aka")

	omsRoutes(omsRouter)
	gnbRouters(gnbRouter)

	return router
}

func omsRoutes(router *gin.RouterGroup) {
	omsLyr := new(controllers.OmsLyrController)
	router.GET("/common/heartbeat", omsLyr.HandlerHeartbeat)
}

func gnbRouters(router *gin.RouterGroup) {
	gnbLyr := new(controllers.GnbLyrController)
	//router.GET("", gnbListLyr.HandlerGetPageGnbList)
	router.POST("", gnbLyr.HandlerAddGnb)
	router.DELETE("", gnbLyr.HandlerDelGnb)
}
