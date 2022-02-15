package convert

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/tidwall/sjson"
)

func (j jsonData) ToStructured(separator ...Separator) jsonData {
	sep := Separator{Before: "."}
	if len(separator) > 0 {
		sep = separator[0]
	}
	mp, isMap := j.Data.(map[string]interface{})
	if isMap {
		return jsonData{Data: j.toStructuredMap(mp, sep)}
	}

	slc, isSlice := j.Data.([]interface{})
	if isSlice {
		var newSlice []interface{}
		for _, s := range slc {
			var newVal interface{}
			sMap, isSMap := s.(map[string]interface{})
			if isSMap {
				newVal = jsonData{Data: sMap}.ToStructured(separator...).Data
			} else if s != nil {
				newVal = s
			}
			newSlice = append(newSlice, newVal)
		}
		return jsonData{Data: newSlice}
	}

	return jsonData{Data: j.Data}
}

func (j jsonData) toStructuredMap(m map[string]interface{}, sep Separator) map[string]interface{} {
	jsonByte := []byte("{}")
	for k, v := range m {
		if sep.Before != "" {
			k = strings.ReplaceAll(k, sep.Before, ".")
		}
		if sep.After != "" {
			k = strings.ReplaceAll(k, sep.After, "")
		}
		slc, isSlice := v.([]interface{})
		if isSlice {
			if len(slc) > 0 {
				for i, s := range slc {
					iString := strconv.Itoa(i)
					sliceMap, isSliceMap := s.(map[string]interface{})
					if isSliceMap {
						jsonByte, _ = sjson.SetBytes(jsonByte, k+"."+iString, j.toStructuredMap(sliceMap, sep))
					} else if s != nil {
						jsonByte, _ = sjson.SetBytes(jsonByte, k+"."+iString, s)
					}
				}
			} else {
				jsonByte, _ = sjson.SetBytes(jsonByte, k, []interface{}{})

			}
		} else if v != nil {
			jsonByte, _ = sjson.SetBytes(jsonByte, k, v)
		}
	}
	res := map[string]interface{}{}
	json.Unmarshal(jsonByte, &res)
	return res
}
