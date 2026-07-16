//go:build windows

package config

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows에서는 DPAPI로 암호화한다. 키는 OS가 현재 사용자 계정에 바인딩해
// 관리하므로 별도 키 파일이 없고, 다른 사용자/컴퓨터에서는 복호화할 수 없다.

func protectSecret(data []byte) ([]byte, error) {
	in := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	var out windows.DataBlob
	if err := windows.CryptProtectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))

	return append([]byte(nil), unsafe.Slice(out.Data, out.Size)...), nil
}

func unprotectSecret(data []byte) ([]byte, error) {
	in := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	var out windows.DataBlob
	if err := windows.CryptUnprotectData(&in, nil, nil, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))

	return append([]byte(nil), unsafe.Slice(out.Data, out.Size)...), nil
}
