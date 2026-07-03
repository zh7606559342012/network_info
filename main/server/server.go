package server

import "network_info/main/conf"

func Start() {
	r := NewRouter()
	routerAddr := conf.CmnConf.Appconf.Addr + ":" + conf.CmnConf.Appconf.Port
	err := r.Run(routerAddr)
	if err != nil {
		panic("Failed to start routes!")
	}
}
