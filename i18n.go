package grest

import (
	"strings"
)

// The Accept-Language request HTTP header indicates the natural language and locale that the client prefers.
//
// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Accept-Language
const LangHeader = "Accept-Language"

// Translator handles translation of messages based on language keys.
type Translator struct {
	i18n map[string]map[string]string
}

// AddTranslation adds translation data for a specific language.
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

// GetTranslation retrieves translation data for a specific language.
func (t *Translator) GetTranslation(lang string) map[string]string {
	if msg, ok := t.i18n[lang]; ok {
		return msg
	}
	return map[string]string{}
}

// SupportedLanguage determines the most appropriate supported language based on client's preference.
func (t Translator) SupportedLanguage(lang string) string {
	lang = strings.Split(lang, ";")[0]
	if lang == "" || lang == "*" {
		lang = "en"
	}
	supportedLang := ""
	langs := strings.Split(lang, ",")
	for _, lg := range langs {
		_, ok := t.i18n[lg]
		if supportedLang == "" {
			if ok {
				supportedLang = lg
			} else {
				for k := range t.i18n {
					if strings.HasPrefix(k, lg) {
						supportedLang = k
					}
				}
			}
		}
	}
	return supportedLang
}

// Trans translates a message based on the language and translation key, with optional parameter replacement.
func (t Translator) Trans(lang, key string, params ...map[string]string) string {
	lang = t.SupportedLanguage(lang)

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
