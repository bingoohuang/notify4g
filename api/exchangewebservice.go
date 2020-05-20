package api

import (
	"github.com/bingoohuang/gou/str"
	"github.com/mhewedy/ews"
	"github.com/sirupsen/logrus"
)

// ExchangeWebService is Exchange Web Service
type ExchangeWebService struct {
	EwsAddr  string `json:"ewsAddr"`
	Username string `json:"username"`
	Password string `json:"password" `
}

var _ Config = (*ExchangeWebService)(nil)

// Config 加载配置
func (q *ExchangeWebService) Config(config string) error {
	q.EwsAddr, q.Username, q.Password = str.Split3(config, "/", true, false)

	return nil
}

func (q *ExchangeWebService) InitMeaning() {
	q.EwsAddr = "https://outlook.office365.com/EWS/Exchange.asmx"
	q.Username = "email@exchangedomain"
	q.Password = "password"
}

func (q ExchangeWebService) NewRequest() Request { return &MailReq{} }
func (q ExchangeWebService) ChannelName() string { return exchangewebservice }

// Notify 发送邮件
func (q ExchangeWebService) Notify(_ *App, request Request) NotifyRsp {
	r := request.(*MailReq)

	c := ews.NewClient(q.EwsAddr, q.Username, q.Password,
		&ews.Config{Dump: true, NTLM: true, SkipTLS: true},
	)

	err := SendEWSEmail(c, r.To, r.Subject, r.Message)

	if err != nil {
		logrus.Warnf("ews to %v err %v: ", r.To, err)
	}

	return MakeRsp(err, err == nil, q.ChannelName(), nil)
}

// SendEWSEmail helper method to send Message
func SendEWSEmail(c ews.Client, to []string, subject, body string) error {
	m := ews.Message{
		ItemClass: "IPM.Note",
		Subject:   subject,
		Body: ews.Body{
			BodyType: "HTML",
			Body:     []byte(body),
		},
		Sender: ews.OneMailbox{
			Mailbox: ews.Mailbox{
				EmailAddress: c.GetUsername(),
			},
		},
	}

	mb := make([]ews.Mailbox, len(to))

	for i, addr := range to {
		mb[i].EmailAddress = addr
	}

	m.ToRecipients.Mailbox = mb

	return ews.CreateMessageItem(c, m)
}
