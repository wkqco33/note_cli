package config

import (
	"encoding/base64"
	"strings"
)

// encPrefix 암호화되어 저장된 값임을 표시하는 접두사
const encPrefix = "enc:"

// encodeSecret 평문 비밀값을 플랫폼별 방식으로 암호화해 저장용 문자열로 변환
func encodeSecret(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}

	ciphertext, err := protectSecret([]byte(plain))
	if err != nil {
		return "", err
	}

	return encPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decodeSecret 저장된 비밀값을 평문으로 복원한다. 접두사가 없으면 그대로 반환.
func decodeSecret(stored string) (string, error) {
	if stored == "" || !strings.HasPrefix(stored, encPrefix) {
		return stored, nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(stored, encPrefix))
	if err != nil {
		return "", err
	}

	plain, err := unprotectSecret(ciphertext)
	if err != nil {
		return "", err
	}

	return string(plain), nil
}
