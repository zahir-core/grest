package grest

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"strings"
)

type JSONSeparator struct {
	Before string
	After  string
}

type JSON struct {
	Data       any
	IsMerge    bool
	IsRootOnly bool
}

func NewJSON(data any, isKeepOriginalData ...bool) JSON {
	var v any
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
				return JSON{Data: data}
			}
		}
	}
	if len(isKeepOriginalData) > 0 {
		isMerge = isKeepOriginalData[0]
	}
	return JSON{Data: v, IsMerge: isMerge}
}

func (j JSON) Marshal() ([]byte, error) {
	b, err := json.Marshal(j.Data)
	if err != nil {
		return b, NewError(http.StatusInternalServerError, err.Error())
	}
	return b, nil
}

func (j JSON) MarshalIndent(indent string) ([]byte, error) {
	b, err := json.MarshalIndent(j.Data, "", indent)
	if err != nil {
		return b, NewError(http.StatusInternalServerError, err.Error())
	}
	return b, nil
}

func (j JSON) Unmarshal(v any) error {
	data, err := j.Marshal()
	if err != nil {
		return NewError(http.StatusInternalServerError, err.Error())
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return NewError(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func (j JSON) ToFlat(separator ...JSONSeparator) JSON {
	sep := JSONSeparator{Before: "."}
	if len(separator) > 0 {
		sep = separator[0]
	}
	mp, isMap := j.Data.(map[string]any)
	if isMap {
		result := make(map[string]any)
		if j.IsMerge {
			result = mp
		}
		j.ToFlatMap(result, mp, sep, true)
		return JSON{Data: result}
	}

	slc, isSlice := j.Data.([]any)
	if isSlice {
		var newSlice []any
		for _, s := range slc {
			var newVal any
			sMap, isSMap := s.(map[string]any)
			if isSMap {
				newVal = JSON{Data: sMap}.ToFlat(separator...).Data
			} else {
				newVal = s
			}
			newSlice = append(newSlice, newVal)
		}
		return JSON{Data: newSlice}
	}

	return JSON{Data: j.Data}
}

func (j JSON) ToFlatMap(flatMap map[string]any, data any, sep JSONSeparator, isTop bool, pref ...string) {
	prefix := ""
	if len(pref) > 0 {
		prefix = pref[0]
	}
	assign := func(newKey string, v any) {
		switch v.(type) {
		case map[string]any:
			j.ToFlatMap(flatMap, v, sep, false, newKey)
		default:
			flatMap[newKey] = JSON{Data: v}.ToFlat(sep).Data
		}
	}

	mp, isMap := data.(map[string]any)
	if isMap {
		for k, v := range mp {
			newKey := j.JoinKey(prefix, k, sep, isTop)
			assign(newKey, v)
		}
	}
}

func (j JSON) JoinKey(prefix, key string, sep JSONSeparator, isTop bool) string {
	newKey := prefix

	if isTop {
		newKey += key
	} else {
		newKey += sep.Before + key + sep.After
	}

	return newKey
}

func (j JSON) ToStructured(separator ...JSONSeparator) JSON {
	sep := JSONSeparator{Before: "."}
	if len(separator) > 0 {
		sep = separator[0]
	}
	mp, isMap := j.Data.(map[string]any)
	if isMap {
		return JSON{Data: j.ToStructuredMap(mp, sep)}
	}

	slc, isSlice := j.Data.([]any)
	if isSlice {
		var newSlice []any
		for _, s := range slc {
			var newVal any
			sMap, isSMap := s.(map[string]any)
			if isSMap {
				newVal = JSON{Data: sMap}.ToStructured(separator...).Data
			} else if s != nil {
				newVal = s
			}
			newSlice = append(newSlice, newVal)
		}
		return JSON{Data: newSlice}
	}

	return JSON{Data: j.Data}
}

func (j JSON) ToStructuredMap(m map[string]any, sep JSONSeparator) map[string]any {
	nested := map[string]any{}
	for k, v := range m {
		if !j.IsRootOnly {
			vSlice, isVSlice := v.([]any)
			if isVSlice {
				vTemp := []map[string]any{}
				for _, d := range vSlice {
					dMap, isDataMap := d.(map[string]any)
					if isDataMap {
						vTemp = append(vTemp, j.ToStructuredMap(dMap, sep))
					}
				}
				if len(vTemp) > 0 {
					v = vTemp
				}
			}
		}
		keys := strings.Split(k, sep.Before)
		for i := len(keys) - 1; i >= 1; i-- {
			v = map[string]any{keys[i]: v}
		}
		nested[keys[0]] = j.FillMap(nested, keys[0], v)
	}
	return nested
}

func (j JSON) FillMap(data map[string]any, key string, val any) any {
	d, exist := data[key].(map[string]any)
	if exist {
		temp, _ := val.(map[string]any)
		for k, v := range temp {
			d[k] = j.FillMap(d, k, v)
		}
		return d
	}
	return val
}
