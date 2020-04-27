package api

import "net/http"

type NotifyRsp struct {
	Status  int           `json:"status"`  // 状态,0成功
	Message string        `json:"message"` // 消息
	Data    NotifyRspData `json:"data"`    // 原始返回
}

type NotifyRspData struct {
	Channel string      `json:"channel"` // 发送渠道
	Raw     interface{} `json:"raw"`
}

func MakeRsp(err error, ok bool, channel string, raw interface{}) NotifyRsp {
	status := 400
	msg := ""

	if ok {
		status = http.StatusOK
		msg = "OK"
	}

	if err != nil {
		msg = err.Error()
	}

	if r, ok := raw.(NotifyRsp); ok {
		raw = r.Data.Raw
	}

	return NotifyRsp{Status: status, Message: msg, Data: NotifyRspData{Channel: channel, Raw: raw}}
}
