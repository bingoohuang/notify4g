package api

import (
	"encoding/json"
	"strings"

	"github.com/bingoohuang/gou"
	"github.com/sirupsen/logrus"
)

// AliyunDaYuSms 表示阿里云大于短信发送器
type AliyunDaYuSms struct {
	AppKey       string `json:"appKey"`
	AppSecret    string `json:"appSecret"`
	TemplateCode string `json:"templateCode"`
	SignName     string `json:"signName" faker:"-"`
}

var _ Config = (*AliyunDaYuSms)(nil)

// Config 创建发送器，要求参数 config 是{AppKey}/{AppSecret}/{TemplateCode}/{SignName}的格式
func (s *AliyunDaYuSms) Config(config string) error {
	s.AppKey, s.AppSecret, s.TemplateCode, s.SignName = gou.Split4(config, "/", true, false)
	return nil
}

func (s *AliyunDaYuSms) InitMeaning() {
	s.AppKey = "appKey"
	s.AppSecret = "appSecret"
	s.TemplateCode = "短信模板ID，可以不设置，然后在发送时再设置"
	s.SignName = "短信模签名，不设置使用默认签名，或者在发送时再设置"
}

type AliyunDaYuSmsReq struct {
	TemplateCode   string            `json:"templateCode" faker:"-"`
	TemplateParams map[string]string `json:"templateParams"`
	SignName       string            `json:"signName" faker:"-"`
	Mobiles        []string          `json:"mobiles" faker:"china_mobile_number"`
}

type AliyunDaYuSmsRsp struct {
	Response map[string]interface{} `json:"alibaba_aliqin_fc_sms_num_send_response"`
	AliyunDaYuSmsErrorRsp
}

type AliyunDaYuSmsErrorRsp struct {
	SubMsg  string `json:"sub_msg"`
	Code    int    `json:"code"`
	SubCode string `json:"sub_code"`
	Msg     string `json:"msg"`
}

func (s AliyunDaYuSms) NewRequest() interface{} { return &AliyunDaYuSmsReq{} }
func (s AliyunDaYuSms) ChannelName() string     { return aliyundayusms }

// Notify 发送短信 https://api.alidayu.com/doc2/apiDetail?apiId=25450
func (s AliyunDaYuSms) Notify(app *App, request interface{}) NotifyRsp {
	req := request.(*AliyunDaYuSmsReq)
	client := NewTopClient(s.AppKey, s.AppSecret)
	param := s.createParams(req)
	response, err := client.Execute(param)
	logrus.Debugf("response:【%+v】,error:【%+v】", response, err)

	var rsp AliyunDaYuSmsRsp
	err = json.Unmarshal([]byte(response), &rsp)

	return MakeRsp(err, rsp.Code == 0, s.ChannelName(), response)
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
