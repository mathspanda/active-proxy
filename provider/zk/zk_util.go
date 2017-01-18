package zk

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type ZKServer struct {
	Address     string
	Port        int
	TempDataDir string
	Cmd         *exec.Cmd
}

// start single fat zookeeper server, temporarily just for test
func StartFatZkServer() (*ZKServer, error) {
	randPort := int(rand.Int31n(5000) + 10000)
	tmpPath, err := ioutil.TempDir("", "gozk")
	if err != nil {
		return nil, err
	}

	// write zoo.cfg
	cfgPath := filepath.Join(tmpPath, "zoo.cfg")
	cfgFile, err := os.Create(cfgPath)
	if err != nil {
		return nil, err
	}
	zc := &ZKConfig{
		DataDir:    tmpPath,
		ClientPort: randPort,
		Id:         1,
	}
	zc.writeZooConfig(cfgFile)
	cfgFile.Close()
	// write myid
	idPath := filepath.Join(tmpPath, "myid")
	idFile, err := os.Create(idPath)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(idFile, "%d\n", 1)
	idFile.Close()

	jarFilePath := findFatJarPath("zookeeper-fatjar.jar")
	if jarFilePath == "" {
		return nil, fmt.Errorf("unable to find zookeeper fat jar")
	}
	cmd := exec.Command("java", "-jar", jarFilePath, "server", cfgPath)
	cmd.Start()
	time.Sleep(time.Duration(1) * time.Second)

	server := &ZKServer{
		Address:     "localhost",
		Port:        randPort,
		TempDataDir: tmpPath,
		Cmd:         cmd,
	}
	if waitResult := server.waitForStart(1000, 5, 500); waitResult {
		return server, nil
	}
	server.Stop()
	return nil, fmt.Errorf("unable to start fat zk")
}

func (server *ZKServer) waitForStart(timeout int, maxRetry int, retryInterval int) bool {
	serverUrl := fmt.Sprintf("%s:%d", server.Address, server.Port)
	for i := 0; i < maxRetry; i++ {
		_, err := net.DialTimeout("tcp", serverUrl, time.Millisecond*time.Duration(timeout))
		if err == nil {
			return true
		}
		if i != maxRetry-1 {
			time.Sleep(time.Millisecond * time.Duration(retryInterval))
		}
	}
	return false
}

func (server *ZKServer) Stop() {
	os.RemoveAll(server.TempDataDir)
	server.Cmd.Process.Signal(os.Kill)
}

func findFatJarPath(jarName string) (jarPath string) {
	curDir, _ := os.Getwd()
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if strings.HasSuffix(path, jarName) {
			jarPath = path
		}
		return nil
	}
	filepath.Walk(curDir, walkFunc)
	return jarPath
}

const (
	DefaultServerTickTime  = 2000
	DefaultServerInitLimit = 10
	DefaultServerSyncLimit = 5
)

type ZKConfig struct {
	TickTime   int
	InitLimit  int
	SyncLimit  int
	DataDir    string
	ClientPort int
	Id         int
}

func (zc *ZKConfig) writeZooConfig(w io.Writer) {
	if zc.TickTime <= 0 {
		zc.TickTime = DefaultServerTickTime
	}
	if zc.InitLimit <= 0 {
		zc.InitLimit = DefaultServerInitLimit
	}
	if zc.SyncLimit <= 0 {
		zc.SyncLimit = DefaultServerSyncLimit
	}
	fmt.Fprintf(w, "tickTime=%d\n", zc.TickTime)
	fmt.Fprintf(w, "initLimit=%d\n", zc.InitLimit)
	fmt.Fprintf(w, "syncLimit=%d\n", zc.SyncLimit)
	fmt.Fprintf(w, "dataDir=%s\n", zc.DataDir)
	fmt.Fprintf(w, "clientPort=%d\n", zc.ClientPort)
	fmt.Fprintf(w, "server.%d=localhost:%d:%d\n", zc.Id, zc.ClientPort+1, zc.ClientPort+2)
}
