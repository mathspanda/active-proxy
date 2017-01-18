package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"testing"

	"active-proxy/middleware"
	. "active-proxy/provider"

	"github.com/stretchr/testify/assert"
)

type mockHDFSProxyProvider struct {
	ProxyProvider
}

func (provider *mockHDFSProxyProvider) Proxy(request *http.Request) (*http.Response, error) {
	if request.Method == "GET" {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(request.RequestURI)),
		}, nil
	}
	return &http.Response{StatusCode: http.StatusMethodNotAllowed}, nil
}

func (provider *mockHDFSProxyProvider) GetStats() ProviderStats {
	return ProviderStats{}
}

var server *ProxyServer

func prepare() {
	if server == nil {
		conf := ProxyConf{
			GlobalConf: GlobalConf{
				ProxyServerPort:   ":8080",
				RetryAttempts:     5,
				RetryDelay:        500,
				RecentRequestNums: 10,
			},
		}
		server = &ProxyServer{proxyConf: conf}
		server.providers = make(map[ProviderType]ProxyProvider)
		server.providers[HDFS] = &mockHDFSProxyProvider{}
		server.statisticsMiddleware = middleware.NewStatisticsMiddleware(conf.RecentRequestNums)
		go server.StartServer()
		runtime.Gosched()
	}
}

const (
	DefaultUrl = "http://localhost:8080/default"
	HdfsUrl    = "http://localhost:8080/webhdfs/v1"
)

func TestDefaultHandler(t *testing.T) {
	prepare()

	client := http.Client{}
	request, _ := http.NewRequest("GET", DefaultUrl, nil)
	resp, _ := client.Do(request)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()

	request, _ = http.NewRequest("GET", HdfsUrl, nil)
	resp, _ = client.Do(request)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "/webhdfs/v1", convertResponseBody2String(resp))
	resp.Body.Close()

	request, _ = http.NewRequest("POST", HdfsUrl, nil)
	resp, _ = client.Do(request)
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	resp.Body.Close()
}

func TestStatesHandler(t *testing.T) {
	prepare()

	client := http.Client{}
	request, _ := http.NewRequest("GET", "http://localhost:8080/states", nil)
	resp, _ := client.Do(request)
	respData, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	statesMap := make(map[string]ProviderStats)
	json.Unmarshal(respData, &statesMap)
	assert.Equal(t, len(server.providers), len(statesMap))
	for providerType, provider := range server.providers {
		if providerState, ok := statesMap[providerType.String()]; ok {
			assert.Equal(t, provider.GetStats(), providerState)
		}
	}
}

func TestStatisticsHandler(t *testing.T) {
	prepare()

	client := http.Client{}
	request, _ := http.NewRequest("GET", DefaultUrl, nil)
	client.Do(request)
	request, _ = http.NewRequest("GET", HdfsUrl, nil)
	client.Do(request)
	request, _ = http.NewRequest("POST", HdfsUrl, nil)
	client.Do(request)

	request, _ = http.NewRequest("GET", "http://localhost:8080/statistics", nil)
	resp, _ := client.Do(request)
	respData, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	statisticsMap := make(map[string]interface{})
	json.Unmarshal(respData, &statisticsMap)

	if stats, ok := statisticsMap["recentRequests"]; ok {
		statsSlice := convert2RequestsRecords(stats.([]interface{}))
		statsLen := len(statsSlice)
		if statsLen < 3 {
			t.Error("TestStatisticsHandler:", "less recent requests")
		}
		assert.Equal(t, http.StatusMethodNotAllowed, statsSlice[statsLen-1].StatusCode)
		assert.Equal(t, http.StatusOK, statsSlice[statsLen-2].StatusCode)
		assert.Equal(t, http.StatusNotFound, statsSlice[statsLen-3].StatusCode)
	} else {
		t.Error("TestStatisticsHandler:", "lack recent requests")
	}
}

func convert2RequestsRecords(stats []interface{}) []middleware.RequestsRecord {
	statsSlice := make([]middleware.RequestsRecord, len(stats))
	for i, statI := range stats {
		stat := statI.(map[string]interface{})
		statsSlice[i] = middleware.RequestsRecord{
			Method:     stat["method"].(string),
			Host:       stat["host"].(string),
			Path:       stat["path"].(string),
			StatusCode: int(stat["status_code"].(float64)),
			Status:     stat["status"].(string),
		}
	}
	return statsSlice
}
