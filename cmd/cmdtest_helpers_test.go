package cmd

import (
	"archive/zip"
	"encoding/json"
	"io"
	"os"
	"strings"

	"note_cli/api"
)

// createTestZip 테스트용 ZIP 아카이브를 생성.
func createTestZip(path string, contents map[string]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

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

// buildBackupZip notes.json/files.json + 파일 엔트리로 구성된 백업 ZIP 생성.
func buildBackupZip(path string, boards []api.BoardRead, files []api.FileRead, fileContents map[string]string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	notesData, err := json.MarshalIndent(boards, "", "  ")
	if err != nil {
		return err
	}
	if entry, err := w.Create("notes.json"); err != nil {
		return err
	} else if _, err := entry.Write(notesData); err != nil {
		return err
	}

	filesData, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return err
	}
	if entry, err := w.Create("files.json"); err != nil {
		return err
	} else if _, err := entry.Write(filesData); err != nil {
		return err
	}

	for name, body := range fileContents {
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
