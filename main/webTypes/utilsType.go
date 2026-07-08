package webTypes

import "time"

// HourlyRecord - 原始小时级日志记录
type HourlyRecord struct {
	StationID uint32    `json:"station_id"`
	Timestamp time.Time `json:"timestamp"`
	RTT       float64   `json:"rtt"` // 单位：ms
	IsOK      bool      `json:"is_ok"`
}

// DailyRecord - 按天聚合后的记录
type DailyRecord struct {
	StationID  uint32    `json:"station_id"`
	Date       time.Time `json:"date"`
	RTTMean    float64   `json:"rtt_mean"`
	RTTP95     float64   `json:"rtt_p95"`
	JitterMean float64   `json:"jitter_mean"` // 当前日志暂无，默认为0
	PacketLoss float64   `json:"packet_loss"` // 当前日志暂无，默认为0
	HourCount  int       `json:"hour_count"`
}

// AnomalyResult - 最终异常检测结果（最重要）
type AnomalyResult struct {
	StationID      uint32  `json:"station_id"`
	Date           string  `json:"date"`
	AnomalyScore   float64 `json:"anomaly_score"`
	AlertLevel     string  `json:"alert_level"`
	RTTMean        float64 `json:"rtt_mean"`
	BaselineRTT    float64 `json:"baseline_rtt_mean"`
	MaxChangeRatio float64 `json:"max_change_ratio"`
	RTTContrib     float64 `json:"rtt_contrib"`
	RTTP95Contrib  float64 `json:"rtt_p95_contrib"`
	JitterContrib  float64 `json:"jitter_contrib"`
	LossContrib    float64 `json:"loss_contrib"`
	DataPoints     int     `json:"data_points"`
	HistoryDays    int     `json:"history_days"`
}
