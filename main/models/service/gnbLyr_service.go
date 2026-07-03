package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"network_info/main/conf"
	"network_info/main/webTypes"
	"network_info/main/webTypes/gnbTypes"
)

type GnbLyrService struct{}

func (b GnbLyrService) Creategnb(c *gin.Context) (webTypes.JavaResp, error) {
	var baseStation gnbTypes.BaseStation
	if err := c.ShouldBindJSON(&baseStation); err != nil {
		conf.Log.Error("json.ShouldBindJSON failed,err:%s.", err)
		return webTypes.JavaResp{Code: 400, Message: "参数解析失败"}, err
	}

	// 参数校验
	if baseStation.StationID == 0 {
		conf.Log.Error("add station_id can not be 0.")
		return webTypes.JavaResp{Code: 400, Message: "station_id 不能为空"}, errors.New("station_id cannot be 0")
	}

	if baseStation.IP == "" {
		conf.Log.Error("add ip can not be null.")
		return webTypes.JavaResp{Code: 400, Message: "IP 不能为空"}, errors.New("ip cannot be empty")
	}

	// ==================== 存入 Redis（永久缓存） ====================
	if err := CacheBaseStation(baseStation); err != nil {
		conf.Log.Error("CacheBaseStation failed, err: %s", err)
		return webTypes.JavaResp{Code: 500, Message: "缓存失败"}, err
	}

	conf.Log.Infof("基站 %d (%s) 已成功缓存到 Redis", baseStation.StationID, baseStation.IP)

	return webTypes.JavaResp{
		Code:    200,
		Message: "ok",
	}, nil
}

func (b GnbLyrService) Deletegnb(c *gin.Context) (webTypes.JavaResp, error) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		conf.Log.Error("ioutil.ReadAll(c.Request.Body) failed,err:%s.", err)
		return webTypes.JavaResp{Code: 400, Message: "读取请求体失败"}, err
	}

	var stationIDs []uint32
	if err := json.Unmarshal(body, &stationIDs); err != nil {
		conf.Log.Error("json.Unmarshal failed,err:%s.", err)
		return webTypes.JavaResp{Code: 400, Message: "JSON 解析失败"}, err
	}

	if len(stationIDs) == 0 {
		return webTypes.JavaResp{Code: 400, Message: "删除列表不能为空"}, errors.New("empty delete list")
	}

	// 删除 Redis 中的记录
	successCount := 0
	for _, id := range stationIDs {
		if err := DeleteBaseStation(id); err != nil {
			conf.Log.Warnf("删除基站 %d 失败: %v", id, err)
		} else {
			successCount++
			conf.Log.Infof("基站 %d 已成功从 Redis 删除", id)
		}
	}

	return webTypes.JavaResp{
		Code:    200,
		Message: fmt.Sprintf("删除完成，成功删除 %d 条记录", successCount),
	}, nil
}
