package grest

import (
	"strings"
)

// see: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Language
const LangHeader = "Content-Language"

var I18n = map[string]map[string]string{
	"en-US": {
		"bad_request":    "The request cannot be performed because of malformed or missing parameters.",
		"forbidden":      "The user does not have permission to access the resource.",
		"internal_error": "Failed to connect to the server, please try again later.",
		"unauthorized":   "Invalid authentication token. Please Re-Login",
	},
	"id-ID": {
		"bad_request":    "Permintaan tidak dapat dilakukan karena ada parameter yang salah atau tidak lengkap.",
		"forbidden":      "Pengguna tidak memiliki izin untuk mengakses data.",
		"internal_error": "Gagal terhubung ke server, silakan coba lagi nanti.",
		"unauthorized":   "Token otentikasi tidak valid. Silahkan logout dan login ulang",
	},
}

func AddTranslation(lang string, message map[string]string) {
	if msg, ok := I18n[lang]; ok {
		for k, v := range message {
			msg[k] = v
		}
		I18n[lang] = msg
	} else {
		I18n[lang] = message
	}
}

func Trans(lang, key string, params ...map[string]string) string {
	message := key
	msg := map[string]string{}
	if val, ok := I18n[lang]; ok {
		msg = val
	} else {
		for k, v := range I18n {
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
