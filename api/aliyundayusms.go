package api

import (
	"github.com/bingoohuang/gou"
	"github.com/sirupsen/logrus"
	"strings"
)

// AliyunSms 表示阿里云短信发送器
type AliyunDaYuSms struct {
	AppKey       string `json:"appKey"`
	AppSecret    string `json:"appSecret"`
	TemplateCode string `json:"templateCode"`
	SignName     string `json:"signName" faker:"-"`
}

var _ Config = (*AliyunDaYuSms)(nil)

// Config 创建发送器，要求参数 config 是{accessKeyId}/{accessKeySecret}/{templateCode}/{signName}的格式
func (s *AliyunDaYuSms) Config(config string) error {
	s.AppKey, s.AppSecret, s.TemplateCode, s.SignName = gou.Split4(config, "/", true, false)
	return nil
}

func (s *AliyunDaYuSms) InitMeaning() {
	s.AppKey = "accessKeyID"
	s.AppSecret = "acessKeySecret"
	s.TemplateCode = "短信模板ID，可以不设置，然后在发送时再设置"
	s.SignName = "短信模签名，不设置使用默认签名，或者在发送时再设置"
}

type AliyunDaYuSmsReq struct {
	TemplateCode   string            `json:"templateCode" faker:"-"`
	TemplateParams map[string]string `json:"templateParams"`
	SignName       string            `json:"signName" faker:"-"`
	Mobiles        []string          `json:"mobiles" faker:"china_mobile_number"`
}

func (s AliyunDaYuSms) NewRequest() interface{} { return &AliyunDaYuSmsReq{} }
func (s AliyunDaYuSms) ChannelName() string     { return aliyundayusms }

// Notify 发送短信
func (s AliyunDaYuSms) Notify(app *App, request interface{}) NotifyRsp {
	req := request.(*AliyunDaYuSmsReq)
	client := NewTopClient(s.AppKey, s.AppSecret)
	outID := gou.RandomString(16)
	param := s.createParams(req)
	response, err := client.Execute(param)
	r := AliyunSmsRsp{OutID: outID}
	if err != nil {
		logrus.Warnf("NotifyErr:【%s】", err.Error())
		return MakeRsp(err, r.Code == "ERROR", s.ChannelName(), r)
	}
	logrus.Debugf("response:【%+v】", response)
	return MakeRsp(err, r.Code == "OK", s.ChannelName(), r)
}

var _ SmsNotifier = (*AliyunDaYuSms)(nil)

func (s AliyunDaYuSms) ConvertRequest(r *SmsReq) interface{} {
	return &AliyunDaYuSmsReq{TemplateParams: r.TemplateParams, Mobiles: r.Mobiles}
}

func (s AliyunDaYuSms) createParams(r *AliyunDaYuSmsReq) *AlibabaAliqinFcSmsNumSendRequest {
	req := NewAlibabaAliqinFcSmsNumSendRequest()
	req.SmsFreeSignName = gou.EmptyTo(r.SignName, s.SignName)
	req.RecNum = strings.Join(r.Mobiles, ",")
	req.SmsTemplateCode = gou.EmptyTo(r.TemplateCode, s.TemplateCode)
	req.SmsParam = string(gou.JSON(r.TemplateParams))
	return req
}
