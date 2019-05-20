package api

import (
	"fmt"
	"github.com/bingoohuang/gou"
	"github.com/sirupsen/logrus"
)

// Dingtalk 表示钉钉消息发送器
type Dingtalk struct {
	AccessToken string `json:"accessToken"`
}

var _ Config = (*Dingtalk)(nil)

// Config 创建发送器，要求参数 config 是{accessToken}的格式
func (s *Dingtalk) Config(config string) error {
	s.AccessToken = config
	return nil
}

func (s *Dingtalk) InitMeaning() {
	s.AccessToken = "自定义机器人的accessToken"
}

type DingtalkReq struct {
	Message   string   `json:"message"`
	AtMobiles []string `json:"atMobiles"`
	AtAll     bool     `json:"atAll"`
}

func (s Dingtalk) NewRequest() interface{} {
	return &DingtalkReq{}
}

// Notify 发送信息
func (s Dingtalk) Notify(request interface{}) (interface{}, error) {
	req := request.(DingtalkReq)
	robot := Robot{Webhook: "https://oapi.dingtalk.com/robot/send?access_token=" + s.AccessToken}
	return robot.SendText(req.Message, req.AtMobiles, req.AtAll)
}

// api https://open-doc.dingtalk.com/microapp/serverapi2/qf2nxq
// 以下代码来自  https://github.com/royeo/dingrobot

// Roboter is the interface implemented by Robot that can send multiple types of messages.
type Roboter interface {
	SendText(content string, atMobiles []string, isAtAll bool) (*DingResponse, error)
	SendLink(title, text, messageURL, picURL string) (*DingResponse, error)
	SendMarkdown(title, text string, atMobiles []string, isAtAll bool) (*DingResponse, error)
	SendActionCard(title, text, singleTitle, singleURL, btnOrientation, hideAvatar string) (*DingResponse, error)
}

// Robot represents a dingtalk custom robot that can send messages to groups.
type Robot struct {
	Webhook string
}

// SendText send a text type message.
func (r Robot) SendText(content string, atMobiles []string, isAtAll bool) (*DingResponse, error) {
	return r.send(&textMessage{
		MsgType: msgTypeText,
		Text:    textParams{Content: content},
		At:      atParams{AtMobiles: atMobiles, IsAtAll: isAtAll},
	})
}

// SendLink send a link type message.
func (r Robot) SendLink(title, text, messageURL, picURL string) (*DingResponse, error) {
	return r.send(&linkMessage{
		MsgType: msgTypeLink,
		Link:    linkParams{Title: title, Text: text, MessageURL: messageURL, PicURL: picURL},
	})
}

// SendMarkdown send a markdown type message.
func (r Robot) SendMarkdown(title, text string, atMobiles []string, isAtAll bool) (*DingResponse, error) {
	return r.send(&markdownMessage{
		MsgType:  msgTypeMarkdown,
		Markdown: markdownParams{Title: title, Text: text},
		At:       atParams{AtMobiles: atMobiles, IsAtAll: isAtAll},
	})
}

// SendActionCard send a action card type message.
func (r Robot) SendActionCard(title, text, singleTitle, singleURL, btnOrientation, hideAvatar string) (*DingResponse, error) {
	return r.send(&actionCardMessage{
		MsgType:    msgTypeActionCard,
		ActionCard: actionCardParams{Title: title, Text: text, SingleTitle: singleTitle, SingleURL: singleURL, BtnOrientation: btnOrientation, HideAvatar: hideAvatar},
	})
}

type DingResponse struct {
	Errcode int    `json:"code"`
	Errmsg  string `json:"message"`
}

func (r Robot) send(msg interface{}) (*DingResponse, error) {
	var dr DingResponse
	_, err := gou.RestPost(r.Webhook, msg, &dr)
	if err != nil {
		return &dr, err
	}

	logrus.Infof("send:%+v", dr)

	if dr.Errcode != 0 {
		return &dr, fmt.Errorf("dingrobot send failed: %v", dr.Errmsg)
	}

	return &dr, nil
}

const (
	msgTypeText       = "text"
	msgTypeLink       = "link"
	msgTypeMarkdown   = "markdown"
	msgTypeActionCard = "actionCard"
)

type textMessage struct {
	MsgType string     `json:"msgtype"`
	Text    textParams `json:"text"`
	At      atParams   `json:"at"`
}

type textParams struct {
	Content string `json:"content"`
}

type atParams struct {
	AtMobiles []string `json:"atMobiles,omitempty"`
	IsAtAll   bool     `json:"isAtAll,omitempty"`
}

type linkMessage struct {
	MsgType string     `json:"msgtype"`
	Link    linkParams `json:"link"`
}

type linkParams struct {
	Title      string `json:"title"`
	Text       string `json:"text"`
	MessageURL string `json:"messageUrl"`
	PicURL     string `json:"picUrl,omitempty"`
}

type markdownMessage struct {
	MsgType  string         `json:"msgtype"`
	Markdown markdownParams `json:"markdown"`
	At       atParams       `json:"at"`
}

type markdownParams struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type actionCardMessage struct {
	MsgType    string           `json:"msgtype"`
	ActionCard actionCardParams `json:"actionCard"`
}

type actionCardParams struct {
	Title          string `json:"title"`
	Text           string `json:"text"`
	SingleTitle    string `json:"singleTitle"`
	SingleURL      string `json:"singleURL"`
	BtnOrientation string `json:"btnOrientation,omitempty"`
	HideAvatar     string `json:"hideAvatar,omitempty"`
}
