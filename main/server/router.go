package server

import (
	"github.com/gin-gonic/gin"
	"network_info/main/controllers"
)

func NewRouter() *gin.Engine {
	router := gin.New()

	omsRouter := router.Group("/monitor_agent/v1")
	akaRouter := router.Group("/monitor_agent/v1/common/eap_aka")

	omsRoutes(omsRouter)
	akaRouters(akaRouter)

	return router
}

func omsRoutes(router *gin.RouterGroup) {
	omsLyr := new(controllers.OmsLyrController)
	router.GET("/common/heartbeat", omsLyr.HandlerHeartbeat)
}
