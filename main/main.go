package main

import (
	"network_info/main/conf"
	"network_info/main/dbConn"
	"network_info/main/modules"
	"network_info/main/server"
)

var Version string

func main() {
	conf.ConfInit(Version)
	conf.LoggerInit()
	conf.Log.Infof("######monitor_agent start,version:%s#########", conf.CmnConf.Appconf.Ver)
	dbConn.DbConnInit()

	//modules func
	modules.Run()

	//server api start
	server.Start()
}
