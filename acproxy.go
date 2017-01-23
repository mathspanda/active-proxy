package main

import (
	"os"

	"active-proxy/cmd"
	. "active-proxy/server"

	"github.com/golang/glog"
)

func main() {
	defer glog.Flush()

	mainCmd := cmd.NewProxyCommand(Start)
	if err := mainCmd.Execute(); err != nil {
		glog.Errorln("Command error: ", err)
		os.Exit(-1)
	}
}

func Start(configFile string) {
	conf, err := NewProxyConf(configFile)
	if err != nil {
		glog.Errorln("Error init proxy configuration: ", err)
		os.Exit(-1)
	}
	glog.V(2).Infof("ProxyConf: %+v", conf)

	proxyServer, err := NewProxyServer(*conf)
	if err != nil {
		glog.Errorln("Error init proxy server: ", err)
		os.Exit(-1)
	}
	proxyServer.StartServer()
}
