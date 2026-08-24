package api

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

// UploadFile multipart/form-data로 서버에 단일 파일을 업로드한다.
// io.Pipe 스트리밍 본문은 401 재시도가 불가능하므로 업로드 전에 토큰을 사전 검증하고,
// 실패 시에도 파이프 리더를 닫아 고루틴이 블록되지 않게 보장한다.
func (c *Client) UploadFile(filePath string) (*FileRead, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("파일을 열지 못했습니다: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("파일 정보를 가져오지 못했습니다: %w", err)
	}

	pr, pw := io.Pipe()
	// 업로드 실패 또는 함수 종료 시 파이프 리더를 닫아, 고루틴이
	// pw.Write()에서 무한 블록되는 것을 방지합니다.
	defer pr.Close()

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
		return nil, fmt.Errorf("multipart 요청 생성 오류: %w", writeErr)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 본문을 읽지 못했습니다: %w", err)
	}

	fr, err := decodeJSON[FileRead](body, nil)
	if err != nil {
		return nil, err
	}
	return &fr, nil
}

// sanitizeFilename 경로 구분자와 상대 경로 요소를 제거해 순수 파일명만 남긴다.
// Windows/Unix 경로 구분자를 모두 처리해 path traversal 공격을 방지하고,
// 안전한 파일명이 남지 않으면 빈 문자열을 반환한다.
func sanitizeFilename(filename string) string {
	normalized := strings.ReplaceAll(filename, "\\", "/")
	base := filepath.Base(normalized)
	if base == "." || base == ".." || base == "/" || base == "" {
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
		return "", fmt.Errorf("파일을 생성하지 못했습니다: %w", err)
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
		return "", fmt.Errorf("파일 쓰기 실패: %w", err)
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
