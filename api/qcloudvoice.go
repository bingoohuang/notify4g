package api

import (
	"github.com/bingoohuang/gou"
	"github.com/sirupsen/logrus"
	"strings"

	"strconv"
	"time"
)

// QcloudVoice 表示腾讯语音短信发送器
type QcloudVoice struct {
	QcloudBase
	TplID        int      `json:"tplID"`
	PlayTimes    int      `json:"playTimes"`
	TmplVarNames []string `json:"tmplVarNames"`
}

var _ Config = (*QcloudVoice)(nil)

// Config 加载配置
func (q *QcloudVoice) Config(config string) error {
	var tplID, playTimes, varNames string
	q.Sdkappid, q.Appkey, tplID, playTimes, varNames = gou.Split5(config, "/", true, false)
	q.TplID, _ = strconv.Atoi(tplID)
	q.PlayTimes, _ = strconv.Atoi(playTimes)
	q.TmplVarNames = strings.SplitN(varNames, "-", -1)

	return nil
}

func (s *QcloudVoice) InitMeaning() {
	s.Sdkappid = "sdkappid"
	s.Appkey = "appkey"
	s.TplID = 326476
	s.PlayTimes = 3
}

type QcloudVoiceReq struct {
	Params map[string]string `json:"params"`
	Mobile string            `json:"mobile" faker:"china_mobile_number"`
}

type RawQcloudVoiceRsp struct {
	Result int    `json:"result"`
	Errmsg string `json:"errmsg"`
	Callid string `json:"callid"`
	Ext    string `json:"ext"`
}

// RawQcloudVoiceReq 表示腾讯云语音短信请求体结构
type RawQcloudVoiceReq struct {
	TplID     int      `json:"tpl_id"`    // 模板 ID，在控制台审核通过的模板 ID
	Params    []string `json:"params"`    // 模板参数，若模板没有参数，请提供为空数组
	PlayTimes int      `json:"playtimes"` // 播放次数，可选，最多3次，默认2次。
	Sig       string   `json:"sig"`       // App 凭证，计算公式：sha256（appkey=$appkey&random=$random&time=$time&mobile=$mobile）
	Tel       Tel      `json:"tel"`
	Time      int64    `json:"time"` // 请求发起时间，UNIX 时间戳（单位：秒），如果和系统时间相差超过 10 分钟则会返回失败
	Ext       string   `json:"ext"`  // 用户的 session 内容，腾讯 server 回包中会原样返回，可选字段，不需要就填空
}

func (q QcloudVoice) NewRequest() interface{} { return &QcloudVoiceReq{} }
func (q QcloudVoice) ChannelName() string     { return qcloudvoice }

// 语音短信模板ID：326476   应用:{1} 监控埋点:{2} 在近{3}分钟内发生{4}, 其中最高{5}, 最低{6}
// 示例：应用:logcenter-flume 监控埋点:events成功写入kafka的数量#mssp_server_sink#192_168_22_1 在近10分钟内发生连续7次请求次数等于0.0, 其中最高2300.0, 最低1800.0

// Notify 发送信息
func (q QcloudVoice) Notify(request interface{}) NotifyRsp {
	r := request.(*QcloudVoiceReq)

	rando := gou.RandomIntAsString()
	// 发送语音通知 https://cloud.tencent.com/document/product/382/18155
	// https://github.com/tencentyun/qcloud-documents/blob/master/product/移动与通信/短信/开发指南/API 文档/语音API/指定模板发送语音.md
	url, _ := gou.BuildURL("https://cloud.tim.qq.com/v5/tlsvoicesvr/sendtvoice",
		map[string]string{"sdkappid": q.Sdkappid, "random": rando})
	logrus.Infof("url:%s", url)

	t := time.Now().Unix()

	req := &RawQcloudVoiceReq{
		TplID:     q.TplID,
		Params:    q.ConvertRequest(r.Params),
		PlayTimes: q.PlayTimes,
		Sig:       q.CreateSignature(rando, t, r.Mobile),
		Tel:       Tel{Mobile: r.Mobile, NationCode: "86"},
		Time:      t,
	}

	var rsp RawQcloudVoiceRsp
	_, err := gou.RestPost(url, req, &rsp)

	return MakeRsp(err, rsp.Result == 0, q.ChannelName(), rsp)
}

func (q QcloudVoice) ConvertRequest(r map[string]string) []string {
	params := make([]string, len(q.TmplVarNames))
	for i, k := range q.TmplVarNames {
		if v, ok := r[k]; ok {
			params[i] = v
		}
	}

	return params
}
