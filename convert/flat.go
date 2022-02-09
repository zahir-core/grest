package convert

func (j jsonData) ToFlat(separator ...Separator) jsonData {
	sep := Separator{Before: "."}
	if len(separator) > 0 {
		sep = separator[0]
	}
	mp, isMap := j.Data.(map[string]interface{})
	if isMap {
		result := make(map[string]interface{})
		j.toFlatMap(result, mp, sep, true)
		return jsonData{Data: result}
	}

	slc, isSlice := j.Data.([]interface{})
	if isSlice {
		var newSlice []interface{}
		for _, s := range slc {
			var newVal interface{}
			sMap, isSMap := s.(map[string]interface{})
			if isSMap {
				newVal = jsonData{Data: sMap}.ToFlat(separator...).Data
			} else {
				newVal = s
			}
			newSlice = append(newSlice, newVal)
		}
		return jsonData{Data: newSlice}
	}

	return jsonData{Data: j.Data}
}

func (j jsonData) toFlatMap(flatMap map[string]interface{}, data interface{}, sep Separator, isTop bool, pref ...string) {
	prefix := ""
	if len(pref) > 0 {
		prefix = pref[0]
	}
	assign := func(newKey string, v interface{}) {
		switch v.(type) {
		case map[string]interface{}:
			j.toFlatMap(flatMap, v, sep, false, newKey)
		default:
			flatMap[newKey] = jsonData{Data: v}.ToFlat(sep).Data
		}
	}

	mp, isMap := data.(map[string]interface{})
	if isMap {
		for k, v := range mp {
			newKey := j.joinKey(prefix, k, sep, isTop)
			assign(newKey, v)
		}
	}
}

func (j jsonData) joinKey(prefix, key string, sep Separator, isTop bool) string {
	newKey := prefix

	if isTop {
		newKey += key
	} else {
		newKey += sep.Before + key + sep.After
	}

	return newKey
}
