package api

import (
	"encoding/json"
	"github.com/bingoohuang/gou"
	"github.com/bxcodec/faker/v3"
	"github.com/pkg/errors"
	"net/http"
	"reflect"
	"strings"
)

type Config interface {
	Config(config string) error
	Notify(req interface{}) (interface{}, error)
	InitMeaning()
	NewRequest() interface{}
}

var configCache *NotifyConfigCache

func InitConfigCache(snapshotDir string) {
	configCache = NewCache(snapshotDir)
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
	var v Config
	switch typ {
	case "aliyunsms":
		v = &AliyunSms{}
	case "dingtalkrobot":
		v = &Dingtalk{}
	case "qcloudsms":
		v = &QcloudSms{}
	case "qcloudvoice":
		v = &QcloudVoice{}
	case "qywx":
		v = &Qywx{}
	case "mail":
		v = &Mail{}
	default:
		return nil, errors.New("unknown config type " + typ)
	}

	return v, nil
}

func NotifyByConfig(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path[len(path):]
		subs := strings.SplitN(p, "/", -1)

		l := len(subs)
		if l != 1 {
			WriteErrorJSON(400, w, Rsp{Status: 400, Message: "invalid path"})
			return
		}
		configId := subs[0]

		switch r.Method {
		case "POST":
			Notify(w, r, configId)
		case "GET":
			PrepareNotify(w, configId)
		default:
			WriteErrorJSON(404, w, Rsp{Status: 404, Message: "Not Found"})
		}
	}
}

func PrepareNotify(w http.ResponseWriter, configId string) error {
	c := configCache.Read(configId)
	if c == nil {
		return WriteErrorJSON(404, w, Rsp{Status: 404, Message: "configId " + configId + " not found"})
	}

	req := c.Config.NewRequest()
	faker.FakeData(req)
	return WriteJSON(w, req)
}

func Notify(w http.ResponseWriter, r *http.Request, configId string) error {
	c := configCache.Read(configId)
	if c == nil {
		return WriteErrorJSON(404, w, Rsp{Status: 404, Message: "configId " + configId + " not found"})
	}

	req := c.Config.NewRequest()
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: err.Error()})
	}

	rsp, err := c.Config.Notify(reflect.ValueOf(req).Elem().Interface())
	if err != nil {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: err.Error()})
	}
	return WriteJSON(w, rsp)
}

func ServeByConfig(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path[len(path):]
		subs := strings.SplitN(p, "/", -1)

		l := len(subs)
		switch r.Method {
		case "GET":
			GetConfig(w, l, subs)
		case "POST":
			PostConfig(w, r, l, subs)
		case "DELETE":
			DeleteConfig(w, l, subs)
		default:
			WriteErrorJSON(404, w, Rsp{Status: 404, Message: "Not Found"})
		}
	}
}

type Rsp struct {
	Status  int
	Message string
}

func DeleteConfig(w http.ResponseWriter, l int, subs []string) error {
	if l != 1 {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: "invalid path"})
	}

	configCache.Delete(subs[0])
	return WriteJSON(w, Rsp{Status: 200, Message: "OK"})
}

func PostConfig(w http.ResponseWriter, r *http.Request, l int, subs []string) error {
	if l != 1 {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: "invalid path"})
	}

	configId := subs[0]

	content := gou.ReadObjectBytes(r.Body)
	config, err := ParseNotifyConfig(content)
	if err != nil {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: err.Error()})
	}

	configCache.Write(configId, config, true)
	return WriteJSON(w, Rsp{Status: 200, Message: "OK"})
}

func GetConfig(w http.ResponseWriter, l int, subs []string) error {
	if l != 1 && l != 2 {
		return WriteErrorJSON(400, w, Rsp{Status: 400, Message: "invalid path"})
	}
	configId := subs[0]
	if l == 2 {
		config, err := NewConfig(subs[1])
		if err != nil {
			return WriteErrorJSON(400, w, Rsp{Status: 400, Message: err.Error()})
		}

		config.InitMeaning()
		return WriteJSON(w, NotifyConfig{Type: subs[1], Config: config})

	}

	c := configCache.Read(configId)
	if c == nil {
		return WriteErrorJSON(404, w, Rsp{Status: 404, Message: "configId " + configId + " not found"})
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
