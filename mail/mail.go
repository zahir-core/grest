package slack

import (
	"gopkg.in/gomail.v2"
)

var (
	dialer *gomail.Dialer
)

func Configure(smtpHost string, smtpPort int, smtpUser, smtpPassword string) {
	dialer = gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPassword)
}

func New(s ...gomail.MessageSetting) *gomail.Message {
	return gomail.NewMessage(s...)
}

func Send(m *gomail.Message) error {
	return dialer.DialAndSend(m)
}
