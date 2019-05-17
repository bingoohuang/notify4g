package notify4g

import (
	"github.com/bingoohuang/gou"
	"github.com/sirupsen/logrus"
	"github.com/tobyzxj/uuid"
	"net/url"

	"strings"
	"time"
)

// AliyunSms 表示阿里云短信发送器
type AliyunSms struct {
	AccessKeyId     string `json:"accessKeyID"`
	AccessKeySecret string `json:"acessKeySecret"`
	TemplateCode    string `json:"templateCode"`
	SignName        string `json:"signName"`
}

// LoadConfig 创建发送器，要求参数 config 是{accessKeyId}/{accessKeySecret}/{templateCode}/{signName}的格式
func (s *AliyunSms) LoadConfig(config string) error {
	s.AccessKeyId, s.AccessKeySecret, s.TemplateCode, s.SignName = gou.Split4(config, "/", true, false)
	return nil
}

type AliyunSmsReq struct {
	TemplateCode   string            `json:"templateCode"`
	TemplateParams map[string]string `json:"templateParams"`
	SignName       string            `json:"signName"`
	Mobiles        []string          `json:"mobiles"`
}

type AliyunSmsRsp struct {
	OutId string `json:"outId"`

	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"requestID"`
	BizId     string `json:"bizID"`
}

type RawAliyunSmsRsp struct {
	Code      string // eg. OK 请求状态码。 返回OK代表请求成功。 其他错误码详见错误码列表。
	Message   string // eg. OK 状态码的描述。
	RequestId string // eg. F655A8D5-B967-440B-8683-DAD6FF8DE990	 请求ID。
	BizId     string // eg. 900619746936498440^0,  发送回执ID，可根据该ID在接口QuerySendDetails中查询具体的发送状态。
}

// Notify 发送短信
func (s AliyunSms) Notify(req AliyunSmsReq) (*AliyunSmsRsp, error) {
	param, outId := s.createParams(req)
	u, _ := gou.BuildURL("http://dysmsapi.aliyuncs.com/", param)

	var r RawAliyunSmsRsp
	err := gou.RestGet(u, &r)
	if err != nil {
		logrus.Warnf("RestGet fail on url %s, error %v", u, err)
		return nil, err
	}

	smsRsp := &AliyunSmsRsp{OutId: outId, Code: r.Code, Message: r.Message, RequestId: r.RequestId, BizId: r.BizId}
	return smsRsp, err
}

// api doc: https://help.aliyun.com/document_detail/101414.html?spm=a2c4g.11186623.6.616.1eee202a1PxPlf
func (s AliyunSms) createParams(req AliyunSmsReq) (map[string]string, string) {
	outId := gou.RandomString(16)
	param := map[string]string{
		"SignatureMethod":  "HMAC-SHA1", // 以下 系统参数
		"SignatureNonce":   uuid.New(),
		"AccessKeyId":      s.AccessKeyId,
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
		"OutId":         outId}

	str := "" // 3. 构造待签名的字符串
	gou.IterateMapSorted(param, func(k, v string) { str += "&" + enc(k) + "=" + enc(v) })

	toSign := "GET&" + enc("/") + "&" + enc(str[1:])
	logrus.Debugf("toSign:【%s】", toSign)

	param["Signature"] = gou.HmacSha1(toSign, s.AccessKeySecret+"&") // 4. 签名
	return param, outId
}

func enc(s string) string {
	s = url.QueryEscape(s)
	s = strings.Replace(s, "+", "%20", -1)
	s = strings.Replace(s, "*", "%2A", -1)
	return strings.Replace(s, "%7E", "~", -1)
}
