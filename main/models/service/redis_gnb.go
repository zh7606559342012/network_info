// redis_gnb.go
package service

import (
	"encoding/json"
	"fmt"

	"network_info/main/dbConn"
	"network_info/main/webTypes/gnbTypes"
)

const BaseStationPrefix = "bs:" // Key 前缀

// CacheBaseStation 缓存基站信息（永久缓存）
func CacheBaseStation(bs gnbTypes.BaseStation) error {
	if bs.StationID == 0 {
		return fmt.Errorf("station_id 不能为空")
	}

	data, err := json.Marshal(bs)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s%d", BaseStationPrefix, bs.StationID)

	// 永久缓存：把过期时间设置为 0
	return dbConn.Rdb.Set(key, data, 0).Err()
}

// DeleteBaseStation 删除基站缓存
func DeleteBaseStation(stationID uint32) error {
	if stationID == 0 {
		return fmt.Errorf("station_id 不能为空")
	}

	key := fmt.Sprintf("%s%d", BaseStationPrefix, stationID)
	return dbConn.Rdb.Del(key).Err()
}

// GetBaseStation 获取单个基站
func GetBaseStation(stationID uint32) (gnbTypes.BaseStation, error) {
	key := fmt.Sprintf("%s%d", BaseStationPrefix, stationID)

	val, err := dbConn.Rdb.Get(key).Result()
	if err != nil {
		return gnbTypes.BaseStation{}, err
	}

	var bs gnbTypes.BaseStation
	err = json.Unmarshal([]byte(val), &bs)
	return bs, err
}
