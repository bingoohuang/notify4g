package main

import (
	"flag"
	"github.com/bingoohuang/statiq/fs"
	"github.com/bxcodec/faker/v3"
	"github.com/sirupsen/logrus"
	"net/http"
	"notify4g/api"
	"os"

	_ "notify4g/statiq"
)

func main() {
	fs, _ := fs.New()
	faker.SetRandomMapAndSliceSize(3)

	http.HandleFunc("/", api.HandleHome(string(fs.Files["/home.html"].Data)))

	http.HandleFunc("/raw/aliyunsms", api.HandleNotifier(&api.AliyunsmsTester{}))
	http.HandleFunc("/raw/dingtalkrobot", api.HandleNotifier(&api.DingtalkReqTester{}))
	http.HandleFunc("/raw/qcloudsms", api.HandleNotifier(&api.QcloudSmsReqTester{}))
	http.HandleFunc("/raw/qcloudvoice", api.HandleNotifier(&api.QcloudSmsVoiceTester{}))
	http.HandleFunc("/raw/qywx", api.HandleNotifier(&api.QywxTester{}))
	http.HandleFunc("/raw/mail", api.HandleNotifier(&api.MailTester{}))

	http.HandleFunc("/config/", api.ServeByConfig("/config/"))
	http.HandleFunc("/notify/", api.NotifyByConfig("/notify/"))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(fs)))

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
