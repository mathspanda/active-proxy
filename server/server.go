package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"active-proxy/middleware"
	. "active-proxy/provider"
	"active-proxy/util"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

type ProxyServer struct {
	proxyConf            ProxyConf
	providers            map[ProviderType]ProxyProvider
	pools                map[ProviderType]util.ProxyTaskPoolInterface
	statisticsMiddleware *middleware.StatisticsMiddleware
}

func NewProxyServer(conf ProxyConf) (*ProxyServer, error) {
	if len(conf.ProviderConfs) == 0 {
		return nil, fmt.Errorf("none proxy providers configured")
	}

	server := &ProxyServer{proxyConf: conf}
	server.providers = make(map[ProviderType]ProxyProvider)
	server.pools = make(map[ProviderType]util.ProxyTaskPoolInterface)

	// config hdfs proxy provider
	if hdfsConf, ok := conf.ProviderConfs[HDFS]; ok {
		hdfsProvider, err := NewHdfsProxyProvider(hdfsConf)
		if err != nil {
			return nil, err
		}
		server.providers[HDFS] = hdfsProvider
		server.pools[HDFS] = hdfsProvider.Pool
	}

	server.statisticsMiddleware = middleware.NewStatisticsMiddleware(conf.RecentRequestNums)

	return server, nil
}

func (server *ProxyServer) StartServer() {
	router := mux.NewRouter()
	router.PathPrefix("/states").HandlerFunc(server.StatesHandler)
	router.PathPrefix("/statistics").HandlerFunc(server.StatisticsHandler)

	defaultRouter := mux.NewRouter()
	defaultRouter.PathPrefix("/").HandlerFunc(server.DefaultHandler)

	// specific middleware for default handler
	router.PathPrefix("/").Handler(negroni.New(
		server.statisticsMiddleware,
		negroni.Wrap(defaultRouter),
	))

	http.ListenAndServe(server.proxyConf.ProxyServerPort, router)
}

func (server *ProxyServer) DefaultHandler(rw http.ResponseWriter, r *http.Request) {
	providerType := getProxyProviderType(r.URL.String())
	if providerType == DEFAULT {
		http.Error(rw, "cannot find corresponding proxy provider", http.StatusNotFound)
		return
	}
	for i := 0; i < server.proxyConf.RetryAttempts; i++ {
		resp, err := server.providers[providerType].Proxy(r)

		// good request
		if err == nil && resp.StatusCode < 400 {
			io.WriteString(rw, convertResponseBody2String(resp))
			return
		}

		var errorMsg string
		var statusCode int
		if resp != nil {
			errorMsg = convertResponseBody2String(resp)
			statusCode = resp.StatusCode
		} else {
			errorMsg = err.Error()
			statusCode = http.StatusBadRequest
		}

		// bad request
		if i == server.proxyConf.RetryAttempts-1 {
			glog.V(1).Infof("Request %s still fails after retrying %d times: %s", r.URL.String(), i+1, errorMsg)
			http.Error(rw, errorMsg, statusCode)
		} else {
			glog.V(3).Infof("Request %s fails at %d/%d times: %s", r.URL.String(), i+1, server.proxyConf.RetryAttempts, errorMsg)
			time.Sleep(time.Millisecond * time.Duration(server.proxyConf.RetryDelay))
		}

		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}
}

func convertResponseBody2String(response *http.Response) string {
	if response.Body != nil {
		body, _ := ioutil.ReadAll(response.Body)
		return string(body)
	}
	return ""
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

// decide proxy provider type according to some rules
func getProxyProviderType(urlPath string) ProviderType {
	switch {
	case strings.HasPrefix(urlPath, WEBHDFS_PREFIX):
		return HDFS
	default:
		return DEFAULT
	}
}
