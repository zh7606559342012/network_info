package dbConn

import (
	"fmt"
	"network_info/main/conf"
	"network_info/main/webTypes/gnbTypes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// StartBaseStationMonitor 启动每分钟Ping监控
func StartBaseStationMonitor() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		seq := uint64(1) // 全局序列号

		for range ticker.C {
			conf.Log.Infof("开始执行基站Ping任务... 当前缓存数量: %d", len(BaseStationCache))

			CacheMutex.RLock()
			for _, bs := range BaseStationCache {
				go pingAndLog(bs, &seq) // 并发Ping，提高效率
			}
			CacheMutex.RUnlock()
		}
	}()
}

// 执行单次Ping并写日志
func pingAndLog(bs gnbTypes.BaseStation, seq *uint64) {
	start := time.Now()
	ip := strings.TrimSpace(bs.IP)

	cmd := exec.Command("ping", "-c", "1", "-W", "2", "-n", ip)
	output, err := cmd.CombinedOutput()

	status := "FAIL"
	rttMs := 0.0

	if err == nil {
		status = "OK"
		outputStr := string(output)

		// 改进的 rtt 提取逻辑
		if idx := strings.LastIndex(outputStr, "time="); idx != -1 {
			// 示例输出: time=0.144 ms
			part := outputStr[idx+5:] // 跳过 "time="
			if end := strings.Index(part, " ms"); end != -1 {
				if val, parseErr := strconv.ParseFloat(strings.TrimSpace(part[:end]), 64); parseErr == nil {
					rttMs = val
				}
			}
		} else {
			// 兜底：使用命令执行时间
			rttMs = float64(time.Since(start).Milliseconds())
		}
	}

	// 日志文件名和写入
	logFileName := fmt.Sprintf("ping_%s.log", strings.ReplaceAll(ip, ".", ""))
	logDir := "/var/log/monitor_agent"

	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, logFileName)

	logLine := fmt.Sprintf("%s | seq=%06d | %-4s | rtt=%.2f ms\n",
		time.Now().Format("2006-01-02 15:04:05"),
		*seq,
		status,
		rttMs,
	)

	f, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(logLine)

	*seq++
}
