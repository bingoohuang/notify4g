package api

import (
	"encoding/json"
	"github.com/bingoohuang/faker"
	"github.com/bingoohuang/gou"
	"html/template"
	"net/http"
	"sort"
	"strings"
)

type NotifierItem struct {
	Name      string
	Path      string
	ConfigIDs []string
}

type HomeData struct {
	Items []NotifierItem
}

func HandleHome(homeTemplate string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		homeTpl := template.Must(template.New("homeTpl").Parse(homeTemplate))

		ids := gou.MakeMultiMap()
		ConfigCache.Walk(func(k string, v *NotifyConfig) {
			ids.Put(v.Type, k)
		})

		items := []NotifierItem{
			{Name: "阿里云短信", Path: "/raw/aliyunsms", ConfigIDs: findConfigIDs(ids, "aliyunsms")},
			{Name: "钉钉机器人", Path: "/raw/dingtalkrobot", ConfigIDs: findConfigIDs(ids, "dingtalkrobot")},
			{Name: "腾讯云短信", Path: "/raw/qcloudsms", ConfigIDs: findConfigIDs(ids, "qcloudsms")},
			{Name: "腾讯云语音", Path: "/raw/qcloudvoice", ConfigIDs: findConfigIDs(ids, "qcloudvoice")},
			{Name: "企业微信", Path: "/raw/qywx", ConfigIDs: findConfigIDs(ids, "qywx")},
			{Name: "SMTP邮件", Path: "/raw/mail", ConfigIDs: findConfigIDs(ids, "mail")},
			{Name: "聚合短信", Path: "/raw/sms", ConfigIDs: findConfigIDs(ids, "sms")},
		}

		if err := homeTpl.Execute(w, HomeData{Items: items}); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
	}
}

func findConfigIDs(m *gou.MultiMap, configType string) []string {
	arr := make([]string, 0)
	v, ok := m.Get(configType)
	if ok {
		for _, i := range v {
			arr = append(arr, i.(string))
		}
	}

	sort.Strings(arr)
	return arr
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

type SmsTester struct {
	Config Sms    `json:"config"`
	Data   SmsReq `json:"data"`
}

func (r AliyunsmsTester) Send() (interface{}, error)      { return r.Config.Notify(&r.Data) }
func (r DingtalkReqTester) Send() (interface{}, error)    { return r.Config.Notify(&r.Data) }
func (r QcloudSmsReqTester) Send() (interface{}, error)   { return r.Config.Notify(&r.Data) }
func (r QcloudSmsVoiceTester) Send() (interface{}, error) { return r.Config.Notify(&r.Data) }
func (r QywxTester) Send() (interface{}, error)           { return r.Config.Notify(&r.Data) }
func (r MailTester) Send() (interface{}, error)           { return r.Config.Notify(&r.Data) }
func (r SmsTester) Send() (interface{}, error)            { return r.Config.Notify(&r.Data) }

func HandleRaw(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = handleRawInternal(path, w, r)
	}
}

func handleRawInternal(path string, w http.ResponseWriter, r *http.Request) error {
	subs := strings.SplitN(r.URL.Path[len(path):], "/", -1)
	if len(subs) != 1 {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: "invalid path"})
	}

	configType := subs[0]
	tester := createTester(configType)
	if tester == nil {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: "invalid type " + configType})
	}

	switch r.Method {
	case "GET":
		_ = faker.Fake(tester)
		return WriteJSON(w, tester)
	case "POST":
		if err := json.NewDecoder(r.Body).Decode(tester); err != nil {
			return WriteErrorJSON(400, w, Rsp{Status: 400, Message: err.Error()})
		}

		if rsp, err := tester.Send(); err != nil {
			return WriteErrorJSON(400, w, Rsp{Status: 400, Message: err.Error()})
		} else {
			return WriteJSON(w, rsp)
		}
	default:
		return WriteErrorJSON(404, w, Rsp{Status: 404, Message: "Not Found"})
	}
}

func createTester(configType string) Tester {
	switch configType {
	case "aliyunsms":
		return &AliyunsmsTester{}
	case "dingtalkrobot":
		return &DingtalkReqTester{}
	case "qcloudsms":
		return &QcloudSmsReqTester{}
	case "qcloudvoice":
		return &QcloudSmsVoiceTester{}
	case "qywx":
		return &QywxTester{}
	case "mail":
		return &MailTester{}
	case "sms":
		return &SmsTester{}
	default:
		return nil
	}
}
