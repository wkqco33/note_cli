//go:build !windows

package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
)

// 비 Windows 환경에서는 설정 디렉토리의 로컬 키 파일(secret.key, 0600)로
// AES-256-GCM 암호화한다. 키가 같은 디스크에 있으므로 동일 사용자 권한의
// 접근까지 막지는 못하지만 설정 파일 열람만으로는 평문이 노출되지 않는다.

func protectSecret(data []byte) ([]byte, error) {
	gcm, err := loadAEAD()
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func unprotectSecret(data []byte) ([]byte, error) {
	gcm, err := loadAEAD()
	if err != nil {
		return nil, err
	}

	if len(data) < gcm.NonceSize() {
		return nil, errors.New("암호화된 데이터가 손상되었습니다")
	}

	return gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], nil)
}

func loadAEAD() (cipher.AEAD, error) {
	key, err := loadOrCreateKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}

func loadOrCreateKey() ([]byte, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}
	keyPath := filepath.Join(filepath.Dir(configPath), "secret.key")

	key, err := os.ReadFile(keyPath)
	if err == nil && len(key) == 32 {
		return key, nil
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	key = make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return nil, err
	}

	return key, nil
}
