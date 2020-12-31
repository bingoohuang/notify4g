package api

import (
	"encoding/json"
	"strings"

	"github.com/bingoohuang/gou/enc"
	"github.com/bingoohuang/gou/str"

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
	s.AppKey, s.AppSecret, s.TemplateCode, s.SignName = str.Split4(config, "/", true, false)
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

func (a *AliyunDaYuSmsReq) FilterRedList(list redList) bool {
	a.Mobiles = list.FilterMobiles(a.Mobiles)

	return len(a.Mobiles) > 0
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

func (s AliyunDaYuSms) NewRequest() Request { return &AliyunDaYuSmsReq{} }
func (s AliyunDaYuSms) ChannelName() string { return aliyundayusms }

// Notify 发送短信 https://api.alidayu.com/doc2/apiDetail?apiId=25450
func (s AliyunDaYuSms) Notify(_ *App, request Request) NotifyRsp {
	req := request.(*AliyunDaYuSmsReq)
	client := NewTopClient(s.AppKey, s.AppSecret)
	param := s.createParams(req)
	response, err := client.Execute(param)
	logrus.Infof("request:%+v,response:【%+v】,error:【%+v】", req, response, err)

	var rsp AliyunDaYuSmsRsp
	err = json.Unmarshal([]byte(response), &rsp)

	return MakeRsp(err, rsp.Code == 0, s.ChannelName(), response)
}

var _ SmsNotifier = (*AliyunDaYuSms)(nil)

func (s AliyunDaYuSms) ConvertRequest(r *SmsReq) Request {
	return &AliyunDaYuSmsReq{TemplateParams: r.TemplateParams, Mobiles: r.Mobiles}
}

func (s AliyunDaYuSms) createParams(r *AliyunDaYuSmsReq) *AlibabaAliqinFcSmsNumSendRequest {
	req := NewAlibabaAliqinFcSmsNumSendRequest()
	req.SmsFreeSignName = str.EmptyThen(r.SignName, s.SignName)
	req.RecNum = strings.Join(r.Mobiles, ",")
	req.SmsTemplateCode = str.EmptyThen(r.TemplateCode, s.TemplateCode)
	req.SmsParam = enc.JSON(r.TemplateParams)

	return req
}
