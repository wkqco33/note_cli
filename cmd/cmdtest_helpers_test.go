package cmd

import (
	"archive/zip"
	"io"
	"os"
	"strings"
)

// createTestZip 테스트용 ZIP 아카이브를 생성.
func createTestZip(path string, contents map[string]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	w := zip.NewWriter(f)
	defer func() {
		_ = w.Close()
	}()

	for name, body := range contents {
		entry, err := w.Create(name)
		if err != nil {
			return err
		}
		if _, err := io.Copy(entry, strings.NewReader(body)); err != nil {
			return err
		}
	}
	return nil
}
