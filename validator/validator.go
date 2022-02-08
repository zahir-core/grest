package validator

import (
	"database/sql/driver"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	translations "github.com/go-playground/validator/v10/translations/en"

	"grest.dev/grest"
	"grest.dev/grest/db"
)

// use a single instance of Validate, it caches struct info
var (
	validate *validator.Validate
	i18n     = map[string]ut.Translator{}
)

func Configure() {
	validate = validator.New()
	validate.RegisterTagNameFunc(validateTagNamer)
	validate.RegisterCustomTypeFunc(validateValuer,
		db.NullBool{},
		db.NullInt64{},
		db.NullFloat64{},
		db.NullString{},
		db.NullDateTime{},
		db.NullDate{},
		db.NullTime{},
		db.NullText{},
		db.NullJSON{},
		db.NullUUID{},
	)
	AddTranslator("en", en.New(), translations.RegisterDefaultTranslations)
}

func validateTagNamer(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

func validateValuer(field reflect.Value) interface{} {
	if valuer, ok := field.Interface().(driver.Valuer); ok {
		val, err := valuer.Value()
		if err == nil {
			return val
		}
	}

	return nil
}

func AddValidator(tag string, fn validator.Func, callValidationEvenIfNull ...bool) error {
	return validate.RegisterValidation(tag, fn, callValidationEvenIfNull...)
}

func AddTranslator(lang string, lt locales.Translator, addDefaultTranslations func(v *validator.Validate, trans ut.Translator) error) error {
	trans, _ := ut.New(lt, lt).GetTranslator(lang)
	err := addDefaultTranslations(validate, trans)
	if err != nil {
		return err
	}
	if i18n == nil {
		i18n = map[string]ut.Translator{lang: trans}
	} else {
		i18n[lang] = trans
	}
	return nil
}

func IsValid(v interface{}, tag string) bool {
	err := validate.Var(v, tag)
	if err != nil {
		return false
	}
	return true
}

func Validate(lang string, v interface{}) error {
	trans, ok := i18n[lang]
	if !ok {
		trans, _ = i18n["en"]
	}
	err := validate.Struct(v)
	if err != nil {
		message := ""
		detail := map[string]interface{}{}
		errs := err.(validator.ValidationErrors)
		for _, e := range errs {
			msg := e.Translate(trans)
			if message == "" {
				message = msg
			}
			detail[e.Namespace()] = map[string]interface{}{e.Tag(): msg}
		}
		return grest.NewError(http.StatusBadRequest, message, detail)
	}
	return err
}
