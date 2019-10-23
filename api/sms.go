package api

import (
	"fmt"

	"github.com/bingoohuang/gou"
)

// Sms 表示聚合短信发送器
type Sms struct {
	ConfigIds []string `json:"configIds"`
	Random    bool     `json:"random"`          // 是否在发送配置中随机发送，不随机时，按照配置顺序发送
	Retry     int      `json:"retry" faker:"-"` // 在发送配置中，重试n次
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
	Retry          int               `json:"retry" faker:"-"` // 在发送配置中，重试n次, -1表示使用默认配置
}

func (q *SmsReq) FilterRedList(list redList) bool {
	q.Mobiles = list.FilterMobiles(q.Mobiles)

	return len(q.Mobiles) > 0
}

func (r Sms) NewRequest() Request { return &SmsReq{} }
func (r Sms) ChannelName() string { return sms }

type SmsNotifier interface {
	ConvertRequest(*SmsReq) Request
}

const BreakIterating = true
const ContinueIterating = false

// Notify 发送短信
func (r Sms) Notify(app *App, request Request) NotifyRsp {
	req := request.(*SmsReq)

	retry := 0
	maxRetry := r.maxRetryTimes(req)

	var err error

	succ := false
	channelName := ""
	rsp := NotifyRsp{}

	f := func(configID string) bool {
		nc := app.configCache.Read(configID)
		if nc == nil {
			err = fmt.Errorf("configID %s not found", configID)
			return BreakIterating
		}

		var smsNotifier SmsNotifier

		ok := false

		if smsNotifier, ok = nc.Config.(SmsNotifier); !ok {
			err = fmt.Errorf("configID %s not a SmsNotifier config", configID)
			return BreakIterating
		}

		channelName = nc.Config.ChannelName()
		r := smsNotifier.ConvertRequest(req)
		rsp = nc.Config.Notify(app, r)

		if rsp.Status == 200 {
			succ = true
			return BreakIterating
		}

		if retry < maxRetry {
			retry++
			return ContinueIterating
		}

		return BreakIterating
	}

	gou.IterateSlice(r.ConfigIds, r.startIndex(), f)

	return MakeRsp(err, succ, channelName, rsp)
}

func (r Sms) maxRetryTimes(req *SmsReq) int {
	retryTimes := req.Retry
	if retryTimes < 0 {
		retryTimes = r.Retry
	}

	return retryTimes
}

func (r Sms) startIndex() int {
	if r.Random {
		return gou.RandomIntN(uint64(len(r.ConfigIds)))
	}

	return 0
}
