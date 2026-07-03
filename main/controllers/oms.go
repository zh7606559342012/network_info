package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"network_info/main/conf"
	"network_info/main/dbConn"
	"network_info/main/models/entity"
	"strconv"
	"time"
)

type OmsLyrController struct {
}

var heartbeatCount = 0

func (b OmsLyrController) HandlerHeartbeat(c *gin.Context) {
	conf.Log.Debugf("HandlerHeartbeat FuncEntry.")
	heartbeatCount += 1
	ip := c.Query("ip")
	if conf.CmnConf.OMSInfo.OMSIp != "" {
		heartbeatCount = 0
		conf.Log.Debug("HandlerHeartbeat FuncEntry SET AGENT OMSIP: ", conf.CmnConf.RdsInfo.RdsOmsIp)
		t := time.Now().Format("2006-01-02 15:04:05") //固定格式产生time.str
		c.JSON(http.StatusOK, entity.Response{TimeStamp: t, Code: strconv.Itoa(http.StatusOK), Message: "ok"})
		return
	}

	if heartbeatCount >= 3 && ip != "" {
		conf.CmnConf.RdsInfo.RdsOmsIp = ip
		heartbeatCount = 0
		dbConn.Rdb.Set("omsIp", ip, 0)
		_, err := dbConn.Rdb.Ping().Result()
		if err != nil {
			conf.Log.Error("ReConnRedisClient@@@")
			initRedisErr := dbConn.InitRedisClient()
			if initRedisErr != nil {
				t := time.Now().Format("2006-01-02 15:04:05") //固定格式产生time.str
				c.JSON(http.StatusOK, entity.Response{TimeStamp: t, Code: strconv.Itoa(http.StatusOK), Message: "Redis disconnect reset connection..."})
				return
			}
		}
	}
	t := time.Now().Format("2006-01-02 15:04:05") //固定格式产生time.str
	c.JSON(http.StatusOK, entity.Response{TimeStamp: t, Code: strconv.Itoa(http.StatusOK), Message: "ok"})
	return
}
