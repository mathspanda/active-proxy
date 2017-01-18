package provider

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"active-proxy/provider/hadoop_hdfs"
	zkClient "active-proxy/provider/zk"
	"active-proxy/util"

	"github.com/golang/glog"
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
	ZkServersConfKey      = "HDFS_ZK_SERVERS"
	ZkLockPathConfKey     = "HDFS_ZK_LOCK_PATH"
	MaxConnectionsConfKey = "HDFS_MAX_CONNECTIONS"
	WebHdfsPortConfKey    = "HDFS_WEBHDFS_PORT"
	RequestTimeoutConfKey = "HDFS_REQUEST_TIMEOUT"
)

func NewHdfsProxyProvider(conf ProviderConf) (*HdfsProxyProvider, error) {
	provider := &HdfsProxyProvider{
		BaseProxyProvider: BaseProxyProvider{
			Conf:      conf,
			Type:      HDFS,
			State:     INIT,
			StateChan: make(chan ProviderState),
		},
		zkLockPath: conf.GetString(ZkLockPathConfKey),
	}
	provider.Pool, _ = util.NewProxyTaskPool(conf.GetInt(MaxConnectionsConfKey))
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
			glog.V(1).Infof("hdfs proxy provider: active namenode address changes from %s to %s.", provider.activeNNAddress, activeNNInfo.GetHostname())
			provider.activeNNAddress = activeNNInfo.GetHostname()
		}
		return true, ch
	}
	return false, ch
}

func (provider *HdfsProxyProvider) monitorZkLockPath() {
	zkServers := strings.Split(provider.Conf.GetString(ZkServersConfKey), ",")
	client, _ := zkClient.NewZKClient(zkServers, 1)
	success, ch := provider.resolveActiveNodeInfo(client)
	if success {
		provider.StateChan <- RUN
	}
	provider.initWg.Done()
	for {
		select {
		case e := <-ch:
			if e.Type == zk.EventNodeDeleted {
				provider.StateChan <- PEND
			}
			_, ch = provider.resolveActiveNodeInfo(client)

		case <-time.After(time.Duration(3) * time.Second):
			success, ch = provider.resolveActiveNodeInfo(client)
			provider.mutex.RLock()
			if success && provider.State != RUN {
				provider.StateChan <- RUN
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
		if provider.State != state {
			provider.State = state
		}
		provider.mutex.Unlock()
	}
}

func (provider *HdfsProxyProvider) Proxy(r *http.Request) (*http.Response, error) {
	provider.mutex.RLock()
	defer provider.mutex.RUnlock()

	if provider.State == PEND {
		return nil, errors.New("hdfs proxy provider not in service temporarily")
	}

	webhdfsPort := provider.Conf.GetString(WebHdfsPortConfKey)
	urlStr := fmt.Sprintf("%s://%s:%s%s", "http", provider.activeNNAddress, webhdfsPort, r.RequestURI)
	proxyReq, _ := NewProxyRequest(r, urlStr)
	select {
	case <-time.After(time.Millisecond * time.Duration(provider.Conf.GetInt(RequestTimeoutConfKey))):
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
	case RUN:
		stats.Explain = "hdfs proxy is in service"
	case PEND:
		stats.Explain = "perhaps namenode election is taking place, or all namenodes are dead"
	default:
		stats.Explain = "perhaps all namenodes are dead"
	}
	return stats
}
