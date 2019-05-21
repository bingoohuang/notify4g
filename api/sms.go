package api

import (
	"fmt"
	"github.com/bingoohuang/gou"
)

// Sms 表示聚合短信发送器
type Sms struct {
	ConfigIds   []string           `json:"configIds"`
	Random      bool               `json:"random"` // 是否在发送配置中随机发送，不随机时，按照配置顺序发送
	Retry       int                `json:"retry"`  // 在发送配置中，重试n次
	ConfigCache *NotifyConfigCache `json:"-"`
}

var _ Config = (*Sms)(nil)

// Config 创建发送器，要求参数 config 是{accessKeyId}/{accessKeySecret}/{templateCode}/{signName}的格式
func (r *Sms) Config(config string) error {
	r.ConfigIds = gou.SplitN(config, "/", true, true)
	return nil
}

func (r *Sms) InitMeaning() {
	r.ConfigIds = []string{"阿里云短信配置ID", "腾讯云短信配置ID"}
}

type SmsReq struct {
	TemplateParams map[string]string `json:"templateParams"`
	Mobiles        []string          `json:"mobiles" faker:"china_mobile_number"`
	Retry          int               `json:"retry"` // 在发送配置中，重试n次, -1表示使用默认配置
}

func (r Sms) NewRequest() interface{} {
	return &SmsReq{}
}

type SmsNotifier interface {
	ConvertRequest(*SmsReq) interface{}
	NotifySms(request interface{}) (interface{}, bool, error)
}

// Notify 发送短信
func (r Sms) Notify(request interface{}) (interface{}, error) {
	req := request.(*SmsReq)

	retryTimes := req.Retry
	if retryTimes < 0 {
		retryTimes = r.Retry
	}

	retry := 0
	start := 0
	if r.Random {
		start = gou.RandomIntN(uint64(len(r.ConfigIds)))
	}

	var err error
	var rsp interface{}

	gou.IterateSlice(r.ConfigIds, start, func(configID string) bool {
		notifyConfig := r.ConfigCache.Read(configID)
		if notifyConfig == nil {
			err = fmt.Errorf("configID %s not found", configID)
			return true
		}

		var yes bool
		if smsNotifier, ok := notifyConfig.Config.(SmsNotifier); !ok {
			err = fmt.Errorf("configID %s not a SmsNotifier config", configID)
			return true
		} else if rsp, yes, err = smsNotifier.NotifySms(smsNotifier.ConvertRequest(req)); yes {
			return true
		}

		if retry < retryTimes {
			retry++
			return false
		} else {
			return true
		}
	})

	return nil, err
}
