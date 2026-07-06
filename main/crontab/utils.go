package crontab

import (
	"encoding/json"
	"github.com/robfig/cron/v3"
	"time"

	"network_info/main/conf"
	"network_info/main/dbConn" // 引入你的 dbConn 包
	"network_info/main/webTypes/gnbTypes"
)

var (
	crontab *cron.Cron
)

// Run 启动所有定时任务
func Run() {
	crontab = cron.New()

	// 每5分钟同步一次 Redis → Map
	_, err := crontab.AddFunc("*/5 * * * *", syncBaseStationCache)
	if err != nil {
		conf.Log.Errorf("注册基站同步定时任务失败: %v", err)
		return
	}

	crontab.Start()
	conf.Log.Info("crontab 定时任务已启动 → 每5分钟同步基站缓存")
}

// syncBaseStationCache 核心同步函数
func syncBaseStationCache() {
	start := time.Now()
	conf.Log.Info("=== 开始执行基站缓存同步任务 ===")

	dbConn.CacheMutex.Lock() // 使用你 dbConn 包里的锁（建议改成导出）
	defer dbConn.CacheMutex.Unlock()

	// 1. 从 Redis 获取所有最新基站数据
	redisData := make(map[uint32]gnbTypes.BaseStation)

	iter := dbConn.Rdb.Scan(0, "bs:*", 200).Iterator()
	for iter.Next() {
		key := iter.Val()
		data, err := dbConn.Rdb.Get(key).Result()
		if err != nil {
			continue
		}

		var bs gnbTypes.BaseStation
		if err := json.Unmarshal([]byte(data), &bs); err != nil {
			continue
		}

		redisData[bs.StationID] = bs
	}

	// 2. 处理删除：Redis中没有，但Map中有的 → 删除
	for id := range dbConn.BaseStationCache {
		if _, exists := redisData[id]; !exists {
			delete(dbConn.BaseStationCache, id)
			conf.Log.Infof("定时同步：基站 %d 已从内存缓存删除（Redis中不存在）", id)
		}
	}

	// 3. 处理新增和更新
	for id, bs := range redisData {
		old, exists := dbConn.BaseStationCache[id]
		if !exists {
			// 新增
			dbConn.BaseStationCache[id] = bs
			conf.Log.Infof("定时同步：新增基站 %d (%s)", id, bs.IP)
		} else if old.IP != bs.IP || old.Name != bs.Name || old.Region != bs.Region {
			// 更新（字段有变化）
			dbConn.BaseStationCache[id] = bs
			conf.Log.Infof("定时同步：更新基站 %d 信息", id)
		}
	}

	conf.Log.Infof("基站缓存同步完成，共 %d 个基站，耗时 %v", len(dbConn.BaseStationCache), time.Since(start))
}
