package notify4g

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

// LoadConfig 加载配置
func (q *Mail) LoadConfig(config string) error {
	var port string
	q.SmtpAddr, port, q.From, q.Username, q.Pass = gou.Split5(config, "/", true, false)
	q.SmtpPort, _ = strconv.Atoi(port)
	return nil
}

type MailReq struct {
	Subject string   `json:"subject"`
	Message string   `json:"message"`
	To      []string `json:"to"`
}

// Notify 发送邮件
func (q Mail) Notify(r MailReq) (interface{}, error) {
	mm := gomail.NewMessage()
	mm.SetHeader("From", q.From)
	mm.SetHeader("To", r.To...)
	mm.SetHeader("Subject", r.Subject)
	mm.SetBody("text/plain", r.Message)

	d := gomail.NewDialer(q.SmtpAddr, q.SmtpPort, q.Username, q.Pass)

	// Notify the email to Bob, Cora and Dan.
	return nil, d.DialAndSend(mm)
}
