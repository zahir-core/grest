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

func NewFlatJSON(data []byte, v interface{}) jsonData {
	return jsonData{}
}

func NewStructuredJSON(data []byte, v interface{}) jsonData {
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

func ToFlatJSON(v interface{}, data []byte) ([]byte, error) {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = reflect.ValueOf(v).Elem().Type()
	}
	if t.Kind() == reflect.Struct {
		return flatJsonFromStruct(t, data)
	} else if t.Kind() == reflect.Slice {
		slcType := t.Elem()
		if slcType.Kind() == reflect.Struct {
			slc := []interface{}{}
			gjson.ParseBytes(data).ForEach(func(key gjson.Result, value gjson.Result) bool {
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

	return data, nil
}

func flatJsonFromStruct(t reflect.Type, data []byte) ([]byte, error) {
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
