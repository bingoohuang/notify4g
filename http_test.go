package notify4g

import (
	"github.com/bingoohuang/statiq/fs"
	"github.com/sirupsen/logrus"
	"net/http"
	"testing"

	_ "notify4g/statiq"
)

func TestHandleAliyunSms(t *testing.T) {
	statiqFS, _ := fs.New()

	http.HandleFunc("/", HandleHome(string(statiqFS.Files["/home.html"].Data)))

	http.HandleFunc("/aliyunsms", HandleNotifier(&AliyunsmsTester{}))
	http.HandleFunc("/dingtalkrobot", HandleNotifier(&DingtalkReqTester{}))
	http.HandleFunc("/qcloudSms", HandleNotifier(&QcloudSmsReqTester{}))
	http.HandleFunc("/qcloudVoice", HandleNotifier(&QcloudSmsVoiceTester{}))
	http.HandleFunc("/qywx", HandleNotifier(&QywxTester{}))
	http.HandleFunc("/mail", HandleNotifier(&MailTester{}))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(statiqFS)))
	logrus.Fatal(http.ListenAndServe(":8080", nil))
}
