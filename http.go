package notify4g

import (
	"encoding/json"
	"html/template"
	"net/http"
)

func HandleHome(homeTemplate string) func(w http.ResponseWriter, r *http.Request) {
	homeTpl := template.Must(template.New("homeTpl").Parse(homeTemplate))

	return func(w http.ResponseWriter, r *http.Request) {
		if err := homeTpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}
}

type Tester interface {
	Send() (interface{}, error)
}

type AliyunsmsTester struct {
	Config AliyunSms    `json:"config"`
	Data   AliyunSmsReq `json:"data"`
}

type DingtalkReqTester struct {
	Config Dingtalk    `json:"config"`
	Data   DingtalkReq `json:"data"`
}

type QcloudSmsReqTester struct {
	Config QcloudSms    `json:"config"`
	Data   QcloudSmsReq `json:"data"`
}

type QcloudSmsVoiceTester struct {
	Config QcloudVoice    `json:"config"`
	Data   QcloudVoiceReq `json:"data"`
}

type QywxTester struct {
	Config Qywx    `json:"config"`
	Data   QywxReq `json:"data"`
}

type MailTester struct {
	Config Mail    `json:"config"`
	Data   MailReq `json:"data"`
}

func (a AliyunsmsTester) Send() (interface{}, error)      { return a.Config.Notify(a.Data) }
func (a DingtalkReqTester) Send() (interface{}, error)    { return a.Config.Notify(a.Data) }
func (a QcloudSmsReqTester) Send() (interface{}, error)   { return a.Config.Notify(a.Data) }
func (a QcloudSmsVoiceTester) Send() (interface{}, error) { return a.Config.Notify(a.Data) }
func (a QywxTester) Send() (interface{}, error)           { return a.Config.Notify(a.Data) }

func (a MailTester) Send() (interface{}, error) { return a.Config.Notify(a.Data) }

func HandleNotifier(u Tester) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			_ = json.NewEncoder(w).Encode(u)
		case "POST":
			if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}

			rsp, err := u.Send()
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}

			_ = json.NewEncoder(w).Encode(rsp)
		}
	}
}
