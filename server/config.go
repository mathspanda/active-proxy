package server

import (
	"io/ioutil"
	"path/filepath"

	. "active-proxy/provider"

	"gopkg.in/yaml.v2"
)

type ProxyConf struct {
	GlobalConf
	ConfigFile string
}

type GlobalConf struct {
	ProxyServerPort   string
	RetryAttempts     int
	RetryDelay        int
	RecentRequestNums int

	ProviderConfs map[ProviderType]ProviderConf
}

func NewProxyConf(filePath string) (*ProxyConf, error) {
	absFilePath, _ := filepath.Abs(filePath)
	data, err := ioutil.ReadFile(absFilePath)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	globalConf := convert2ProviderConf(m["GLOBAL"])
	hdfsConf := convert2ProviderConf(m["HDFS"])

	proxyPort := globalConf.GetString("PROXY_SERVER_PORT")
	retryAttempts := globalConf.GetInt("PROXY_RETRY_ATTEMPTS")
	retryDelay := globalConf.GetInt("PROXY_RETRY_DELAY")
	recentRequestNums := globalConf.GetInt("PROXY_RECENT_REQUEST_NUMS")

	providerConfs := make(map[ProviderType]ProviderConf)
	providerConfs[HDFS] = hdfsConf

	return &ProxyConf{
		GlobalConf: GlobalConf{
			ProxyServerPort:   ":" + proxyPort,
			RetryAttempts:     retryAttempts,
			RetryDelay:        retryDelay,
			RecentRequestNums: recentRequestNums,
			ProviderConfs:     providerConfs,
		},
		ConfigFile: absFilePath,
	}, nil
}

func convert2ProviderConf(m interface{}) ProviderConf {
	conf := make(map[string]interface{})
	for key, value := range m.(map[interface{}]interface{}) {
		conf[key.(string)] = value
	}
	return ProviderConf(conf)
}
