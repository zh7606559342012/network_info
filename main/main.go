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

	go dbConn.DbConnInit()

	//modules func
	modules.Run()

	//server api start
	server.Start()
}

/*
type BaseStation struct {
	StationID uint32 `json:"station_id"`
	IP        string `json:"ip"`
	Name      string `json:"name,omitempty"` // 可选：站名
}

var (
	stations     = make(map[uint32]BaseStation)
	stationsLock sync.RWMutex
	logDir       = "./ping_logs"
	dataFile     = "base_stations.json"
)

func main() {
	os.MkdirAll(logDir, 0755)
	loadStations()

	// 启动所有 Ping 任务
	startAllPingTasks()

	r := gin.Default()

	// === API 路由 ===
	api := r.Group("/api")
	{
		api.GET("/stations", listStationsHandler)
		api.POST("/stations", addStationHandler)
		api.DELETE("/stations/:id", deleteStationHandler)
	}

	fmt.Println("🚀 基站 Ping 监控系统已启动")
	fmt.Println("HTTP 服务运行在 http://localhost:8080")

	r.Run(":8080") // 监听 8080 端口
}

// ====================== API Handlers ======================

type AddRequest struct {
	StationID uint32 `json:"station_id" binding:"required"`
	IP        string `json:"ip" binding:"required"`
	Name      string `json:"name"`
}

func addStationHandler(c *gin.Context) {
	var req AddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	stationsLock.Lock()
	defer stationsLock.Unlock()

	if _, exists := stations[req.StationID]; exists {
		c.JSON(409, gin.H{"error": "基站已存在"})
		return
	}

	stations[req.StationID] = BaseStation{
		StationID: req.StationID,
		IP:        req.IP,
		Name:      req.Name,
	}

	saveStations()
	go startPingTask(req.StationID, req.IP)

	c.JSON(200, gin.H{"message": "添加成功", "station": stations[req.StationID]})
}

func deleteStationHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	stationsLock.Lock()
	defer stationsLock.Unlock()

	if _, exists := stations[uint32(id)]; !exists {
		c.JSON(404, gin.H{"error": "基站不存在"})
		return
	}

	delete(stations, uint32(id))
	saveStations()

	c.JSON(200, gin.H{"message": "删除成功"})
}

func listStationsHandler(c *gin.Context) {
	stationsLock.RLock()
	defer stationsLock.RUnlock()

	var list []BaseStation
	for _, s := range stations {
		list = append(list, s)
	}

	c.JSON(200, gin.H{
		"total":    len(list),
		"stations": list,
	})
}

// ====================== Ping 相关（保持不变） ======================

func startAllPingTasks() {
	stationsLock.RLock()
	defer stationsLock.RUnlock()
	for id, bs := range stations {
		go startPingTask(id, bs.IP)
	}
}

func startPingTask(stationID uint32, ip string) {
	fmt.Printf("[站号 %d] Ping 任务已启动 → %s\n", stationID, ip)
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		pingOnce(stationID, ip)
	}
}

func pingOnce(stationID uint32, ip string) {
	start := time.Now()
	conn, err := net.DialTimeout("ip:icmp", ip, 5*time.Second)
	rtt := time.Since(start)

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	var result string

	if err != nil {
		result = fmt.Sprintf("%s | seq=000000 | TIMEOUT | rtt=0.0 ms", timestamp)
	} else {
		defer conn.Close()
		seq := time.Now().Unix() % 1000000
		result = fmt.Sprintf("%s | seq=%06d | OK   | rtt=%.1f ms", timestamp, seq, float64(rtt.Microseconds())/1000)
	}

	writePingLog(stationID, result)
}

func writePingLog(stationID uint32, line string) {
	dateStr := time.Now().Format("20060102")
	filename := fmt.Sprintf("%s/ping_%d_%s.log", logDir, stationID, dateStr)

	f, _ := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(line + "\n")
}

// ====================== 持久化（保持不变） ======================
*/

func saveStations() { /* ... 同之前代码 ... */ }
func loadStations() { /* ... 同之前代码 ... */ }
