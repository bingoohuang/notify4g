package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bingoohuang/faker"
	"github.com/bingoohuang/gou"
	"github.com/pkg/errors"
)

type Config interface {
	Config(config string) error
	Notify(app *App, req interface{}) NotifyRsp
	ChannelName() string
	InitMeaning()
	NewRequest() interface{}
}

type App struct {
	configCache *NotifyConfigCache
}

func CreateApp(snapshotDir string) *App {
	return &App{configCache: NewCache(snapshotDir)}
}

type RawNotifyConfig struct {
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

func (r RawNotifyConfig) ParseConfig() (Config, error) {
	v, err := NewConfig(r.Type)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(r.Config), v); err != nil {
		return nil, err
	}

	return v, nil
}

func ParseNotifyConfig(content []byte) (*NotifyConfig, error) {
	var raw RawNotifyConfig
	if err := json.Unmarshal(content, &raw); err != nil {
		return nil, err
	}

	c, err := NewConfig(raw.Type)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(raw.Config), c); err != nil {
		return nil, err
	}

	return &NotifyConfig{Type: raw.Type, Config: c}, nil
}

type NotifyConfig struct {
	Type   string `json:"type"`
	Config Config `json:"config"`
}

func NewConfig(typ string) (Config, error) {
	v := gou.Decode(typ, "aliyunsms", &AliyunSms{}, "dingtalkrobot", &Dingtalk{},
		"qcloudsms", &QcloudSms{}, "qcloudvoice", &QcloudVoice{}, "qywx", &Qywx{}, "mail", &Mail{}, "sms", &Sms{},
		"aliyundayusms", &AliyunDaYuSms{})
	if v != nil {
		return v.(Config), nil
	}
	return nil, errors.New("unknown config type " + typ)
}

func (a *App) NotifyByConfig(removePath string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		subs := strings.SplitN(r.URL.Path[len(removePath):], "/", -1)
		if len(subs) != 1 {
			_ = WriteErrorJSON(400, w, Rsp{Status: 400, Message: "invalid path"})
			return
		}
		configID := subs[0]

		switch r.Method {
		case GET:
			_ = a.prepareNotify(w, configID)
		case POST:
			_ = a.postNotify(w, r, configID)
		default:
			_ = WriteErrorJSON(404, w, Rsp{Status: 404, Message: "Not Found"})
		}
	}
}

func (a *App) prepareNotify(w http.ResponseWriter, configID string) error {
	c := a.configCache.Read(configID)
	if c == nil {
		return WriteErrorJSON(404, w, Rsp{Status: 404, Message: "configID " + configID + " not found"})
	}

	req := c.Config.NewRequest()
	_ = faker.Fake(req)
	return WriteJSON(w, req)
}

func (a *App) postNotify(w http.ResponseWriter, r *http.Request, configID string) error {
	c := a.configCache.Read(configID)
	if c == nil {
		return WriteErrorJSON(404, w, Rsp{Status: 404, Message: "configID " + configID + " not found"})
	}

	req := c.Config.NewRequest()
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: err.Error()})
	}

	rsp := a.NotifyByConfigID(configID, req)
	return WriteJSON(w, rsp)
}

func (a *App) NotifyByConfigID(configID string, req interface{}) NotifyRsp {
	c := a.configCache.Read(configID)
	if c == nil {
		return MakeRsp(fmt.Errorf("configID %s not found", configID), false, "", nil)
	}

	rsp := c.Config.Notify(a, req)
	return rsp
}

func (a *App) ServeByConfig(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path[len(path):]
		subs := strings.SplitN(p, "/", -1)

		l := len(subs)
		switch r.Method {
		case GET:
			_ = a.GetConfig(w, l, subs)
		case POST:
			_ = a.PostConfig(w, r, l, subs)
		case "DELETE":
			_ = a.DeleteConfig(w, l, subs)
		default:
			_ = WriteErrorJSON(404, w, Rsp{Status: 404, Message: "Not Found"})
		}
	}
}

type Rsp struct {
	Status  int
	Message string
}

func (a *App) DeleteConfig(w http.ResponseWriter, l int, subs []string) error {
	if l != 1 {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: "invalid path"})
	}

	a.configCache.Delete(subs[0])
	return WriteJSON(w, Rsp{Status: 200, Message: "OK"})
}

func (a *App) PostConfig(w http.ResponseWriter, r *http.Request, l int, subs []string) error {
	if l != 1 {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: "invalid path"})
	}

	content := gou.ReadObjectBytes(r.Body)
	config, err := ParseNotifyConfig(content)
	if err != nil {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: err.Error()})
	}

	configID := subs[0]
	_ = a.configCache.Write(configID, config, true)
	return WriteJSON(w, Rsp{Status: 200, Message: "OK"})
}

func (a *App) GetConfig(w http.ResponseWriter, l int, subs []string) error {
	if l != 1 && l != 2 {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: "invalid path"})
	}
	configID := subs[0]
	if l == 2 {
		config, err := NewConfig(subs[1])
		if err != nil {
			return WriteErrorJSON(400, w, Rsp{Status: 400, Message: err.Error()})
		}

		config.InitMeaning()
		return WriteJSON(w, NotifyConfig{Type: subs[1], Config: config})
	}

	c := a.configCache.Read(configID)
	if c == nil {
		return WriteErrorJSON(404, w, Rsp{Status: 404, Message: "configID " + configID + " not found"})
	}

	return WriteJSON(w, c)
}

func WriteErrorJSON(statusCode int, w http.ResponseWriter, v interface{}) error {
	w.WriteHeader(statusCode)
	return WriteJSON(w, v)
}

func WriteJSON(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(v)
}
