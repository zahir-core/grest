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
	Data    interface{}
	IsMerge bool
}

func NewJSON(data interface{}, isKeepOriginalData ...bool) jsonData {
	var v interface{}
	var isMerge bool
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
	if len(isKeepOriginalData) > 0 {
		isMerge = isKeepOriginalData[0]
	}
	return jsonData{Data: v, IsMerge: isMerge}
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
