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

// Validator struct wraps the validator.Validate and provides additional methods for validation.
type Validator struct {
	*validator.Validate
	I18n map[string]ut.Translator
}

// New initializes a new instance of the Validator.
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

// ValidateTagNamer returns the validation tag name for a struct field.
func (*Validator) ValidateTagNamer(fld reflect.StructField) string {
	name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
	if name == "-" {
		return ""
	}
	return name
}

// ValidateValuer attempts to retrieve the value for validation from a Valuer interface.
func (*Validator) ValidateValuer(field reflect.Value) any {
	if valuer, ok := field.Interface().(driver.Valuer); ok {
		val, err := valuer.Value()
		if err == nil {
			return val
		}
	}
	return nil
}

// RegisterTranslator registers a translator for a specific language and associates it with the Validator instance.
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

// IsValid checks if a value is valid based on a validation tag.
func (v *Validator) IsValid(val any, tag string) bool {
	err := v.Var(val, tag)
	if err != nil {
		return false
	}
	return true
}

// ValidateStruct validates a struct and returns an error with translated validation messages.
func (v *Validator) ValidateStruct(val any, lang string) error {
	err := v.Struct(val)
	if err != nil {
		return v.TranslateError(err, lang)
	}
	return err
}

// TranslateError translates validation errors and creates an error instance with translated details.
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
