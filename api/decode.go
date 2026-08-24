package api

import "encoding/json"

// decodeJSON (body, err) 결과를 지정된 타입으로 역직렬화한다.
func decodeJSON[T any](body []byte, err error) (T, error) {
	var out T
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return out, err
	}

	return out, nil
}
