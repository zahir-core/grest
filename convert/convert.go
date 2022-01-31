package convert

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

func ToCamelCase(str string, d ...string) string {
	delimiter := "_"
	if len(d) > 0 {
		delimiter = d[0]
	}
	link := regexp.MustCompile("(^[A-Za-z])|" + delimiter + "([A-Za-z])")
	return link.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.Replace(s, delimiter, "", -1))
	})
}

func ToSnakeCase(str string, d ...string) string {
	delimiter := "_"
	if len(d) > 0 {
		delimiter = d[0]
	}
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(str, "${1}"+delimiter+"${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}"+delimiter+"${2}")
	return strings.ToLower(snake)
}

func ToFlatJSON(m interface{}, b []byte) ([]byte, error) {
	t := reflect.TypeOf(m)
	if t.Kind() == reflect.Ptr {
		t = reflect.ValueOf(m).Elem().Type()
	}
	if t.Kind() == reflect.Struct {
		return flatJsonFromStruct(t, b)
	} else if t.Kind() == reflect.Slice {
		slcType := t.Elem()
		if slcType.Kind() == reflect.Struct {
			slc := []interface{}{}
			gjson.ParseBytes(b).ForEach(func(key gjson.Result, value gjson.Result) bool {
				var slcVal interface{}
				slcByte, err := flatJsonFromStruct(slcType, []byte(value.String()))
				if err == nil {
					err = json.Unmarshal(slcByte, &slcVal)
					if err == nil {
						slc = append(slc, slcVal)
					}
				}
				return true
			})
			return json.Marshal(slc)
		}
	}

	return b, nil
}

func flatJsonFromStruct(t reflect.Type, b []byte) ([]byte, error) {
	flatMap := map[string]interface{}{}
	for i := 0; i < t.NumField(); i++ {
		hasTag := true
		key := t.Field(i).Tag.Get("json")
		if key == "" {
			hasTag = false
			key = ToSnakeCase(t.Field(i).Name)
		}

		jsonPath := strings.Split(key, ",")
		result := gjson.GetBytes(b, jsonPath[0])
		flatKey := jsonPath[0]
		if !hasTag {
			flatKey = t.Field(i).Name
		}

		switch t.Field(i).Type.Kind() {
		case reflect.Bool:
			flatMap[flatKey] = result.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			flatMap[flatKey] = result.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			flatMap[flatKey] = result.Uint()
		case reflect.Float32, reflect.Float64:
			flatMap[flatKey] = result.Float()
		case reflect.Slice:
			slc := []interface{}{}
			slcType := t.Field(i).Type.Elem()
			result.ForEach(func(key gjson.Result, value gjson.Result) bool {
				if slcType.Kind() != reflect.Struct {
					slc = append(slc, value.Value())
				} else {
					var slcVal interface{}
					slcByte, err := flatJsonFromStruct(slcType, []byte(value.String()))
					if err == nil {
						err = json.Unmarshal(slcByte, &slcVal)
						if err == nil {
							slc = append(slc, slcVal)
						}
					}
				}
				return true
			})
			flatMap[flatKey] = slc
		default:
			flatMap[flatKey] = result.Value()
		}
	}
	return json.Marshal(flatMap)
}
