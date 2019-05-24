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
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
)

// refer : https://blog.kowalczyk.info/article/vEja/embedding-build-number-in-go-executable.html
var (
	sha1ver   string // sha1 revision used to build the program
	buildTime string // when the executable was built
)

var addr *string
var snapshotDir *string

func init() {
	help := pflag.BoolP("help", "h", false, "help")
	v := pflag.BoolP("version", "v", false, "show version and exit")
	addr = pflag.StringP("addr", "a", ":11472", "http address to listen and serve")
	snapshotDir = pflag.StringP("snapshotDir", "s", "./etc/snapshots", "snapshots for config")
	ipo := pflag.BoolP("init", "i", false, "init to create template config file and ctl.sh")
	loglevel := pflag.StringP("loglevel", "l", "info", "debug/info/warn/error")
	logrusEnabled := pflag.BoolP("logrus", "o", true, "enable logrus")
	pflag.Parse()

	if *v {
		fmt.Printf("Build on %s from sha1 %s\n", buildTime, sha1ver)
		os.Exit(0)
	}
	if *help {
		pflag.PrintDefaults()
		os.Exit(0)
	}

	if *ipo {
		if err := ipoInit(); err != nil {
			fmt.Println(err)
		}
		os.Exit(0)
	}

	api.InitSha1verBuildTime(sha1ver, buildTime)

	if *logrusEnabled {
		gou.InitLogger(*loglevel, "./var", filepath.Base(os.Args[0])+".log")
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

	api.InitConfigCache(*snapshotDir)

	logrus.SetLevel(logrus.InfoLevel)
	logrus.Infof("start to listen and serve on address %s", *addr)
	logrus.Fatal(http.ListenAndServe(*addr, nil))
}

func ipoInit() error {
	sfs, err := fs.New()
	if err != nil {
		return err
	}

	if err = initCtl(sfs, "/ctl.tpl.sh", "./ctl", ""); err != nil {
		return err
	}

	return nil
}

func initCtl(sfs *fs.StatiqFS, ctlTplName, ctlFilename, binArgs string) error {
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

	var content bytes.Buffer
	if err := tpl.Execute(&content, map[string]string{"BinName": os.Args[0], "BinArgs": binArgs}); err != nil {
		return err
	}

	// 0755->即用户具有读/写/执行权限，组用户和其它用户具有读写权限；
	if err = ioutil.WriteFile(ctlFilename, content.Bytes(), 0755); err != nil {
		return err
	}

	fmt.Println(ctlFilename + " created!")
	return nil
}
