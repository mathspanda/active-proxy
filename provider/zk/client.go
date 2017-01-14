package zk

import (
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

type ZKClient struct {
	zkServers []string
	conn      *zk.Conn
}

func NewZKClient(zkServers []string, timeout int) (*ZKClient, error) {
	conn, _, err := zk.Connect(zkServers, time.Second*time.Duration(timeout))
	if err != nil {
		return nil, err
	}

	client := &ZKClient{zkServers: zkServers, conn: conn}
	return client, nil
}

func (client *ZKClient) Close() {
	client.conn.Close()
}

func (client *ZKClient) GetW(zkPath string) ([]byte, <-chan zk.Event, error) {
	data, _, event, err := client.conn.GetW(zkPath)
	return data, event, err
}
