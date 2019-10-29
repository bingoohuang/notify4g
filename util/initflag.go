package util

import (
	"fmt"
	_ "net/http/pprof" // nolint G108
	"os"
	"path/filepath"
	"strings"

	"github.com/bingoohuang/gou"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func InitFlags() {
	help := pflag.BoolP("help", "h", false, "help")
	ipo := pflag.BoolP("init", "i", false, "init to create template config file and ctl.sh")
	pflag.StringP("addr", "a", ":11472", "http address to listen and serve")
	configFile := pflag.StringP("config", "c", "./config.toml", "config file path")
	pflag.StringP("loglevel", "l", "info", "debug/info/warn/error")
	pflag.StringP("logdir", "d", "./var", "log dir")
	pflag.StringP("auth", "u", "", "basic auth username and password eg admin:admin")
	pflag.BoolP("logrus", "o", true, "enable logrus")
	pflag.StringP("snapshotDir", "s", "./etc/snapshots", "snapshots for config")

	pprofAddr := gou.PprofAddrPflag()

	// Add more pflags can be set from command line
	// ...

	pflag.Parse()

	args := pflag.Args()
	if len(args) > 0 {
		fmt.Printf("Unknown args %s\n", strings.Join(args, " "))
		pflag.PrintDefaults()
		os.Exit(-1)
	}

	if *help {
		fmt.Printf("Built on %s from sha1 %s\n", Compile, Version)
		pflag.PrintDefaults()
		os.Exit(0)
	}

	Ipo(*ipo)
	gou.StartPprof(*pprofAddr)

	viper.SetEnvPrefix("NOTIFY4G")
	viper.AutomaticEnv()

	_ = viper.BindPFlags(pflag.CommandLine)

	if fileExists(*configFile) {
		viper.SetConfigFile(*configFile)
		if err := viper.ReadInConfig(); err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	}

	if viper.GetBool("logrus") {
		logdir := viper.GetString("logdir")
		if err := os.MkdirAll(logdir, os.ModePerm); err != nil {
			logrus.Panicf("failed to create %s error %v\n", logdir, err)
		}

		loglevel := viper.GetString("loglevel")
		gou.InitLogger(loglevel, logdir, filepath.Base(os.Args[0])+".log")
	} else {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
