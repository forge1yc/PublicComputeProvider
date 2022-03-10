package utils

import "encoding/json"

func GetJsonBody(body []byte) map[string]interface{} {
	var  object interface{}
	err := json.Unmarshal(body, &object)
	if err != nil {
		return nil
	} else {
		return object.(map[string]interface{})
	}
}