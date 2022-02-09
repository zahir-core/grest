package convert

import (
	"encoding/json"
	"encoding/xml"
)

type Separator struct {
	Before string
	After  string
}

type jsonData struct {
	Data interface{}
}

func NewJSON(data interface{}) jsonData {
	var v interface{}
	bt, isByte := data.([]byte) // from json byte
	if !isByte {
		s, isString := data.(string) // from json string
		if isString {
			bt = []byte(s)
		} else {
			bt, _ = json.Marshal(data) // from struct or other
		}
	}
	if bt != nil {
		err := json.Unmarshal(bt, &v)
		if err != nil {
			err = xml.Unmarshal(bt, &v)
			if err != nil {
				return jsonData{Data: data}
			}
		}
	}
	return jsonData{Data: v}
}

func (j jsonData) Marshal() ([]byte, error) {
	return json.Marshal(j.Data)
}

func (j jsonData) MarshalIndent(indent string) ([]byte, error) {
	return json.MarshalIndent(j.Data, "", indent)
}

func (j jsonData) Unmarshal(v interface{}) error {
	data, err := j.Marshal()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
