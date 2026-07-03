package conf

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

const (
	AgentYamlFilePath = "/opt/monitor_agent/bin/conf/agent.yaml"
)

var (
	CmnConf *CommonConf
)

type CommonConf struct {
	Appconf AppConfig  `yaml:"app"`
	LogConf LogInfo    `yaml:"log"`
	DBConn  DBConnInfo `yaml:"db"`
	RdsInfo RedisInfo  `yaml:"-"`
	OMSInfo OMSInfo    `yaml:"oms"`
}

type AppConfig struct {
	Addr       string `yaml:"addr"`
	Port       string `yaml:"port"`
	Ver        string `yaml:"version"`
	Nfip       string `yaml:"Nfip"`
	TelnetAddr string `yaml:"telnetAddr"`
}

type LogInfo struct {
	LogPath  string `yaml:"logPath"`
	LogLevel string `yaml:"logLevel"`
}

type DBConnInfo struct {
	RedisInfo AddrInfo `yaml:"redis"`
}

type AddrInfo struct {
	Addr     string `yaml:"addr"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
}

type RedisInfo struct {
	RdsOmsIp string
}

type OMSInfo struct {
	OMSProto string `yaml:"omsProto"`
	OMSPort  string `yaml:"omsPort"`
	OMSIp    string `yaml:"omsIp"`
	OMSK8s   bool   `yaml:"omsK8s"`
}

func ConfInit(ver string) {
	LoadConf(ver)
}

func LoadConf(ver string) {
	data, err := ioutil.ReadFile(AgentYamlFilePath)
	if err != nil {
		Log.Errorf("read %s config file failed:%s.", AgentYamlFilePath, err)
		return
	}
	err = yaml.Unmarshal(data, &CmnConf)
	if err != nil {
		Log.Errorf("unmarshal data failed:%s.", err)
	}

	CmnConf.Appconf.Ver = ver
}
