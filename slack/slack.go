package slack

import (
	"errors"

	"grest.dev/grest"
)

// https://api.slack.com/messaging/webhooks
var (
	webhookUrls = map[string]string{}
	chatID      = ""
)

type Message struct {
	ChatID  string
	Message string
}

func Configure(webhookUrl string, cid ...string) {
	if len(cid) > 0 {
		if chatID == "" {
			chatID = cid[0]
		}
		webhookUrls[cid[0]] = webhookUrl
	} else {
		chatID = "default"
		webhookUrls[chatID] = webhookUrl
	}
}

func New() *Message {
	return &Message{}
}

func (m *Message) AddMessage(message string) {
	m.Message = message
}

func (m *Message) Send() error {
	if m.ChatID == "" {
		m.ChatID = chatID
	}
	webhookUrl, ok := webhookUrls[m.ChatID]
	if !ok {
		return errors.New("Slack webhook URL for " + m.ChatID + " is not found")
	}
	c := grest.NewHttpClient("POST", webhookUrl)
	c.AddJsonBody(map[string]string{"text": m.Message})
	_, err := c.Send()
	return err
}
