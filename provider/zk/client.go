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
	conn.SetLogger(NilLogger{})

	client := &ZKClient{zkServers: zkServers, conn: conn}
	return client, nil
}

type NilLogger struct {
	zk.Logger
}

func (log NilLogger) Printf(string, ...interface{}) {

}

func (client *ZKClient) Close() {
	client.conn.Close()
}

func (client *ZKClient) GetW(zkPath string) ([]byte, <-chan zk.Event, error) {
	data, _, event, err := client.conn.GetW(zkPath)
	return data, event, err
}

func (client *ZKClient) Delete(zkPath string) error {
	return client.conn.Delete(zkPath, -1)
}

func (client *ZKClient) Create(zkPath string, value []byte) error {
	_, err := client.conn.Create(zkPath, value, 0, zk.WorldACL(zk.PermAll))
	return err
}
