package convert

import "errors"

func NestedMapLookup(m map[string]interface{}, keys ...string) (interface{}, error) {
	if len(keys) == 0 {
		return nil, errors.New("keys is empty")
	}

	val, keyExist := m[keys[0]]
	if !keyExist {
		return nil, errors.New("keys is not exist")
	}

	mp, isMap := val.(map[string]interface{})
	if isMap {
		return NestedMapLookup(mp, keys[1:]...)
	}

	if len(keys) > 1 {
		return val, errors.New("keys is not exist")
	}

	return val, nil
}
