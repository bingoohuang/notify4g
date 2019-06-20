package api

import (
	"strings"

	"github.com/bingoohuang/gou"
	"github.com/thoas/go-funk"

	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// QcloudSms 表示腾讯云短信发送器
type QcloudSms struct {
	QcloudBase
	TplID        int      `json:"tplID"`
	Sign         string   `json:"sign" faker:"-"`
	TmplVarNames []string `json:"tmplVarNames"`
}

var _ Config = (*QcloudSms)(nil)

// Config 加载配置
func (q *QcloudSms) Config(config string) error {
	var tplID string
	var varNames string
	q.Sdkappid, q.Appkey, tplID, q.Sign, varNames = gou.Split5(config, "/", true, false)
	q.TplID, _ = strconv.Atoi(tplID)
	q.TmplVarNames = strings.SplitN(varNames, "-", -1)

	return nil
}

func (q *QcloudSms) InitMeaning() {
	q.Sdkappid = "sdkappid"
	q.Appkey = "appkey"
	q.TplID = 157749
	q.Sign = "短信签名，可以为空"
	q.TmplVarNames = []string{"var1", "var2"}
}

// Tel 表示电话号码
type Tel struct {
	Mobile     string `json:"mobile" faker:"china_mobile_number"`
	NationCode string `json:"nationcode"`
}

// RawQcloudSmsReq 表示腾讯云短消息请求体结构
type RawQcloudSmsReq struct {
	Ext    string   `json:"ext"`            // 用户的 session 内容，腾讯 server 回包中会原样返回，可选字段，不需要就填空
	Params []string `json:"params"`         // 模板参数，若模板没有参数，请提供为空数组
	Sig    string   `json:"sig"`            // App 凭证，计算公式：sha256（appkey=$appkey&random=$random&time=$time&mobile=$mobile）
	Sign   string   `json:"sign" faker:"-"` // 短信签名，如果使用默认签名，该字段可缺省
	Tel    []Tel    `json:"tel"`
	Time   int64    `json:"time"`   // 请求发起时间，UNIX 时间戳（单位：秒），如果和系统时间相差超过 10 分钟则会返回失败
	TplID  int      `json:"tpl_id"` // 模板 ID，在控制台审核通过的模板 ID
}

type RawQcloudSmsRspDetail struct {
	Fee        int    `json:"fee"`
	Mobile     string `json:"mobile" faker:"china_mobile_number"`
	Nationcode string `json:"nationcode"`
	Sid        string `json:"sid"`
	Result     int    `json:"result"`
	Errmsg     string `json:"errmsg"`
}

type QcloudSmsRsp struct {
	Result int                     `json:"result"`
	Errmsg string                  `json:"errmsg"`
	Ext    string                  `json:"ext"`
	Detail []RawQcloudSmsRspDetail `json:"detail"`
}

type QcloudSmsReq struct {
	Params  []string `json:"params"`
	Mobiles []string `json:"mobiles" faker:"china_mobile_number"`
}

func (q QcloudSms) ChannelName() string     { return qcloudsms }
func (q QcloudSms) NewRequest() interface{} { return &QcloudSmsReq{} }

// 目前业务埋点监控告警模板如下:
// 短信模板ID：157749   应用:{1} 监控埋点:{2} 在近{3}分钟内发生{4}, 其中最高{5}, 最低{6}
// 示例：【北京数字认证】应用:logcenter-flume 监控埋点:events成功写入kafka的数量#mssp_server_sink#192_168_22_1 在近10分钟内发生连续7次请求次数等于0.0, 其中最高2300.0, 最低1800.0
// 模板 157749 参数列表 ["appName", "key","minutes","counts","max","min"]

// Notify 发送信息
func (q QcloudSms) Notify(request interface{}) NotifyRsp {
	r := request.(*QcloudSmsReq)

	rando := gou.RandomIntAsString()
	// 指定模板群发短信 https://cloud.tencent.com/document/product/382/5977
	url, _ := gou.BuildURL("https://yun.tim.qq.com/v5/tlssmssvr/sendmultisms2",
		map[string]string{"sdkappid": q.Sdkappid, "random": rando})
	logrus.Infof("url:%s", url)

	t := time.Now().Unix()
	req := &RawQcloudSmsReq{
		Params: r.Params,
		Sig:    q.CreateSignature(rando, t, r.Mobiles...),
		// Sign:   "北京数字认证",
		Tel:   funk.Map(r.Mobiles, func(tel string) Tel { return Tel{Mobile: tel, NationCode: "86"} }).([]Tel),
		Time:  t,
		TplID: q.TplID,
	}

	var rsp QcloudSmsRsp
	_, err := gou.RestPost(url, req, &rsp)

	return MakeRsp(err, rsp.Result == 0, q.ChannelName(), rsp)
}

var _ SmsNotifier = (*QcloudSms)(nil)

func (q QcloudSms) ConvertRequest(r *SmsReq) interface{} {
	params := make([]string, len(q.TmplVarNames))
	for i, k := range q.TmplVarNames {
		if v, ok := r.TemplateParams[k]; ok {
			params[i] = v
		}
	}

	return &QcloudSmsReq{
		Params:  params,
		Mobiles: r.Mobiles,
	}
}
