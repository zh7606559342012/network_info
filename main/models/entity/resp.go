package entity

type Response struct {
	Uid       string      `json:"uuid"`
	TimeStamp string      `json:"timestamp"`
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}
