package convert

import (
	"encoding/json"
	"encoding/xml"
	"reflect"
	"strings"

	"github.com/tidwall/gjson"
)

type jsonData struct {
	data interface{}
	err  error
}

// ToFlatJSON convert from structured json byte or struct or slice of struct to dot notation key flat map
// if "data" and "v" is not nil and "v" is valid struct or slice of struct, flat map is converted from "data" based on "v"
// if "data" is not nil and "v" is nil or not valid struct or slice of struct, flat map is converted from "data" based on "data"
// if "data" is nil and "v" is not nil and "v" is valid struct or slice of struct, flat map is converted from "v" based on "v"
func ToFlatJSON(data []byte, v interface{}) jsonData {
	if data != nil {
		return flatFromStructuredJSONByte(data, v)
	}
	return newFlatFromStruct(v)
}

// ToStructuredJSON convert from flat json byte or struct or slice of struct to structured map based on dot notation key
// if "data" and "v" is not nil and "v" is valid struct or slice of struct, structured map is converted from "data" based on "v"
// if "data" is not nil and "v" is nil or not valid struct or slice of struct, structured map is converted from "data" based on "data"
// if "data" is nil and "v" is not nil and "v" is valid struct or slice of struct, structured map is converted from "v" based on "v"
func ToStructuredJSON(data []byte, v interface{}) jsonData {
	return jsonData{}
}

func (d jsonData) Marshal() ([]byte, error) {
	if d.err != nil {
		return []byte{}, d.err
	}
	return json.Marshal(d.data)
}

func (d jsonData) MarshalXml() ([]byte, error) {
	if d.err != nil {
		return []byte{}, d.err
	}
	return xml.Marshal(d.data)
}

func (d jsonData) Unmarshal(v interface{}) error {
	if d.err != nil {
		return d.err
	}
	b, err := d.Marshal()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func (d jsonData) UnmarshalXml(v interface{}) error {
	if d.err != nil {
		return d.err
	}
	b, err := d.MarshalXml()
	if err != nil {
		return err
	}
	return xml.Unmarshal(b, v)
}

// flatFromStructuredJSONByte convert from structured json byte to dot notation key flat map
// if "data" and "v" is not nil and "v" is valid struct or slice of struct, flat map is converted from "data" based on "v"
// if "data" is not nil and "v" is nil or not valid struct or slice of struct, flat map is converted from "data" based on "data"
func flatFromStructuredJSONByte(data []byte, v interface{}) jsonData {
	if v != nil {
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Ptr {
			t = reflect.ValueOf(v).Elem().Type()
		}
		if t.Kind() == reflect.Struct {
			return jsonData{data: flatMapFromStructuredJSONObjectBasedOnStructType(data, t)}
		} else if t.Kind() == reflect.Slice {
			slc := []interface{}{}
			slcType := t.Elem()
			gjson.ParseBytes(data).ForEach(func(key gjson.Result, value gjson.Result) bool {
				if slcType.Kind() == reflect.Struct {
					slc = append(slc, flatMapFromStructuredJSONObjectBasedOnStructType([]byte(value.String()), slcType))
				} else {
					slc = append(slc, value.Value())
				}
				return true
			})
			return jsonData{data: slc}
		}
	}

	var tempData interface{}
	err := json.Unmarshal(data, &tempData)
	if err != nil {
		return jsonData{err: err}
	}

	m, isMap := tempData.(map[string]interface{})
	if isMap {
		return jsonData{data: flatMapFromStructuredMap(m)}
	}
	slc, isSlice := tempData.([]interface{})
	if isSlice {
		newSlice := []interface{}{}
		for i, s := range slc {
			m, isMap := s.(map[string]interface{})
			if isMap {
				newSlice[i] = jsonData{data: flatMapFromStructuredMap(m)}
			} else {
				newSlice[i] = s
			}
		}
		return jsonData{data: newSlice}
	}

	return jsonData{data: tempData}
}

func flatMapFromStructuredJSONObjectBasedOnStructType(data []byte, t reflect.Type) map[string]interface{} {
	flatMap := map[string]interface{}{}
	for i := 0; i < t.NumField(); i++ {
		hasTag := true
		key := t.Field(i).Tag.Get("json")
		if key == "" {
			hasTag = false
			key = ToSnakeCase(t.Field(i).Name)
		}

		jsonPath := strings.Split(key, ",")
		result := gjson.GetBytes(data, jsonPath[0])
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
					slc = append(slc, flatMapFromStructuredJSONObjectBasedOnStructType([]byte(value.String()), slcType))
				}
				return true
			})
			flatMap[flatKey] = slc
		default:
			flatMap[flatKey] = result.Value()
		}
	}
	return flatMap
}

func flatMapFromStructuredMap(m map[string]interface{}) map[string]interface{} {
	return m
}

func newFlatFromStruct(v interface{}) jsonData {
	return jsonData{}
}

func newStructuredFromFlatJSON(data []byte, v interface{}) jsonData {
	return jsonData{}
}

func newStructuredFromStruct(v interface{}) jsonData {
	return jsonData{}
}
