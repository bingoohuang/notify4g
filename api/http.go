package api

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sort"
	"strings"

	"github.com/bingoohuang/gou/coll"
	"github.com/bingoohuang/gou/str"

	"github.com/bingoohuang/notify4g/util"

	"github.com/bingoohuang/faker"
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
		ids := coll.MakeMultiMap()

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
			{Name: "Exchange邮件", Channel: exchangewebservice, ConfigIDs: findConfigIDs(ids, exchangewebservice)},
			{Name: "聚合短信", Channel: sms, ConfigIDs: findConfigIDs(ids, sms)},
			{Name: "阿里大鱼短信", Channel: aliyundayusms, ConfigIDs: findConfigIDs(ids, aliyundayusms)},
		}

		homeTpl := template.Must(template.New("homeTpl").Parse(homeTemplate))
		if err := homeTpl.Execute(w, HomeData{Sha1ver: util.Version, BuildTime: util.Compile, Items: items}); err != nil {
			http.Error(w, err.Error(), 400)
		}
	}
}

func findConfigIDs(m *coll.MultiMap, configType string) []string {
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
	RedListFilter

	Send(*App) NotifyRsp
}

type AliyunsmsTester struct {
	Config AliyunSms    `json:"config"`
	Data   AliyunSmsReq `json:"data"`
}

var _ Tester = (*AliyunsmsTester)(nil)

type AlidayuTester struct {
	Config AliyunDaYuSms    `json:"config"`
	Data   AliyunDaYuSmsReq `json:"data"`
}

var _ Tester = (*AlidayuTester)(nil)

type DingtalkReqTester struct {
	Config Dingtalk    `json:"config"`
	Data   DingtalkReq `json:"data"`
}

var _ Tester = (*DingtalkReqTester)(nil)

type QcloudSmsReqTester struct {
	Config QcloudSms    `json:"config"`
	Data   QcloudSmsReq `json:"data"`
}

var _ Tester = (*QcloudSmsReqTester)(nil)

type QcloudSmsVoiceTester struct {
	Config QcloudVoice    `json:"config"`
	Data   QcloudVoiceReq `json:"data"`
}

var _ Tester = (*QcloudSmsVoiceTester)(nil)

type QywxTester struct {
	Config Qywx    `json:"config"`
	Data   QywxReq `json:"data"`
}

var _ Tester = (*QywxTester)(nil)

type MailTester struct {
	Config Mail    `json:"config"`
	Data   MailReq `json:"data"`
}

var _ Tester = (*MailTester)(nil)

type SmsTester struct {
	Config Sms    `json:"config"`
	Data   SmsReq `json:"data"`
}

var _ Tester = (*SmsTester)(nil)

func (r AlidayuTester) FilterRedList(list redList) bool        { return r.Data.FilterRedList(list) }
func (r AliyunsmsTester) FilterRedList(list redList) bool      { return r.Data.FilterRedList(list) }
func (r DingtalkReqTester) FilterRedList(list redList) bool    { return r.Data.FilterRedList(list) }
func (r QcloudSmsReqTester) FilterRedList(list redList) bool   { return r.Data.FilterRedList(list) }
func (r QcloudSmsVoiceTester) FilterRedList(list redList) bool { return r.Data.FilterRedList(list) }
func (r QywxTester) FilterRedList(list redList) bool           { return r.Data.FilterRedList(list) }
func (r MailTester) FilterRedList(list redList) bool           { return r.Data.FilterRedList(list) }
func (r SmsTester) FilterRedList(list redList) bool            { return r.Data.FilterRedList(list) }

func (r AlidayuTester) Send(app *App) NotifyRsp        { return r.Config.Notify(app, &r.Data) }
func (r AliyunsmsTester) Send(app *App) NotifyRsp      { return r.Config.Notify(app, &r.Data) }
func (r DingtalkReqTester) Send(app *App) NotifyRsp    { return r.Config.Notify(app, &r.Data) }
func (r QcloudSmsReqTester) Send(app *App) NotifyRsp   { return r.Config.Notify(app, &r.Data) }
func (r QcloudSmsVoiceTester) Send(app *App) NotifyRsp { return r.Config.Notify(app, &r.Data) }
func (r QywxTester) Send(app *App) NotifyRsp           { return r.Config.Notify(app, &r.Data) }
func (r MailTester) Send(app *App) NotifyRsp           { return r.Config.Notify(app, &r.Data) }
func (r SmsTester) Send(app *App) NotifyRsp            { return r.Config.Notify(app, &r.Data) }

func HandleRedlist(a *App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case GET:
			list := a.configCache.ReadRedList()
			_ = WriteJSON(w, list)
		case POST:
			var list RedList
			if err := json.NewDecoder(r.Body).Decode(&list); err != nil {
				_ = WriteErrorJSON(http.StatusBadRequest, w,
					Rsp{Status: http.StatusBadRequest, Message: err.Error()})
			} else {
				_ = a.configCache.WriteRedList(list, true)
				_ = WriteJSON(w, Rsp{Status: http.StatusOK, Message: "OK"})
			}
		default:
			_ = WriteErrorJSON(http.StatusNotFound, w,
				Rsp{Status: http.StatusNotFound, Message: "Not Found"})
		}
	}
}

func HandleRaw(app *App, path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = handleRawInternal(app, path, w, r)
	}
}

func handleRawInternal(app *App, path string, w http.ResponseWriter, r *http.Request) error {
	subs := strings.SplitN(r.URL.Path[len(path):], "/", -1)
	if len(subs) != 1 {
		return WriteErrorJSON(http.StatusBadRequest, w,
			Rsp{Status: http.StatusBadRequest, Message: "invalid path"})
	}

	configType := subs[0]
	tester := newTester(app, configType)

	if tester == nil {
		return WriteErrorJSON(http.StatusBadRequest, w,
			Rsp{Status: http.StatusBadRequest, Message: "invalid type " + configType})
	}

	switch r.Method {
	case GET:
		_ = faker.Fake(tester)

		return WriteJSON(w, tester)
	case POST:
		if err := json.NewDecoder(r.Body).Decode(tester); err != nil {
			return WriteErrorJSON(http.StatusBadRequest, w,
				Rsp{Status: http.StatusBadRequest, Message: err.Error()})
		}

		rsp := tester.Send(app)

		return WriteJSON(w, rsp)
	default:
		return WriteErrorJSON(http.StatusNotFound, w,
			Rsp{Status: http.StatusNotFound, Message: "Not Found"})
	}
}

func newTester(a *App, configType string) Tester {
	if v := str.Decode(configType,
		aliyunsms, &AliyunsmsTester{},
		aliyundayusms, &AlidayuTester{},
		dingtalkrobot, &DingtalkReqTester{},
		qcloudsms, &QcloudSmsReqTester{},
		qcloudvoice, &QcloudSmsVoiceTester{},
		qywx, &QywxTester{},
		mail, &MailTester{}, sms, &SmsTester{}); v != nil {
		tester := v.(Tester)
		list := a.configCache.ReadRedList()

		if tester.FilterRedList(list.prepare()) {
			return tester
		}
	}

	return nil
}
