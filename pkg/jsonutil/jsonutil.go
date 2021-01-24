package jsonutil

import (
	"encoding/json"
)

func ToJson(v interface{}) string {
	if v == nil {
		return "{}"
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
