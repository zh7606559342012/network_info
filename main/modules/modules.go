package modules

import "network_info/main/crontab"

func Run() {
	go crontab.Run()
}
