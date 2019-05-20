package api

import (
	"github.com/bingoohuang/gou"
	"gopkg.in/gomail.v2"

	"strconv"
)

// Mail 表示邮件发送器
type Mail struct {
	SmtpAddr string `json:"smtpAddr"` // smtp.gmail.com
	SmtpPort int    `json:"smtpPort"` // 587
	From     string `json:"from"`     // ...@gmail.com
	Username string `json:"username"` // ...
	Pass     string `json:"pass"`     // ...
}

var _ Config = (*Mail)(nil)

// Config 加载配置
func (q *Mail) Config(config string) error {
	var port string
	q.SmtpAddr, port, q.From, q.Username, q.Pass = gou.Split5(config, "/", true, false)
	q.SmtpPort, _ = strconv.Atoi(port)
	return nil
}

func (q *Mail) InitMeaning() {
	q.SmtpAddr = "SMTP地址"
	q.SmtpPort = 587       // 587
	q.From = `发送人地址`       // ...@gmail.com
	q.Username = `邮箱登录用户名` // ...
	q.Pass = `邮箱登录密码`      // ...
}

type MailReq struct {
	Subject string   `json:"subject"`
	Message string   `json:"message"`
	To      []string `json:"to"`
}

func (q Mail) NewRequest() interface{} {
	return &MailReq{}
}

// Notify 发送邮件
func (q Mail) Notify(request interface{}) (interface{}, error) {
	r := request.(MailReq)

	mm := gomail.NewMessage()
	mm.SetHeader("From", q.From)
	mm.SetHeader("To", r.To...)
	mm.SetHeader("Subject", r.Subject)
	mm.SetBody("text/plain", r.Message)

	d := gomail.NewDialer(q.SmtpAddr, q.SmtpPort, q.Username, q.Pass)

	// Notify the email to Bob, Cora and Dan.
	return nil, d.DialAndSend(mm)
}
