package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"github.com/waittttting/cRPC-control-center/conf"
	"github.com/waittttting/cRPC-control-center/server"
)

func main() {

	var configPath string
	flag.StringVar(&configPath, "config", "", "config path")
	flag.Parse()

	var config conf.CCSConf
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		logrus.Fatal(err)
	}
	ccs := server.NewControlCenterServer(&config)
	ccs.Start()
}
