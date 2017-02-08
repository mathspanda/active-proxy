package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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
	provider             ProxyProvider
	pool                 util.ProxyTaskPoolInterface
	statisticsMiddleware *middleware.StatisticsMiddleware
}

func NewProxyServer(conf ProxyConf) (*ProxyServer, error) {
	server := &ProxyServer{proxyConf: conf}

	switch conf.ProxyProviderType {
	case "hdfs":
		hdfsProvider, err := NewHdfsProxyProvider(conf.ProxyProviderConf)
		if err != nil {
			return nil, err
		}
		server.provider = hdfsProvider
		server.pool = hdfsProvider.Pool
	default:
		return nil, fmt.Errorf("invalid proxy provider: %s", conf.ProxyProviderType)
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
	for i := 0; i < server.proxyConf.RetryAttempts; i++ {
		resp, err := server.provider.Proxy(r)

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
	io.WriteString(rw, server.provider.GetStats().Json())
}

func (server *ProxyServer) StatisticsHandler(rw http.ResponseWriter, r *http.Request) {
	io.WriteString(rw, server.statisticsMiddleware.Json())
}
