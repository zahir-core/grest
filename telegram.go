package grest

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

var (
	TelegramBaseURL       = "https://api.telegram.org"
	TelegramBotToken      = ""
	TelegramDefaultChatID = ""
)

type Telegram struct {
	ChatID      string
	ParseMode   string
	Text        string
	Photo       *multipart.FileHeader
	Audio       *multipart.FileHeader
	Video       *multipart.FileHeader
	Document    *multipart.FileHeader
	ReplyMarkup interface{}
}

func (t Telegram) Send() error {
	if t.ChatID == "" {
		t.ChatID = TelegramDefaultChatID
	}
	if t.ParseMode == "" {
		t.ParseMode = "MarkdownV2"
	}
	endPoint := "sendMessage"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("chat_id", t.ChatID)
	writer.WriteField("parse_mode", t.ParseMode)
	writer.WriteField("text", t.Text)
	if t.Photo != nil {
		f, err := t.Photo.Open()
		if err == nil {
			key := "photo"
			endPoint = "sendPhoto"
			if t.Photo.Size > 10485760 {
				key = "document"
				endPoint = "sendDocument"
			}
			part, err := writer.CreateFormFile(key, t.Photo.Filename)
			if err != nil {
				return NewError(http.StatusInternalServerError, err.Error())
			}
			_, err = io.Copy(part, f)
			if err != nil {
				return NewError(http.StatusInternalServerError, err.Error())
			}
		}
	} else if t.Audio != nil {
		f, err := t.Audio.Open()
		if err == nil {
			key := "audio"
			endPoint = "sendAudio"
			if t.Audio.Size > 52428800 {
				key = "document"
				endPoint = "sendDocument"
			}
			part, err := writer.CreateFormFile(key, t.Audio.Filename)
			if err != nil {
				return NewError(http.StatusInternalServerError, err.Error())
			}
			_, err = io.Copy(part, f)
			if err != nil {
				return NewError(http.StatusInternalServerError, err.Error())
			}
		}
	} else if t.Video != nil {
		f, err := t.Video.Open()
		if err == nil {
			key := "video"
			endPoint = "sendVideo"
			if t.Video.Size > 52428800 {
				key = "document"
				endPoint = "sendDocument"
			}
			part, err := writer.CreateFormFile(key, t.Video.Filename)
			if err != nil {
				return NewError(http.StatusInternalServerError, err.Error())
			}
			_, err = io.Copy(part, f)
			if err != nil {
				return NewError(http.StatusInternalServerError, err.Error())
			}
		}
	} else if t.Document != nil {
		f, err := t.Document.Open()
		if err == nil {
			key := "Document"
			endPoint = "sendDocument"
			part, err := writer.CreateFormFile(key, t.Document.Filename)
			if err != nil {
				return NewError(http.StatusInternalServerError, err.Error())
			}
			_, err = io.Copy(part, f)
			if err != nil {
				return NewError(http.StatusInternalServerError, err.Error())
			}
		}
	}
	if t.ReplyMarkup != nil {
		rm, err := json.Marshal(t.ReplyMarkup)
		if err == nil {
			writer.WriteField("reply_markup", string(rm))
		}
	}
	err := writer.Close()
	if err != nil {
		return NewError(http.StatusInternalServerError, err.Error())
	}

	req, err := http.NewRequest("POST", TelegramBaseURL+"/"+TelegramBotToken+"/"+endPoint, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return NewError(http.StatusInternalServerError, err.Error())
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return NewError(http.StatusInternalServerError, err.Error())
		}
		r := map[string]interface{}{}
		json.Unmarshal(b, &r)
		msg, _ := r["description"].(string)
		return NewError(http.StatusInternalServerError, msg, r)
	}

	return nil
}
