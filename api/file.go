package api

import (
	"encoding/json"
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

	var fr FileRead
	if err := json.Unmarshal(body, &fr); err != nil {
		return nil, err
	}
	return &fr, nil
}

// GetFiles 현재 사용자가 업로드한 모든 파일 조회
func (c *Client) GetFiles() ([]FileRead, error) {
	body, err := c.get("/files")
	if err != nil {
		return nil, err
	}

	var files []FileRead
	if err := json.Unmarshal(body, &files); err != nil {
		return nil, err
	}
	return files, nil
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

		if filename == "" {
			filename = fmt.Sprintf("downloaded_file_%d", fileID)
		}
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
