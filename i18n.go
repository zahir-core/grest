package grest

import (
	"strings"
)

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Language
var i18n = map[string]map[string]string{
	"en-US": {
		"bad_request":  "The request cannot be performed because of malformed or missing parameters.",
		"forbidden":    "The user does not have permission to access the resource.",
		"unauthorized": "Request was rejected because of invalid or expired Authentication Token.",
	},
	"id-ID": {
		"bad_request":  "The request cannot be performed because of malformed or missing parameters.",
		"forbidden":    "The user does not have permission to access the resource.",
		"unauthorized": "Request was rejected because of invalid or expired Authentication Token.",
	},
}

func AddTranslation(lang string, message map[string]string) {
	if msg, ok := i18n[lang]; ok {
		for k, v := range message {
			msg[k] = v
		}
		i18n[lang] = msg
	} else {
		i18n[lang] = message
	}
}

func Trans(lang, key string, params ...map[string]string) string {
	message := key
	msg := map[string]string{}
	if val, ok := i18n[lang]; ok {
		msg = val
	} else {
		for k, v := range i18n {
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
