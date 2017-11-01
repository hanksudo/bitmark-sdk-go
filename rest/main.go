package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/logger"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/hcl"
)

var (
	cfg    *config
	netork sdk.Network
	log    *logger.L
)

type config struct {
	Chain   string `hcl:"network"`
	Port    int    `hcl:"port"`
	DataDir string `hcl:"datadir"`
}

func init() {
	var confpath string
	flag.StringVar(&confpath, "conf", "", "Specify configuration file")
	flag.Parse()

	cfg = readConfig(confpath)

	if err := logger.Initialise(logger.Configuration{
		Directory: cfg.DataDir,
		File:      "trade.log",
		Size:      1048576,
		Count:     10,
		Levels:    map[string]string{"DEFAULT": "info"},
	}); err != nil {
		panic(fmt.Sprintf("logger initialization failed: %s", err))
	}

	netork = sdk.Livenet
	if cfg.Chain == "test" {
		netork = sdk.Testnet
	}

	log = logger.New("")
}

func readConfig(confpath string) *config {
	var cfg config

	dat, err := ioutil.ReadFile(confpath)
	if err != nil {
		panic(fmt.Sprintf("unable to read the configuration: %v", err))
	}

	if err = hcl.Unmarshal(dat, &cfg); nil != err {
		panic(fmt.Sprintf("unable to parse the configuration: %v", err))
	}

	return &cfg
}

func main() {
	r := gin.Default()
	r.POST("/account", handleCreateAccount())
	r.POST("/issue", handleIssue())
	r.POST("/transfer", handleTransfer())
	r.POST("/download", handleDownloadAsset())
	r.Run(fmt.Sprintf(":%d", cfg.Port))
}
