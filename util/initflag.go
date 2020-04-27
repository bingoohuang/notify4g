package util

import (
	"fmt"
	_ "net/http/pprof" // nolint G108
	"os"

	"github.com/bingoohuang/gou/cnf"
	"github.com/bingoohuang/gou/file"
	"github.com/bingoohuang/gou/lo"

	"github.com/bingoohuang/gou/htt"

	"github.com/bingoohuang/gostarter/util"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func InitFlags() {
	help := pflag.BoolP("help", "h", false, "help")
	ipo := pflag.BoolP("init", "i", false, "init to create template config file and ctl.sh")
	pflag.StringP("addr", "a", ":11472", "http address to listen and serve")
	configFile := pflag.StringP("config", "c", "./config.toml", "config file path")
	pflag.StringP("auth", "u", "", "basic auth username and password eg admin:admin")
	pflag.StringP("nopConfID", "", "nop", "nopConfID for no op testing")
	pflag.StringP("snapshotDir", "s", "./etc/snapshots", "snapshots for config")

	lo.DeclareLogPFlags()

	pprofAddr := htt.PprofAddrPflag()

	// Add more pflags can be set from command line
	// ...

	pflag.Parse()

	cnf.CheckUnknownPFlags()

	if *help {
		fmt.Printf("Built on %s from sha1 %s\n", Compile, Version)
		pflag.PrintDefaults()
		os.Exit(0)
	}

	Ipo(*ipo)
	htt.StartPprof(*pprofAddr)

	viper.SetEnvPrefix("NOTIFY4G")
	viper.AutomaticEnv()

	_ = viper.BindPFlags(pflag.CommandLine)

	if file.ExistsAsFile(*configFile) {
		viper.SetConfigFile(*configFile)

		if err := viper.ReadInConfig(); err != nil {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	util.InitGin(lo.SetupLog())
}
