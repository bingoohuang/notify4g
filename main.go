package main

import (
	"bytes"
	"fmt"
	"github.com/bingoohuang/faker"
	"github.com/bingoohuang/gou"
	"github.com/bingoohuang/notify4g/api"
	_ "github.com/bingoohuang/notify4g/statiq"
	"github.com/bingoohuang/statiq/fs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

// refer : https://blog.kowalczyk.info/article/vEja/embedding-build-number-in-go-executable.html
var (
	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
)

func init() {
	help := pflag.BoolP("help", "h", false, "help")
	ipo := pflag.BoolP("init", "i", false, "init to create template config file and ctl.sh")
	pflag.StringP("addr", "a", ":11472", "http address to listen and serve")
	pflag.StringP("snapshotDir", "s", "./etc/snapshots", "snapshots for config")
	pflag.StringP("loglevel", "l", "info", "debug/info/warn/error")
	pflag.StringP("logdir", "d", "./var", "log dir")
	pflag.BoolP("logrus", "o", true, "enable logrus")
	pflag.Parse()

	if *help {
		fmt.Printf("Build on %s from sha1 %s\n", buildTime, sha1ver)
		pflag.PrintDefaults()
		os.Exit(0)
	}

	if *ipo {
		if err := ipoInit(); err != nil {
			fmt.Println(err)
		}
		os.Exit(0)
	}

	viper.SetDefault("addr", ":11472")
	viper.SetDefault("snapshotDir", "./etc/snapshots")
	viper.SetDefault("loglevel", "info")
	viper.SetDefault("logdir", "./var")
	viper.SetDefault("logrus", false)

	viper.SetEnvPrefix("notify4g")
	viper.AutomaticEnv()

	_ = viper.BindPFlags(pflag.CommandLine)

	api.InitSha1verBuildTime(sha1ver, buildTime)

	if viper.GetBool("logrus") {
		gou.InitLogger(viper.GetString("loglevel"), viper.GetString("logdir"), filepath.Base(os.Args[0])+".log")
	} else {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	defer gou.Recover()

	sfs, _ := fs.New()
	_ = faker.SetRandomMapAndSliceSize(1, 3)

	http.HandleFunc("/", api.HandleHome(string(sfs.Files["/home.html"].Data)))
	http.HandleFunc("/raw/", api.HandleRaw("/raw/"))
	http.HandleFunc("/config/", api.ServeByConfig("/config/"))
	http.HandleFunc("/notify/", api.NotifyByConfig("/notify/"))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(sfs)))

	api.InitConfigCache(viper.GetString("snapshotDir"))

	logrus.SetLevel(logrus.InfoLevel)
	addr := viper.GetString("addr")
	logrus.Infof("start to listen and serve on address %s", addr)
	logrus.Fatal(http.ListenAndServe(addr, nil))
}

func ipoInit() error {
	sfs, err := fs.New()
	if err != nil {
		return err
	}

	if err = initCtl(sfs, "/ctl.tpl.sh", "./ctl"); err != nil {
		return err
	}

	return nil
}

func initCtl(sfs *fs.StatiqFS, ctlTplName, ctlFilename string) error {
	if _, err := os.Stat(ctlFilename); err == nil {
		fmt.Println(ctlFilename + " already exists, ignored!")
		return nil
	} else if os.IsNotExist(err) {
		// continue
	} else {
		return err
	}

	ctl := string(sfs.Files[ctlTplName].Data)
	tpl, err := template.New(ctlTplName).Parse(ctl)
	if err != nil {
		return err
	}

	binArgs := make([]string, 0, len(os.Args)-2)
	for i, arg := range os.Args {
		if i == 0 {
			continue
		}
		if strings.Index(arg, "-i") == 0 || strings.Index(arg, "--init") == 0 {
			continue
		}

		if strings.Index(arg, "-") != 0 {
			arg = strconv.Quote(arg)
		}

		binArgs = append(binArgs, arg)
	}

	var content bytes.Buffer
	m := map[string]string{"BinName": os.Args[0], "BinArgs": strings.Join(binArgs, " ")}
	if err := tpl.Execute(&content, m); err != nil {
		return err
	}

	// 0755->即用户具有读/写/执行权限，组用户和其它用户具有读写权限；
	if err = ioutil.WriteFile(ctlFilename, content.Bytes(), 0755); err != nil {
		return err
	}

	fmt.Println(ctlFilename + " created!")
	return nil
}
