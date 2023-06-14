package grest

import (
	"database/sql/driver"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/locales"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

type ValidatorInterface interface {
	New()
	ValidateTagNamer(fld reflect.StructField) string
	ValidateValuer(field reflect.Value) any
	RegisterTranslator(lang string, lt locales.Translator, regFunc func(v *validator.Validate, trans ut.Translator) error) error
	IsValid(val any, tag string) bool
	ValidateStruct(val any, lang string) error
	TranslateError(err error, lang string) error
}

type Validator struct {
	*validator.Validate
	I18n map[string]ut.Translator
}

func (v *Validator) New() {
	v.Validate = validator.New()
	v.RegisterTagNameFunc(v.ValidateTagNamer)
	v.RegisterCustomTypeFunc(v.ValidateValuer,
		NullBool{},
		NullInt64{},
		NullFloat64{},
		NullString{},
		NullDateTime{},
		NullDate{},
		NullTime{},
		NullText{},
		NullJSON{},
		NullUUID{},
	)
}

func (*Validator) ValidateTagNamer(fld reflect.StructField) string {
	name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}

func (*Validator) ValidateValuer(field reflect.Value) any {
	if valuer, ok := field.Interface().(driver.Valuer); ok {
		val, err := valuer.Value()
		if err == nil {
			return val
		}
	}
	return nil
}

func (v *Validator) RegisterTranslator(lang string, lt locales.Translator, regFunc func(v *validator.Validate, trans ut.Translator) error) error {
	trans, _ := ut.New(lt, lt).GetTranslator(lang)
	err := regFunc(v.Validate, trans)
	if err != nil {
		return err
	}
	if v.I18n == nil {
		v.I18n = map[string]ut.Translator{lang: trans}
	} else {
		v.I18n[lang] = trans
	}
	return nil
}

func (v *Validator) IsValid(val any, tag string) bool {
	err := v.Var(val, tag)
	if err != nil {
		return false
	}
	return true
}

func (v *Validator) ValidateStruct(val any, lang string) error {
	err := v.Struct(val)
	if err != nil {
		return v.TranslateError(err, lang)
	}
	return err
}

func (v *Validator) TranslateError(err error, lang string) error {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}
	trans, ok := v.I18n[lang]
	if !ok {
		trans, _ = v.I18n["en"]
	}
	message := ""
	detail := map[string]any{}
	for _, e := range errs {
		msg := e.Translate(trans)
		if message == "" {
			message = msg
		}
		detail[e.Field()] = map[string]any{e.Tag(): msg}
	}
	return NewError(http.StatusBadRequest, message, detail)
}
