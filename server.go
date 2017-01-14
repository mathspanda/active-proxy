package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"active-proxy/log"
	"active-proxy/middleware"
	"active-proxy/pool"
	. "active-proxy/provider"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/negroni"
	"gopkg.in/yaml.v2"
)

func main() {
	// parse flag
	var configFile string
	flag.StringVar(&configFile, "config", "config.yaml", "location of config file")
	flag.Parse()

	// init proxy conf
	conf, err := NewProxyConf(configFile)
	if err != nil {
		log.Error("Error init proxy configuration: ", err)
		return
	}

	// init logger
	level, _ := logrus.ParseLevel(strings.ToLower(conf.LogLevel))
	log.SetLevel(level)
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	log.Infof("ProxyConf: %+v", conf)

	// init proxy server
	proxyServer, _ := NewProxyServer(*conf)
	proxyServer.StartServer()
}

type ProxyServer struct {
	proxyConf            ProxyConf
	providers            map[ProviderType]ProxyProvider
	pools                map[ProviderType]*pool.ProxyTaskPool
	statisticsMiddleware *middleware.StatisticsMiddleware
}

func NewProxyServer(conf ProxyConf) (*ProxyServer, error) {
	server := &ProxyServer{proxyConf: conf}
	server.providers = make(map[ProviderType]ProxyProvider)
	server.pools = make(map[ProviderType]*pool.ProxyTaskPool)

	hdfsProvider, err := NewHdfsProxyProvider(conf.ProviderConfs[HDFS])
	server.providers[HDFS] = hdfsProvider
	server.pools[HDFS] = hdfsProvider.Pool

	server.statisticsMiddleware = middleware.NewStatisticsMiddleware(conf.RecentRequestNums)

	return server, err
}

func (server *ProxyServer) StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.DefaultHandler)
	mux.HandleFunc("/states", server.StatesHandler)
	mux.HandleFunc("/statistics", server.StatisticsHandler)

	n := negroni.New()
	n.Use(server.statisticsMiddleware)
	n.UseHandler(mux)

	http.ListenAndServe(server.proxyConf.ProxyServerPort, n)
}

func (server *ProxyServer) DefaultHandler(rw http.ResponseWriter, r *http.Request) {
	providerType := getProxyProviderType(r.URL.String())
	if providerType == DEFAULT {
		rw.WriteHeader(http.StatusNotFound)
		io.WriteString(rw, "cannot find corresponding proxy provider")
		return
	}
	for i := 0; i < server.proxyConf.RetryAttempts; i++ {
		resp, err := server.providers[providerType].Proxy(r)
		if err == nil && resp.StatusCode < 400 {
			io.WriteString(rw, convertResponseBody2String(resp))
			return
		}
		if i == server.proxyConf.RetryAttempts-1 {
			if resp != nil {
				rw.WriteHeader(resp.StatusCode)
				io.WriteString(rw, convertResponseBody2String(resp))
			} else {
				rw.WriteHeader(http.StatusBadRequest)
				io.WriteString(rw, err.Error())
			}
		} else {
			time.Sleep(time.Millisecond * time.Duration(server.proxyConf.RetryDelay))
		}
	}
	log.Warnf("Request %s still failed after retrying %d times.", r.RequestURI, server.proxyConf.RetryAttempts)
}

func convertResponseBody2String(response *http.Response) string {
	body, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	return string(body)
}

func (server *ProxyServer) StatesHandler(rw http.ResponseWriter, r *http.Request) {
	providerStats := make(map[string]ProviderStats)
	for providerType, provider := range server.providers {
		providerStats[providerType.String()] = provider.GetStats()
	}
	buf, _ := json.Marshal(providerStats)
	io.WriteString(rw, string(buf))
}

func (server *ProxyServer) StatisticsHandler(rw http.ResponseWriter, r *http.Request) {
	io.WriteString(rw, server.statisticsMiddleware.Json())
}

const (
	WEBHDFS_PREFIX = "/webhdfs/v1"
)

func getProxyProviderType(urlPath string) ProviderType {
	switch {
	case strings.HasPrefix(urlPath, WEBHDFS_PREFIX):
		return HDFS
	default:
		return DEFAULT
	}
}

type ProxyConf struct {
	GlobalConf
	ConfigFile string
}

type GlobalConf struct {
	LogLevel          string
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

	globalConf := convert2ProviderConf(m["GLOBAL"].(map[interface{}]interface{}))
	hdfsConf := convert2ProviderConf(m["HDFS"].(map[interface{}]interface{}))

	logLevel := globalConf.GetString("PROXY_LOG_LEVEL")
	proxyPort := globalConf.GetString("PROXY_SERVER_PORT")
	retryAttempts := globalConf.GetInt("PROXY_RETRY_ATTEMPTS")
	retryDelay := globalConf.GetInt("PROXY_RETRY_DELAY")
	recentRequestNums := globalConf.GetInt("PROXY_RECENT_REQUEST_NUMS")

	providerConfs := make(map[ProviderType]ProviderConf)
	providerConfs[HDFS] = hdfsConf

	return &ProxyConf{
		GlobalConf: GlobalConf{
			LogLevel:          logLevel,
			ProxyServerPort:   ":" + proxyPort,
			RetryAttempts:     retryAttempts,
			RetryDelay:        retryDelay,
			RecentRequestNums: recentRequestNums,
			ProviderConfs:     providerConfs,
		},
		ConfigFile: absFilePath,
	}, nil
}

func convert2ProviderConf(m map[interface{}]interface{}) ProviderConf {
	conf := make(map[string]interface{})
	for key, value := range m {
		conf[key.(string)] = value
	}
	return ProviderConf(conf)
}
