package provider

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"active-proxy/log"
	"active-proxy/pool"
	"active-proxy/provider/hadoop_hdfs"
	zkClient "active-proxy/provider/zk"

	"github.com/golang/protobuf/proto"
	"github.com/samuel/go-zookeeper/zk"
)

type HdfsProxyProvider struct {
	BaseProxyProvider
	activeNNAddress string `description:"active namenode address"`
	zkLockPath      string `description:"zkPath which contains active namenode info"`

	initWg sync.WaitGroup
	mutex  sync.RWMutex
}

const (
	ZK_SERVERS_CONF_KEY      = "HDFS_ZK_SERVERS"
	ZK_LOCK_PATH_CONF_KEY    = "HDFS_ZK_LOCK_PATH"
	MAX_CONNECTIONS_CONF_KEY = "HDFS_MAX_CONNECTIONS"
	WEBHDFS_PORT_CONF_KEY    = "HDFS_WEBHDFS_PORT"
	REQUEST_TIMEOUT_CONF_KEY = "HDFS_REQUEST_TIMEOUT"
)

func NewHdfsProxyProvider(conf ProviderConf) (*HdfsProxyProvider, error) {
	provider := &HdfsProxyProvider{
		BaseProxyProvider: BaseProxyProvider{
			Conf:      conf,
			Type:      HDFS,
			State:     INIT,
			StateChan: make(chan ProviderState),
		},
		zkLockPath: conf.GetString(ZK_LOCK_PATH_CONF_KEY),
	}
	provider.Pool, _ = pool.NewProxyTaskPool(conf.GetInt(MAX_CONNECTIONS_CONF_KEY))
	go provider.Pool.Do()

	provider.initWg.Add(2)
	go provider.monitorZkLockPath()
	go provider.monitorProviderState()
	// wait until two monitor goroutines finish initialization
	provider.initWg.Wait()

	return provider, nil
}

func (provider *HdfsProxyProvider) resolveActiveNodeInfo(client *zkClient.ZKClient) (bool, <-chan zk.Event) {
	data, ch, _ := client.GetW(provider.zkLockPath)
	provider.mutex.Lock()
	defer provider.mutex.Unlock()

	if data != nil && len(data) != 0 {
		activeNNInfo := &hadoop_hdfs.ActiveNodeInfo{}
		proto.Unmarshal(data, activeNNInfo)
		if provider.activeNNAddress != activeNNInfo.GetHostname() {
			log.Infof("hdfs proxy provider: active namenode address changes from %s to %s.", provider.activeNNAddress, activeNNInfo.GetHostname())
			provider.activeNNAddress = activeNNInfo.GetHostname()
		}
		return true, ch
	}
	return false, ch
}

func (provider *HdfsProxyProvider) monitorZkLockPath() {
	zkServers := strings.Split(provider.Conf.GetString(ZK_SERVERS_CONF_KEY), ",")
	client, _ := zkClient.NewZKClient(zkServers, 1)
	success, ch := provider.resolveActiveNodeInfo(client)
	if success {
		provider.StateChan <- START
	}
	provider.initWg.Done()
	for {
		select {
		case e := <-ch:
			if e.Type == zk.EventNodeDeleted {
				provider.StateChan <- STOP
			}
			_, ch = provider.resolveActiveNodeInfo(client)

		case <-time.After(time.Duration(3) * time.Second):
			success, ch = provider.resolveActiveNodeInfo(client)
			provider.mutex.RLock()
			if success && provider.State != START {
				provider.StateChan <- START
			}
			provider.mutex.RUnlock()
		}
	}
}

func (provider *HdfsProxyProvider) monitorProviderState() {
	provider.initWg.Done()
	for {
		state := <-provider.StateChan
		provider.mutex.Lock()
		switch state {
		case STOP:
			provider.State = STOP
		case START:
			provider.State = START
		}
		provider.mutex.Unlock()
	}
}

func (provider *HdfsProxyProvider) Proxy(r *http.Request) (*http.Response, error) {
	provider.mutex.RLock()
	defer provider.mutex.RUnlock()

	if provider.State == STOP {
		return nil, errors.New("hdfs proxy provider not in service temporarily")
	}

	webhdfsPort := provider.Conf.GetString(WEBHDFS_PORT_CONF_KEY)
	urlStr := fmt.Sprintf("%s://%s:%s%s", "http", provider.activeNNAddress, webhdfsPort, r.RequestURI)
	proxyReq, _ := NewProxyRequest(r, urlStr)
	select {
	case <-time.After(time.Millisecond * time.Duration(provider.Conf.GetInt(REQUEST_TIMEOUT_CONF_KEY))):
		return nil, errors.New(fmt.Sprintf("request %s timeout", r.RequestURI))

	case resp := <-provider.Pool.Push(proxyReq):
		return resp.Resp, resp.Error
	}

	return nil, errors.New(fmt.Sprintf("request %s failed", r.RequestURI))
}

func (provider *HdfsProxyProvider) GetStats() ProviderStats {
	provider.mutex.RLock()
	defer provider.mutex.RUnlock()

	stats := ProviderStats{State: provider.State.String()}
	switch provider.State {
	case START:
		stats.Explain = "hdfs proxy is in service"
	case STOP:
		stats.Explain = "maybe active namenode election is taking place"
	default:
		stats.Explain = "do not know what happened"
	}
	return stats
}
