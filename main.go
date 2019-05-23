package main

import (
	"fmt"
	"github.com/bingoohuang/faker"
	"github.com/bingoohuang/gou"
	"github.com/bingoohuang/notify4g/api"
	_ "github.com/bingoohuang/notify4g/statiq"
	"github.com/bingoohuang/statiq/fs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"net/http"
	"os"
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

	pflag.Parse()

	if *v {
		fmt.Printf("Build on %s from sha1 %s\n", buildTime, sha1ver)
		os.Exit(0)
	}
	if *help {
		pflag.PrintDefaults()
		os.Exit(0)
	}

	api.InitSha1verBuildTime(sha1ver, buildTime)
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
