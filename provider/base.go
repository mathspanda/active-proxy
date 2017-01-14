package provider

import (
	"net/http"
	"os"
	"strconv"

	"active-proxy/pool"
)

type ProviderState int

const (
	INIT = ProviderState(iota)
	START
	STOP
)

func (state ProviderState) String() string {
	switch state {
	case INIT:
		return "init"
	case START:
		return "start"
	case STOP:
		return "stop"
	default:
		return "unknown"
	}
}

type ProviderType int

const (
	HDFS = ProviderType(iota)
	DEFAULT
)

func (providerType ProviderType) String() string {
	switch providerType {
	case HDFS:
		return "hdfs_proxy_provider"
	default:
		return "unknown_proxy_provider"
	}
}

type ProviderConf map[string]interface{}

func (providerConf ProviderConf) GetInt(key string) int {
	envVal := os.Getenv(key)
	if len(envVal) > 0 {
		intVal, err := strconv.Atoi(envVal)
		if err == nil {
			return intVal
		}
	}
	return providerConf[key].(int)
}

func (providerConf ProviderConf) GetString(key string) string {
	envVal := os.Getenv(key)
	if len(envVal) > 0 {
		return envVal
	}
	return providerConf[key].(string)
}

type ProviderStats struct {
	State   string `json:"provider_state"`
	Explain string `json:"state_explanation"`
}

// ProxyProvider defines methods of a provider
type ProxyProvider interface {
	Proxy(r *http.Request) (*http.Response, error)
	GetStats() ProviderStats
}

// BaseProxyProvider should be inherited by providers
type BaseProxyProvider struct {
	Pool *pool.ProxyTaskPool

	Conf      ProviderConf
	State     ProviderState
	StateChan chan ProviderState
	Type      ProviderType
}

func NewProxyRequest(req *http.Request, urlStr string) (*http.Request, error) {
	proxyReq, err := http.NewRequest(req.Method, urlStr, req.Body)
	if err != nil {
		return nil, err
	}
	proxyReq.Header = req.Header
	return proxyReq, nil
}
