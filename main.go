package main

import (
	"net/http"

	"github.com/bingoohuang/faker"
	"github.com/bingoohuang/notify4g/api"
	_ "github.com/bingoohuang/notify4g/statiq"
	"github.com/bingoohuang/notify4g/util"
	"github.com/bingoohuang/statiq/fs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	util.InitFlags()

	sfs, _ := fs.New()
	_ = faker.SetRandomMapAndSliceSize(1, 3)

	app := api.CreateApp(viper.GetString("snapshotDir"))

	http.HandleFunc("/", api.HandleHome(app, string(sfs.Files["/home.html"].Data)))
	http.HandleFunc("/raw/", api.HandleRaw(app, "/raw/"))

	http.HandleFunc("/config/", app.ServeByConfig("/config/"))
	http.HandleFunc("/notify/", app.NotifyByConfig("/notify/"))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(sfs)))

	logrus.SetLevel(logrus.InfoLevel)
	addr := viper.GetString("addr")
	logrus.Infof("start to listen and serve on address %s", addr)
	logrus.Fatal(http.ListenAndServe(addr, nil))
}
