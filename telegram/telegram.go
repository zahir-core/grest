package telegram

import (
	"grest.dev/grest"
)

var (
	baseUrl  = "https://api.telegram.org/bot"
	botToken = ""
	chatId   = ""
)

type Message struct {
	ChatID  string
	Message string
}

func Configure(token, chat_id string) {
	botToken = token
	chatId = chat_id
}

func New() *Message {
	return &Message{}
}

func (m *Message) AddMessage(message string) {
	m.Message = message
}

func (m *Message) Send() error {
	if m.ChatID == "" {
		m.ChatID = chatId
	}
	c := grest.NewHttpClient("POST", baseUrl+botToken+"/sendMessage")
	c.AddJsonBody(map[string]interface{}{
		"chat_id":    m.ChatID,
		"parse_mode": "MarkdownV2",
		"text":       m.Message,
	})
	_, err := c.Send()
	return err
}

func SendAlert(message string) error {
	c := grest.NewHttpClient("POST", baseUrl+botToken+"/sendMessage")
	c.AddJsonBody(map[string]interface{}{
		"chat_id":    chatId,
		"parse_mode": "MarkdownV2",
		"text":       message,
	})
	_, err := c.Send()
	return err
}
