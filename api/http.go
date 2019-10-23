package api

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sort"
	"strings"

	"github.com/bingoohuang/notify4g/util"

	"github.com/bingoohuang/faker"
	"github.com/bingoohuang/gou"
)

type NotifierItem struct {
	Name      string
	Channel   string
	ConfigIDs []string
}

type HomeData struct {
	Sha1ver   string
	BuildTime string
	Items     []NotifierItem
}

func HandleHome(app *App, homeTemplate string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ids := gou.MakeMultiMap()
		app.configCache.Walk(func(k string, v *NotifyConfig) {
			ids.Put(v.Type, k)
		})

		items := []NotifierItem{
			{Name: "阿里云短信", Channel: aliyunsms, ConfigIDs: findConfigIDs(ids, aliyunsms)},
			{Name: "钉钉机器人", Channel: dingtalkrobot, ConfigIDs: findConfigIDs(ids, dingtalkrobot)},
			{Name: "腾讯云短信", Channel: qcloudsms, ConfigIDs: findConfigIDs(ids, qcloudsms)},
			{Name: "腾讯云语音", Channel: qcloudvoice, ConfigIDs: findConfigIDs(ids, qcloudvoice)},
			{Name: "企业微信", Channel: qywx, ConfigIDs: findConfigIDs(ids, qywx)},
			{Name: "SMTP邮件", Channel: mail, ConfigIDs: findConfigIDs(ids, mail)},
			{Name: "聚合短信", Channel: sms, ConfigIDs: findConfigIDs(ids, sms)},
			{Name: "阿里大鱼短信", Channel: aliyundayusms, ConfigIDs: findConfigIDs(ids, aliyundayusms)},
		}

		homeTpl := template.Must(template.New("homeTpl").Parse(homeTemplate))
		if err := homeTpl.Execute(w, HomeData{Sha1ver: util.Version, BuildTime: util.Compile, Items: items}); err != nil {
			http.Error(w, err.Error(), 400)
		}
	}
}

func findConfigIDs(m *gou.MultiMap, configType string) []string {
	arr := make([]string, 0)

	if v, ok := m.Get(configType); ok {
		for _, i := range v {
			arr = append(arr, i.(string))
		}
	}

	sort.Strings(arr)

	return arr
}

type Tester interface {
	Send(*App) NotifyRsp
}

type AliyunsmsTester struct {
	Config AliyunSms    `json:"config"`
	Data   AliyunSmsReq `json:"data"`
}

type AlidayuTester struct {
	Config AliyunDaYuSms    `json:"config"`
	Data   AliyunDaYuSmsReq `json:"data"`
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

func (r AlidayuTester) Send(app *App) NotifyRsp        { return r.Config.Notify(app, &r.Data) }
func (r AliyunsmsTester) Send(app *App) NotifyRsp      { return r.Config.Notify(app, &r.Data) }
func (r DingtalkReqTester) Send(app *App) NotifyRsp    { return r.Config.Notify(app, &r.Data) }
func (r QcloudSmsReqTester) Send(app *App) NotifyRsp   { return r.Config.Notify(app, &r.Data) }
func (r QcloudSmsVoiceTester) Send(app *App) NotifyRsp { return r.Config.Notify(app, &r.Data) }
func (r QywxTester) Send(app *App) NotifyRsp           { return r.Config.Notify(app, &r.Data) }
func (r MailTester) Send(app *App) NotifyRsp           { return r.Config.Notify(app, &r.Data) }
func (r SmsTester) Send(app *App) NotifyRsp            { return r.Config.Notify(app, &r.Data) }

func HandleRaw(app *App, path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = handleRawInternal(app, path, w, r)
	}
}

func handleRawInternal(app *App, path string, w http.ResponseWriter, r *http.Request) error {
	subs := strings.SplitN(r.URL.Path[len(path):], "/", -1)
	if len(subs) != 1 {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: "invalid path"})
	}

	configType := subs[0]
	tester := newTester(configType)

	if tester == nil {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: "invalid type " + configType})
	}

	switch r.Method {
	case GET:
		_ = faker.Fake(tester)

		return WriteJSON(w, tester)
	case POST:
		if err := json.NewDecoder(r.Body).Decode(tester); err != nil {
			return WriteErrorJSON(400, w, Rsp{Status: 400, Message: err.Error()})
		}

		rsp := tester.Send(app)

		return WriteJSON(w, rsp)
	default:
		return WriteErrorJSON(404, w, Rsp{Status: 404, Message: "Not Found"})
	}
}

func newTester(configType string) Tester {
	v := gou.Decode(configType,
		aliyunsms, &AliyunsmsTester{},
		aliyundayusms, &AlidayuTester{},
		dingtalkrobot, &DingtalkReqTester{},
		qcloudsms, &QcloudSmsReqTester{},
		qcloudvoice, &QcloudSmsVoiceTester{},
		qywx, &QywxTester{},
		mail, &MailTester{}, sms, &SmsTester{})
	if v != nil {
		return v.(Tester)
	}

	return nil
}
