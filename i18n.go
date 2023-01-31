package grest

import (
	"strings"
)

// see: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Language
const LangHeader = "Content-Language"

type TranslatorInterface interface {
	AddTranslation(lang string, messages map[string]string)
	GetTranslation(lang string) map[string]string
	Trans(lang, key string, params ...map[string]string) string
}

type Translator struct {
	i18n map[string]map[string]string
}

// AddTranslation add translation data based on language key
func (t *Translator) AddTranslation(lang string, messages map[string]string) {
	if t.i18n == nil {
		t.i18n = map[string]map[string]string{}
	}
	if msg, ok := t.i18n[lang]; ok {
		for k, v := range messages {
			msg[k] = v
		}
		t.i18n[lang] = msg
	} else {
		t.i18n[lang] = messages
	}
}

// GetTranslation get translation data based on language key
func (t *Translator) GetTranslation(lang string) map[string]string {
	if msg, ok := t.i18n[lang]; ok {
		return msg
	}
	return map[string]string{}
}

// Translate message data based on language and translation key
func (t Translator) Trans(lang, key string, params ...map[string]string) string {
	message := key
	msg := map[string]string{}
	if val, ok := t.i18n[lang]; ok {
		msg = val
	} else {
		for k, v := range t.i18n {
			if strings.HasPrefix(k, lang) {
				msg = v
			}
		}
	}
	if val, ok := msg[key]; ok {
		message = val
	}
	if len(params) > 0 {
		for k, v := range params[0] {
			message = strings.ReplaceAll(message, ":"+k, v)
		}
	}
	return message
}
