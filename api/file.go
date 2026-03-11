package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// UploadFile uploads a single file to the server using multipart/form-data
func (c *Client) UploadFile(filePath string) (*FileRead, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("could not create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("could not copy file content: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("could not close multipart writer: %w", err)
	}

	respBody, err := c.post("/files/upload", writer.FormDataContentType(), body)
	if err != nil {
		return nil, err
	}

	var fr FileRead
	if err := json.Unmarshal(respBody, &fr); err != nil {
		return nil, err
	}
	return &fr, nil
}

// GetFiles fetches all files uploaded by the current user
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

// DownloadFile downloads a file by its ID and saves it to the specified directory.
// Returns the absolute path where the file was saved.
func (c *Client) DownloadFile(fileID int, destDir string, filename string) (string, error) {
	endpoint := fmt.Sprintf("/files/download/%d", fileID)
	
	// API 클라이언트 구조상 c.get은 메모리에 모두 올리는 구조일 수 있음.
	// c.httpClient를 직접 노출시키거나, c.downloadReq 등의 스트림 전용 메서드가 필요함.
	// client.go 의 구조를 살펴보고 필요한 경우 보완합니다.
	
	// 임시로 c.get을 통해 바이트를 받아오는 구조로 구현 (작은 파일에 한정)
	// 추후 client.go에 Download 전용 메서드가 추가되면 교체 가능.
	body, err := c.get(endpoint)
	if err != nil {
		return "", err
	}
	
	if filename == "" {
		filename = fmt.Sprintf("downloaded_file_%d", fileID)
	}
	destPath := filepath.Join(destDir, filename)
	
	err = os.WriteFile(destPath, body, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return destPath, nil
}

// DeleteFile deletes a file by ID
func (c *Client) DeleteFile(id int) error {
	endpoint := fmt.Sprintf("/files/%d", id)
	_, err := c.deleteReq(endpoint)
	return err
}
