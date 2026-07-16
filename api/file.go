package api

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
)

// UploadFile multipart/form-data를 사용하여 서버에 단일 파일 업로드
func (c *Client) UploadFile(filePath string) (*FileRead, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not get file stat: %w", err)
	}

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	contentType := writer.FormDataContentType()

	errChan := make(chan error, 1)

	go func() {
		defer pw.Close()
		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			errChan <- err
			return
		}

		bar := progressbar.DefaultBytes(stat.Size(), "업로드")

		if _, err := io.Copy(io.MultiWriter(part, bar), file); err != nil {
			errChan <- err
			return
		}

		if err := writer.Close(); err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	resp, err := c.postStream("/files/upload", contentType, pr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if writeErr := <-errChan; writeErr != nil {
		return nil, fmt.Errorf("error building multipart request: %w", writeErr)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fr, err := decodeJSON[FileRead](body, nil)
	if err != nil {
		return nil, err
	}
	return &fr, nil
}

// sanitizeFilename 경로 구분자와 상대 경로 요소를 제거해 순수 파일명만 남긴다.
// 안전한 파일명이 남지 않으면 빈 문자열을 반환한다.
func sanitizeFilename(filename string) string {
	base := filepath.Base(filename)
	if base == "." || base == ".." || base == string(filepath.Separator) || base == "/" {
		return ""
	}

	return base
}

// GetFiles 현재 사용자가 업로드한 모든 파일 조회
func (c *Client) GetFiles() ([]FileRead, error) {
	return decodeJSON[[]FileRead](c.get("/files"))
}

// DownloadFile ID로 파일을 다운로드하여 지정된 디렉토리에 저장
// 파일이 저장된 절대 경로 반환
func (c *Client) DownloadFile(fileID int, destDir string, filename string) (string, error) {
	endpoint := fmt.Sprintf("/files/download/%d", fileID)

	resp, err := c.getStream(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if filename == "" {
		contentDisp := resp.Header.Get("Content-Disposition")
		if contentDisp != "" {
			_, params, err := mime.ParseMediaType(contentDisp)
			if err == nil && params["filename"] != "" {
				filename = params["filename"]
			}
		}
	}

	// 서버가 제공한 파일명에 경로 구분자가 섞여 있어도 대상 디렉토리를
	// 벗어나지 못하도록 파일명 부분만 사용
	filename = sanitizeFilename(filename)
	if filename == "" {
		filename = fmt.Sprintf("downloaded_file_%d", fileID)
	}
	destPath := filepath.Join(destDir, filename)

	outFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	var bar *progressbar.ProgressBar
	if resp.ContentLength > 0 {
		bar = progressbar.DefaultBytes(resp.ContentLength, "다운로드")
	} else {
		bar = progressbar.DefaultBytes(-1, "다운로드")
	}

	_, err = io.Copy(io.MultiWriter(outFile, bar), resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	fmt.Println()

	return destPath, nil
}

// DownloadFileTemp ID로 파일을 임시 디렉토리에 조용히 다운로드
// 반환된 임시 파일 경로는 사용 후 호출자가 직접 삭제해야 함
func (c *Client) DownloadFileTemp(fileID int, ext string) (string, error) {
	endpoint := fmt.Sprintf("/files/download/%d", fileID)

	resp, err := c.getStream(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", fmt.Sprintf("note_img_*%s", ext))
	if err != nil {
		return "", fmt.Errorf("임시 파일 생성 실패: %w", err)
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("파일 다운로드 실패: %w", err)
	}

	return tmpFile.Name(), nil
}

// DeleteFile ID로 파일 삭제
func (c *Client) DeleteFile(id int) error {
	endpoint := fmt.Sprintf("/files/%d", id)
	_, err := c.deleteReq(endpoint)
	return err
}

// GetFileStream ID로 파일 데이터를 가져오는 Reader와 파일 크기를 반환
func (c *Client) GetFileStream(fileID int) (io.ReadCloser, int64, error) {
	endpoint := fmt.Sprintf("/files/download/%d", fileID)
	resp, err := c.getStream(endpoint)
	if err != nil {
		return nil, 0, err
	}
	return resp.Body, resp.ContentLength, nil
}
