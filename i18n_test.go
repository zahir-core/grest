package grest

import (
	"testing"
)

func TestTranslateSimple(t *testing.T) {
	lang := "en-US"
	key := "translation_key1"
	message := "translation testing"
	tr := &Translator{}
	tr.AddTranslation(lang, map[string]string{
		key: message,
	})
	msg := tr.Trans(lang, key)
	if message != msg {
		t.Errorf("Expected message [%v], got [%v]", message, msg)
	}
}

func TestTranslateWithShortLangKey(t *testing.T) {
	lang := "en-US"
	key := "translation_key2"
	message := "translation testing"
	tr := &Translator{}
	tr.AddTranslation(lang, map[string]string{
		key: message,
	})
	msg := tr.Trans("en", key)
	if message != msg {
		t.Errorf("Expected message [%v], got [%v]", message, msg)
	}
}

func TestTranslateWithNotExistsLangOrKey(t *testing.T) {
	key := "translation_key3"
	tr := &Translator{}
	msg := tr.Trans("ar", key)
	if key != msg {
		t.Errorf("Expected message [%v], got [%v]", key, msg)
	}
	msg2 := tr.Trans("en-US", key)
	if key != msg2 {
		t.Errorf("Expected message [%v], got [%v]", key, msg2)
	}
}

func TestTranslateWithParams(t *testing.T) {
	lang := "en-US"
	key := "translation_key4"
	message := "translation testing param1 = :param1, param2 = :param2, param3 = :param3"
	expectedMessage := "translation testing param1 = foo, param2 = bar, param3 = baz"
	tr := &Translator{}
	tr.AddTranslation(lang, map[string]string{
		key: message,
	})
	msg := tr.Trans(lang, key, map[string]string{"param1": "foo", "param2": "bar", "param3": "baz"})
	if expectedMessage != msg {
		t.Errorf("Expected message [%v], got [%v]", expectedMessage, msg)
	}
}
