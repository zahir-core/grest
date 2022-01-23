package telegram

import (
	"grest.dev/grest/httpclient"
)

var (
	baseUrl  = "https://api.telegram.org/bot"
	botToken = ""
	chatId   = ""
)

func Configure(token, chat_id string) {
	botToken = token
	chatId = chat_id
}

func SendAlert(message string) error {
	c := httpclient.New("POST", baseUrl+botToken+"/sendMessage")
	c.AddJsonBody(map[string]interface{}{
		"chat_id":    chatId,
		"parse_mode": "MarkdownV2",
		"text":       message,
	})
	_, err := c.Send()
	return err
}
