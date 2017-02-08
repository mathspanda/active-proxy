package server

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	. "active-proxy/provider"

	"gopkg.in/yaml.v2"
)

type ProxyConf struct {
	GlobalConf
	ConfigFile        string
	ProxyProviderType string
	ProxyProviderConf ProviderConf
}

type GlobalConf struct {
	ProxyServerPort   string
	RetryAttempts     int
	RetryDelay        int
	RecentRequestNums int
}

func NewProxyConf(providerType string, filePath string) (*ProxyConf, error) {
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

	// global config
	globalConf := convert2ProviderConf(m["GLOBAL"])
	proxyPort := globalConf.GetString("PROXY_SERVER_PORT")
	retryAttempts := globalConf.GetInt("PROXY_RETRY_ATTEMPTS")
	retryDelay := globalConf.GetInt("PROXY_RETRY_DELAY")
	recentRequestNums := globalConf.GetInt("PROXY_RECENT_REQUEST_NUMS")

	var providerConf ProviderConf
	if conf, ok := m[strings.ToUpper(providerType)]; ok {
		providerConf = convert2ProviderConf(conf)
	} else {
		return nil, fmt.Errorf("cannot find %s proxy provider configuration in file %s", providerType, filePath)
	}

	return &ProxyConf{
		GlobalConf: GlobalConf{
			ProxyServerPort:   ":" + proxyPort,
			RetryAttempts:     retryAttempts,
			RetryDelay:        retryDelay,
			RecentRequestNums: recentRequestNums,
		},
		ConfigFile:        absFilePath,
		ProxyProviderType: providerType,
		ProxyProviderConf: providerConf,
	}, nil
}

func convert2ProviderConf(m interface{}) ProviderConf {
	conf := make(map[string]interface{})
	for key, value := range m.(map[interface{}]interface{}) {
		conf[key.(string)] = value
	}
	return ProviderConf(conf)
}
