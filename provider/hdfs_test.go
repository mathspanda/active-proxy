package provider

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"active-proxy/provider/hadoop_hdfs"
	"active-proxy/provider/zk"
	"active-proxy/util"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

type mockPool struct {
	util.ProxyTaskPoolInterface
}

func (pool *mockPool) Push(request *http.Request) <-chan util.ProxyTaskResult {
	respChan := make(chan util.ProxyTaskResult, 1)
	respChan <- util.ProxyTaskResult{Resp: &http.Response{StatusCode: 200, Request: request}}
	return respChan
}

func (pool *mockPool) Do() {

}

func prepare() (*HdfsProxyProvider, *zk.ZKServer, error) {
	zkServer, err := zk.StartFatZkServer()
	if err != nil {
		return nil, nil, err
	}
	confMap := make(map[string]interface{})
	confMap[ZkServersConfKey] = fmt.Sprintf("%s:%d", zkServer.Address, zkServer.Port)
	confMap[ZkLockPathConfKey] = "/hadoop-ha"
	confMap[MaxConnectionsConfKey] = 16
	confMap[WebHdfsPortConfKey] = "50070"
	confMap[RequestTimeoutConfKey] = 1000
	provider, err := NewHdfsProxyProvider(ProviderConf(confMap))
	if err != nil {
		zkServer.Stop()
		return nil, nil, err
	}
	provider.Pool = &mockPool{}
	return provider, zkServer, nil
}

func marshalActiveNodeInfo(hostname string) []byte {
	nnInfo := &hadoop_hdfs.ActiveNodeInfo{Hostname: &hostname}
	data, _ := proto.Marshal(nnInfo)
	return data
}

func TestProviderStateTransformation(t *testing.T) {
	provider, zkServer, err := prepare()
	if err != nil {
		t.Error("TestProviderStateTransformation:", err.Error())
	}
	defer zkServer.Stop()

	zkAddress := fmt.Sprintf("%s:%d", zkServer.Address, zkServer.Port)
	zkClient, _ := zk.NewZKClient([]string{zkAddress}, 10)
	defer zkClient.Close()

	assert.Equal(t, INIT, provider.State)

	hostname := "localhost"
	nnInfo := marshalActiveNodeInfo(hostname)

	zkClient.Create(provider.zkLockPath, nnInfo)
	time.Sleep(time.Duration(3) * time.Second)
	assert.Equal(t, RUN, provider.State)
	assert.Equal(t, hostname, provider.activeNNAddress)

	zkClient.Delete(provider.zkLockPath)
	time.Sleep(time.Duration(1) * time.Second)
	assert.Equal(t, PEND, provider.State)

	zkClient.Create(provider.zkLockPath, nnInfo)
	time.Sleep(time.Duration(3) * time.Second)
	assert.Equal(t, RUN, provider.State)
}

func TestProviderProxy(t *testing.T) {
	provider, zkServer, err := prepare()
	if err != nil {
		t.Error("TestProviderProxy:", err.Error())
	}
	defer zkServer.Stop()

	zkAddress := fmt.Sprintf("%s:%d", zkServer.Address, zkServer.Port)
	zkClient, _ := zk.NewZKClient([]string{zkAddress}, 10)
	defer zkClient.Close()
	zkClient.Create(provider.zkLockPath, marshalActiveNodeInfo("localhost"))
	time.Sleep(time.Duration(3) * time.Second)

	response, err := provider.Proxy(&http.Request{Method: "GET"})
	if err != nil {
		t.Error("TestProviderProxy:", err.Error())
	}
	proxyUrlStr := fmt.Sprintf("localhost:%s", provider.Conf.GetString(WebHdfsPortConfKey))
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, proxyUrlStr, response.Request.Host)
}
