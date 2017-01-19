package server

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"active-proxy/middleware"
	. "active-proxy/provider"
	"active-proxy/util"

	"github.com/golang/glog"
	"github.com/urfave/negroni"
)

type ProxyServer struct {
	proxyConf            ProxyConf
	providers            map[ProviderType]ProxyProvider
	pools                map[ProviderType]util.ProxyTaskPoolInterface
	statisticsMiddleware *middleware.StatisticsMiddleware
}

func NewProxyServer(conf ProxyConf) (*ProxyServer, error) {
	server := &ProxyServer{proxyConf: conf}
	server.providers = make(map[ProviderType]ProxyProvider)
	server.pools = make(map[ProviderType]util.ProxyTaskPoolInterface)

	hdfsProvider, err := NewHdfsProxyProvider(conf.ProviderConfs[HDFS])
	if err != nil {
		return nil, err
	}
	server.providers[HDFS] = hdfsProvider
	server.pools[HDFS] = hdfsProvider.Pool

	server.statisticsMiddleware = middleware.NewStatisticsMiddleware(conf.RecentRequestNums)

	return server, nil
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
		// good request
		if err == nil && resp.StatusCode < 400 {
			io.WriteString(rw, convertResponseBody2String(resp))
			return
		}
		// bad request
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
	glog.V(2).Infof("Request %s still failed after retrying %d times.", r.RequestURI, server.proxyConf.RetryAttempts)
}

func convertResponseBody2String(response *http.Response) string {
	if response.Body != nil {
		body, _ := ioutil.ReadAll(response.Body)
		defer response.Body.Close()
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
