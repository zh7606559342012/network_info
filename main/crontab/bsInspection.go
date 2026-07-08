package crontab

import (
	"bufio"
	"fmt"
	"math"
	"network_info/main/conf"
	"network_info/main/dbConn"
	"network_info/main/webTypes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ====================== 异常检测主函数 ======================
func detectNetworkAnomaly() {
	start := time.Now()
	conf.Log.Info("=== 开始执行基站网络异常检测任务 ===")

	rawData := fetchLast30DaysData()
	if len(rawData) == 0 {
		conf.Log.Warn("未找到任何日志数据，跳过检测")
		return
	}

	results := calculateAnomaly(rawData)

	// 输出告警
	alertCount := 0
	for _, r := range results {
		if r.AnomalyScore >= 50 {
			alertCount++
			conf.Log.Warnf("【异常】基站 %d | 分数 %.1f | %s | 当前RTT %.2fms | 基线 %.2fms | 突变 %.1f%%",
				r.StationID, r.AnomalyScore, r.AlertLevel,
				r.RTTMean, r.BaselineRTT, r.MaxChangeRatio)
		}
	}

	conf.Log.Infof("异常检测完成！共检测 %d 个基站，发现 %d 个异常/关注，耗时 %v",
		len(results), alertCount, time.Since(start))

	saveAnomalyResults(results)
}

// ====================== 计算异常核心逻辑 ======================
func calculateAnomaly(rawData []webTypes.HourlyRecord) []webTypes.AnomalyResult {
	dailyMap := aggregateToDaily(rawData)

	var results []webTypes.AnomalyResult
	for stationID, days := range dailyMap {
		if len(days) == 0 {
			continue
		}
		res := computeSingleStation(stationID, days)
		results = append(results, res)
	}

	// 按分数降序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].AnomalyScore > results[j].AnomalyScore
	})

	return results
}

// 小时数据 → 每天聚合
func aggregateToDaily(records []webTypes.HourlyRecord) map[uint32][]webTypes.DailyRecord {
	dailyMap := make(map[uint32]map[string]webTypes.DailyRecord)

	for _, r := range records {
		if !r.IsOK || r.RTT <= 0 {
			continue
		}

		dateStr := r.Timestamp.Format("2006-01-02")

		if _, ok := dailyMap[r.StationID]; !ok {
			dailyMap[r.StationID] = make(map[string]webTypes.DailyRecord)
		}

		m := dailyMap[r.StationID]
		if existing, ok := m[dateStr]; ok {
			existing.RTTMean = (existing.RTTMean*float64(existing.HourCount) + r.RTT) / float64(existing.HourCount+1)
			existing.RTTP95 = math.Max(existing.RTTP95, r.RTT)
			existing.HourCount++
			m[dateStr] = existing
		} else {
			m[dateStr] = webTypes.DailyRecord{
				StationID: r.StationID,
				Date:      r.Timestamp,
				RTTMean:   r.RTT,
				RTTP95:    r.RTT,
				HourCount: 1,
			}
		}
	}

	// 转有序 slice
	result := make(map[uint32][]webTypes.DailyRecord)
	for sid, m := range dailyMap {
		var list []webTypes.DailyRecord
		for _, d := range m {
			list = append(list, d)
		}
		sort.Slice(list, func(i, j int) bool {
			return list[i].Date.Before(list[j].Date)
		})
		result[sid] = list
	}
	return result
}

// 单个基站计算
func computeSingleStation(stationID uint32, days []webTypes.DailyRecord) webTypes.AnomalyResult {
	current := days[len(days)-1]
	histStart := len(days) - 1 - 30
	if histStart < 0 {
		histStart = 0
	}
	history := days[histStart : len(days)-1]

	if len(history) < 7 {
		return webTypes.AnomalyResult{
			StationID:    stationID,
			Date:         current.Date.Format("2006-01-02"),
			AnomalyScore: 0,
			AlertLevel:   "数据不足",
			RTTMean:      current.RTTMean,
			DataPoints:   current.HourCount,
			HistoryDays:  len(history),
		}
	}

	// 参数
	const (
		AnomalyThreshold = 70.0
		WarningThreshold = 50.0
		ChangeThreshold  = 0.30
	)

	weights := map[string]float64{"rtt_mean": 0.25, "rtt_p95": 0.25}

	totalScore := 0.0
	maxChange := 0.0
	contribs := make(map[string]float64)

	features := []string{"rtt_mean", "rtt_p95"}
	for _, feat := range features {
		var vals []float64
		for _, d := range history {
			if feat == "rtt_mean" {
				vals = append(vals, d.RTTMean)
			} else {
				vals = append(vals, d.RTTP95)
			}
		}

		histMean := mean(vals)
		histStd := stdDev(vals)
		histP95 := percentile95(vals)

		var curVal float64
		if feat == "rtt_mean" {
			curVal = current.RTTMean
		} else {
			curVal = current.RTTP95
		}

		change := math.Abs(curVal-histMean) / (histMean + 1e-6)
		if feat == "rtt_mean" {
			maxChange = change
		}

		z := 0.0
		if histStd > 0 {
			z = math.Abs(curVal-histMean) / histStd
		}
		percentScore := math.Max(0, (curVal-histP95)/(histP95+1e-6))

		score := math.Min(100, z*15+percentScore*40)
		contribs[feat] = math.Round(score*100) / 100
		totalScore += score * weights[feat]
	}

	score := math.Min(100, math.Round(totalScore*10)/10)

	var alert string
	if maxChange > ChangeThreshold && score >= WarningThreshold {
		alert = "🚨 重大突变"
	} else if score >= AnomalyThreshold {
		alert = "⚠️ 告警"
	} else if score >= WarningThreshold {
		alert = "⚠️ 关注"
	} else {
		alert = "正常"
	}

	baseline := meanFunc(history, func(d webTypes.DailyRecord) float64 { return d.RTTMean })

	return webTypes.AnomalyResult{
		StationID:      stationID,
		Date:           current.Date.Format("2006-01-02"),
		AnomalyScore:   score,
		AlertLevel:     alert,
		RTTMean:        math.Round(current.RTTMean*100) / 100,
		BaselineRTT:    math.Round(baseline*100) / 100,
		MaxChangeRatio: math.Round(maxChange*1000) / 10,
		RTTContrib:     contribs["rtt_mean"],
		RTTP95Contrib:  contribs["rtt_p95"],
		JitterContrib:  0,
		LossContrib:    0,
		DataPoints:     current.HourCount,
		HistoryDays:    len(history),
	}
}

// ====================== 辅助函数 ======================
func mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func stdDev(vals []float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	m := mean(vals)
	sum := 0.0
	for _, v := range vals {
		sum += (v - m) * (v - m)
	}
	return math.Sqrt(sum / float64(len(vals)-1))
}

func percentile95(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	sort.Float64s(sorted)
	idx := int(0.95 * float64(len(sorted)-1))
	return sorted[idx]
}

func meanFunc(days []webTypes.DailyRecord, f func(webTypes.DailyRecord) float64) float64 {
	if len(days) == 0 {
		return 0
	}
	sum := 0.0
	for _, d := range days {
		sum += f(d)
	}
	return sum / float64(len(days))
}

// ====================== 保存结果（可后续扩展） ======================
func saveAnomalyResults(results []webTypes.AnomalyResult) {
	// 目前只打印，后面可以改成写CSV或入库
	filename := fmt.Sprintf("anomaly_%s.csv", time.Now().Format("20060102_1504"))
	// TODO: 实现CSV写入...
	conf.Log.Infof("检测结果已生成（可保存为 %s）", filename)
}

// fetchLast30DaysData 读取最近30天所有基站的日志文件
func fetchLast30DaysData() []webTypes.HourlyRecord {
	start := time.Now()
	var records []webTypes.HourlyRecord

	cutoff := time.Now().AddDate(0, 0, -30) // 30天前

	dbConn.CacheMutex.RLock()
	defer dbConn.CacheMutex.RUnlock()

	for stationID, bs := range dbConn.BaseStationCache {
		logFile := filepath.Join("/var/log/monitor_agent", fmt.Sprintf("ping_%s.log", strings.ReplaceAll(bs.IP, ".", "")))

		fileRecords := parseLogFile(logFile, stationID, cutoff)
		records = append(records, fileRecords...)
	}

	conf.Log.Infof("日志解析完成，共读取 %d 条记录，耗时 %v", len(records), time.Since(start))
	return records
}

// 解析单个日志文件
func parseLogFile(filename string, stationID uint32, cutoff time.Time) []webTypes.HourlyRecord {
	var records []webTypes.HourlyRecord

	file, err := os.Open(filename)
	if err != nil {
		// conf.Log.Debugf("日志文件不存在: %s", filename) // 大部分基站可能没日志，安静处理
		return nil
	}
	defer file.Close()

	re := regexp.MustCompile(`(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) \| seq=.* \| (\w+) \| rtt=([\d.]+) ms`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) != 4 {
			continue
		}

		tsStr := matches[1]
		status := matches[2]
		rttStr := matches[3]

		ts, err := time.Parse("2006-01-02 15:04:05", tsStr)
		if err != nil || ts.Before(cutoff) {
			continue
		}

		rtt, _ := strconv.ParseFloat(rttStr, 64)

		records = append(records, webTypes.HourlyRecord{
			StationID: stationID,
			Timestamp: ts,
			RTT:       rtt,
			IsOK:      status == "OK",
		})
	}

	return records
}
