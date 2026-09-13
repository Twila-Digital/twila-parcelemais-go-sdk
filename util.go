package parcelemais

import "encoding/base64"
import "encoding/json"

func base64Encode(content []byte) string {
	return base64.StdEncoding.EncodeToString(content)
}

func unmarshalJSON(body []byte, target interface{}) error {
	if len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, target)
}
