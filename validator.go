package grest

import (
	"database/sql/driver"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/locales"
	ut "github.com/go-playground/universal-translator"
	goValidator "github.com/go-playground/validator/v10"

	"grest.dev/grest/db"
)

type Validator struct {
	*goValidator.Validate
	I18n map[string]ut.Translator
}

func NewValidator() Validator {
	validator := Validator{Validate: goValidator.New()}
	validator.RegisterTagNameFunc(validator.ValidateTagNamer)
	validator.RegisterCustomTypeFunc(validator.ValidateValuer,
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
	return validator
}

func (Validator) ValidateTagNamer(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

func (Validator) ValidateValuer(field reflect.Value) interface{} {
	if valuer, ok := field.Interface().(driver.Valuer); ok {
		val, err := valuer.Value()
		if err == nil {
			return val
		}
	}
	return nil
}

func (validator *Validator) RegisterTranslator(lang string, lt locales.Translator, regFunc func(v *goValidator.Validate, trans ut.Translator) error) error {
	trans, _ := ut.New(lt, lt).GetTranslator(lang)
	err := regFunc(validator.Validate, trans)
	if err != nil {
		return err
	}
	if validator.I18n == nil {
		validator.I18n = map[string]ut.Translator{lang: trans}
	} else {
		validator.I18n[lang] = trans
	}
	return nil
}

func (validator *Validator) IsValid(v interface{}, tag string) bool {
	err := validator.Var(v, tag)
	if err != nil {
		return false
	}
	return true
}

func (validator *Validator) ValidateStruct(v interface{}, lang string) error {
	trans, ok := validator.I18n[lang]
	if !ok {
		trans, _ = validator.I18n["en"]
	}
	err := validator.Struct(v)
	if err != nil {
		message := ""
		detail := map[string]interface{}{}
		errs := err.(goValidator.ValidationErrors)
		for _, e := range errs {
			msg := e.Translate(trans)
			if message == "" {
				message = msg
			}
			detail[e.Field()] = map[string]interface{}{e.Tag(): msg}
		}
		return NewError(http.StatusBadRequest, message, detail)
	}
	return err
}
