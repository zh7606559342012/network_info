package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"network_info/main/conf"
	"network_info/main/models/entity"
	"network_info/main/models/service"
	"strconv"
)

type GnbLyrController struct {
}

var GnbLyrService = new(service.GnbLyrService)

func (g GnbLyrController) HandlerAddGnb(c *gin.Context) {
	conf.Log.Debugf("HandlerAddGnb FuncEntry.")

	creatV, createError := GnbLyrService.Creategnb(c)
	if createError != nil {
		c.JSON(http.StatusInternalServerError, entity.Response{Message: createError.Error()})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, entity.Response{Code: strconv.Itoa(creatV.Code), Message: creatV.Message})
	return
}

func (g GnbLyrController) HandlerDelGnb(c *gin.Context) {
	conf.Log.Debugf("HandlerDelGnb FuncEntry.")

	delV, createError := GnbLyrService.Deletegnb(c)
	if createError != nil {
		c.JSON(http.StatusInternalServerError, entity.Response{Message: createError.Error()})
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, entity.Response{Code: strconv.Itoa(delV.Code), Message: delV.Message})
	return
}
