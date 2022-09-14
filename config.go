package grest

import (
	"os"
	"reflect"
	"strconv"
	"time"
)

func LoadEnv(key string, value any) {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Ptr {
		return
	}

	envValue := os.Getenv(key)
	if envValue == "" {
		return
	}

	switch val.Elem().Kind() {
	case reflect.String:
		val.Elem().SetString(envValue)
		return
	case reflect.Bool:
		v, err := strconv.ParseBool(envValue)
		if err == nil {
			val.Elem().SetBool(v)
		}
		return
	case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(envValue, 10, 64)
		if err == nil {
			val.Elem().SetUint(v)
		}
		return
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(envValue, 10, 64)
		if err == nil {
			val.Elem().SetInt(v)
		}
		return
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(envValue, 64)
		if err == nil {
			val.Elem().SetFloat(v)
		}
		return
	case reflect.Complex64, reflect.Complex128:
		v, err := strconv.ParseComplex(envValue, 64)
		if err == nil {
			val.Elem().SetComplex(v)
		}
		return
	}

	_, isBytes := val.Elem().Interface().([]byte)
	if isBytes {
		val.Elem().SetBytes([]byte(envValue))
		return
	}

	_, isTimeDuration := val.Elem().Interface().(time.Duration)
	if isTimeDuration {
		timeDuration, err := time.ParseDuration(envValue)
		if err == nil {
			val.Elem().Set(reflect.ValueOf(timeDuration))
		}
		return
	}
}
