package api

import (
	"net/url"

	"github.com/bingoohuang/gou"
	"github.com/sirupsen/logrus"
	"github.com/tobyzxj/uuid"

	"strings"
	"time"
)

// AliyunSms 表示阿里云短信发送器
type AliyunSms struct {
	AccessKeyID     string `json:"accessKeyID"`
	AccessKeySecret string `json:"accessKeySecret"`
	TemplateCode    string `json:"templateCode"`
	SignName        string `json:"signName" faker:"-"`
}

var _ Config = (*AliyunSms)(nil)

// Config 创建发送器，要求参数 config 是{accessKeyId}/{accessKeySecret}/{templateCode}/{signName}的格式
func (s *AliyunSms) Config(config string) error {
	s.AccessKeyID, s.AccessKeySecret, s.TemplateCode, s.SignName = gou.Split4(config, "/", true, false)
	return nil
}

func (s *AliyunSms) InitMeaning() {
	s.AccessKeyID = "accessKeyID"
	s.AccessKeySecret = "accessKeySecret"
	s.TemplateCode = "短信模板ID，可以不设置，然后在发送时再设置"
	s.SignName = "短信模签名，不设置使用默认签名，或者在发送时再设置"
}

type AliyunSmsReq struct {
	TemplateCode   string            `json:"templateCode" faker:"-"`
	TemplateParams map[string]string `json:"templateParams"`
	SignName       string            `json:"signName" faker:"-"`
	Mobiles        []string          `json:"mobiles" faker:"china_mobile_number"`
}

func (a *AliyunSmsReq) FilterRedList(list redList) bool {
	a.Mobiles = list.FilterMobiles(a.Mobiles)

	return len(a.Mobiles) > 0
}

type AliyunSmsRsp struct {
	OutID string `json:"outId"`

	Code      string `json:"code"`      // eg. OK 请求状态码。 返回OK代表请求成功。 其他错误码详见错误码列表。
	Message   string `json:"message"`   // eg. OK 状态码的描述。
	RequestID string `json:"requestID"` // eg. F655A8D5-B967-440B-8683-DAD6FF8DE990	 请求ID。

	BizID string `json:"bizID"` // eg. 900619746936498440^0,  发送回执ID，可根据该ID在接口QuerySendDetails中查询具体的发送状态。
}

func (s AliyunSms) NewRequest() Request { return &AliyunSmsReq{} }
func (s AliyunSms) ChannelName() string { return aliyunsms }

// Notify 发送短信
func (s AliyunSms) Notify(_ *App, request Request) NotifyRsp {
	req := request.(*AliyunSmsReq)
	param, outID := s.createParams(req)
	u, _ := gou.BuildURL("http://dysmsapi.aliyuncs.com/", param)

	r := AliyunSmsRsp{OutID: outID}
	err := gou.RestGet(u, &r)

	return MakeRsp(err, r.Code == "OK", s.ChannelName(), r)
}

var _ SmsNotifier = (*AliyunSms)(nil)

func (s AliyunSms) ConvertRequest(r *SmsReq) Request {
	return &AliyunSmsReq{TemplateParams: r.TemplateParams, Mobiles: r.Mobiles}
}

// api doc: https://help.aliyun.com/document_detail/101414.html?spm=a2c4g.11186623.6.616.1eee202a1PxPlf
func (s AliyunSms) createParams(req *AliyunSmsReq) (map[string]string, string) {
	outID := gou.RandomString(16)
	param := map[string]string{
		"SignatureMethod":  "HMAC-SHA1", // 以下 系统参数
		"SignatureNonce":   uuid.New(),
		"AccessKeyId":      s.AccessKeyID,
		"SignatureVersion": "1.0",
		"Timestamp":        time.Now().UTC().Format(time.RFC3339),
		"Format":           "JSON",

		"Action":        "SendSms", // 以下 业务API参数
		"Version":       "2017-05-25",
		"RegionId":      "cn-hangzhou",
		"PhoneNumbers":  strings.Join(req.Mobiles, ","),
		"SignName":      gou.EmptyTo(req.SignName, s.SignName),
		"TemplateParam": string(gou.JSON(req.TemplateParams)),
		"TemplateCode":  gou.EmptyTo(req.TemplateCode, s.TemplateCode),
		"OutID":         outID}
	str := "" // 3. 构造待签名的字符串

	gou.IterateMapSorted(param, func(k, v string) { str += "&" + enc(k) + "=" + enc(v) })

	toSign := "GET&" + enc("/") + "&" + enc(str[1:])
	logrus.Debugf("toSign:【%s】", toSign)

	param["Signature"] = gou.HmacSha1(toSign, s.AccessKeySecret+"&") // 4. 签名

	return param, outID
}

func enc(s string) string {
	s = url.QueryEscape(s)
	s = strings.Replace(s, "+", "%20", -1)
	s = strings.Replace(s, "*", "%2A", -1)

	return strings.Replace(s, "%7E", "~", -1)
}
