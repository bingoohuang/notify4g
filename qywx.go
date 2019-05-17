package notify4g

import (
	"encoding/json"
	"errors"
	"github.com/bingoohuang/gou"
	"github.com/sirupsen/logrus"
	"strings"
)

// QywxTokenResult 表示企业微信的令牌结果
type QywxTokenResult struct {
	ErrCode          int    `json:"errcode"`
	ErrMsg           string `json:"errmsg"`
	AccessToken      string `json:"access_token"`
	ExpiresInSeconds int    `json:"expires_in"`
}

// GetQywxAccessToken 获得企业微信的令牌
func GetQywxAccessToken(corpID, corpSecret string) (string, error) {
	url := "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=" + corpID + "&corpsecret=" + corpSecret
	logrus.Debugf("url:%s", url)
	resp, err := gou.UrlGet(url)
	logrus.Debugf("resp:%+v, err:%+v", resp, err)
	if err != nil {
		return "", err
	}

	var tokenResult QywxTokenResult
	if err := resp.ToJson(&tokenResult); err != nil {
		return "", err
	}

	if tokenResult.ErrCode == 0 {
		return tokenResult.AccessToken, nil
	}

	return "", errors.New(tokenResult.ErrMsg)
}

// https://qydev.weixin.qq.com/wiki/index.php?title=发送接口说明
// SendQywxMsg 发送企业微信消息
func SendQywxMsg(accessToken, agentID, content string, userIds []string) ([]byte, error) {
	touser := strings.Join(userIds, "|")
	msg := map[string]interface{}{
		"touser": touser, "msgtype": "text", "agentid": agentID, "safe": 0,
		"text": map[string]string{"content": content}}
	ret, err := gou.RestPost("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token="+accessToken, msg, nil)
	return ret, err
}

// FastSendQywxMsg 快速发送企业微信消息
func FastSendQywxMsg(corpID, corpSecret, agentID, content string, userIds []string) ([]byte, error) {
	token, err := GetQywxAccessToken(corpID, corpSecret)
	if err != nil {
		return nil, err
	}

	return SendQywxMsg(token, agentID, content, userIds)
}

// Qywx 表示企业微信消息发送器
type Qywx struct {
	CorpID     string `json:"corpID"`
	CorpSecret string `json:"corpSecret"`
	AgentID    string `json:"agentID"`
}

// LoadConfig 创建发送器，要求参数 config 是{corpID}/{corpSecret}/{agentID}的格式
func (s *Qywx) LoadConfig(config string) error {
	s.CorpID, s.CorpSecret, s.AgentID = gou.Split3(config, "/", true, false)

	return nil
}

type QywxReq struct {
	Msg     string   `json:"msg"`
	UserIds []string `json:"userIds"`
}

// Notify 发送企业消息
func (s Qywx) Notify(r QywxReq) (interface{}, error) {
	result, err := FastSendQywxMsg(s.CorpID, s.CorpSecret, s.AgentID, r.Msg, r.UserIds)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(result), nil
}
