package dbConn

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"network_info/main/conf"
	"network_info/main/webTypes/gnbTypes"
	"sync"
)

var Rdb *redis.Client
var (
	BaseStationCache map[uint32]gnbTypes.BaseStation
	CacheMutex       sync.RWMutex
)

func DbConnInit() {
	InitRedisClient()
}

func InitRedisClient() (err error) {
	rdsAddr := conf.CmnConf.DBConn.RedisInfo.Addr + ":" + conf.CmnConf.DBConn.RedisInfo.Port
	rdsPw := conf.CmnConf.DBConn.RedisInfo.Password
	Rdb = redis.NewClient(&redis.Options{
		Addr:     rdsAddr,
		Password: rdsPw, // no password set
		DB:       0,     // use default DB
	})

	conf.Log.Debugf("the redis password is %s:", rdsPw)
	_, err = Rdb.Ping().Result()
	if err != nil {
		conf.Log.Errorf("rdb ping err")
		return err
	}
	err = GetConfFromRds()
	if err != nil {
		conf.Log.Errorf("rdb GetConfFromRds err is %s", err)
		return err
	}

	// 初始化基站缓存
	err = InitBaseStationCache()
	if err != nil {
		conf.Log.Errorf("rdb InitBaseStationCache err is %s", err)
		return err
	}

	// 启动Ping监控
	StartBaseStationMonitor()

	return nil
}

func GetConfFromRds() (err error) {
	if conf.CmnConf.OMSInfo.OMSIp != "" {
		conf.CmnConf.RdsInfo.RdsOmsIp = conf.CmnConf.OMSInfo.OMSIp
		Rdb.Set("omsIp", conf.CmnConf.OMSInfo.OMSIp, 0).Err()
	} else {
		conf.CmnConf.RdsInfo.RdsOmsIp = Rdb.Get("omsIp").Val()
	}
	if err != nil {
		return err
	}
	return nil
}

func InitBaseStationCache() error {
	CacheMutex.Lock()
	defer CacheMutex.Unlock()

	BaseStationCache = make(map[uint32]gnbTypes.BaseStation)

	// 使用 SCAN 遍历所有 bs: 开头的 key（推荐，防止 key 过多）
	iter := Rdb.Scan(0, "bs:*", 200).Iterator()

	for iter.Next() {
		key := iter.Val()
		data, err := Rdb.Get(key).Result()
		if err != nil {
			conf.Log.Warnf("获取基站数据失败 %s: %v", key, err)
			continue
		}

		var bs gnbTypes.BaseStation
		if err := json.Unmarshal([]byte(data), &bs); err != nil {
			conf.Log.Warnf("解析基站JSON失败 %s: %v", key, err)
			continue
		}

		BaseStationCache[bs.StationID] = bs
		conf.Log.Debugf("已加载基站: %d (%s) %s", bs.StationID, bs.IP, bs.Name)
	}

	if err := iter.Err(); err != nil {
		conf.Log.Errorf("Redis SCAN 出错: %v", err)
		return err
	}

	conf.Log.Infof("基站缓存初始化完成，共加载 %d 个基站", len(BaseStationCache))
	return nil
}
