package dbConn

import (
	"github.com/go-redis/redis"
	"network_info/main/conf"
)

var Rdb *redis.Client

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
		conf.Log.Errorf("rdb GetConfFromRds err")
		return err
	}

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
