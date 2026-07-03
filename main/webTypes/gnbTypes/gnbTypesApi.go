package gnbTypes

type BaseStation struct {
	StationID uint32 `json:"station_id"`
	IP        string `json:"ip"`
	Name      string `json:"name"`
	Region    string `json:"region"`
}
