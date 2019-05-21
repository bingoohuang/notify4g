package main

import (
	"flag"
	"github.com/bingoohuang/faker"
	"github.com/bingoohuang/gou"
	"github.com/bingoohuang/statiq/fs"
	"github.com/sirupsen/logrus"
	"net/http"
	"notify4g/api"
	"os"

	_ "notify4g/statiq"
)

func main() {
	defer gou.Recover()

	sfs, _ := fs.New()
	_ = faker.SetRandomMapAndSliceSize(1, 3)

	http.HandleFunc("/", api.HandleHome(string(sfs.Files["/home.html"].Data)))
	http.HandleFunc("/raw/", api.HandleRaw("/raw/"))
	http.HandleFunc("/config/", api.ServeByConfig("/config/"))
	http.HandleFunc("/notify/", api.NotifyByConfig("/notify/"))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(sfs)))

	help := flag.Bool("h", false, "help")
	addr := flag.String("addr", ":8080", "http address to listen and serve")
	snapshotDir := flag.String("snapshotDir", "./etc/snapshots", "snapshots for config")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	api.InitConfigCache(*snapshotDir)

	logrus.SetLevel(logrus.InfoLevel)
	logrus.Infof("start to listen and serve on address %s", *addr)
	logrus.Fatal(http.ListenAndServe(*addr, nil))
}
