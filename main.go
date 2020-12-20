package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"github.com/waittttting/cx-rpc/control_center/conf"
)

func main() {

	var configPath string
	flag.StringVar(&configPath, "config", "", "config path")
	flag.Parse()

	var config conf.CCSConf
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		logrus.Fatal(err)
	}
	ccs := NewControlCenterServer(&config)
	ccs.Start()
}
