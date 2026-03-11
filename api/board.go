package api

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// GetBoards 현재 사용자가 작성한 모든 노트 목록 조회
func (c *Client) GetBoards() ([]BoardRead, error) {
	body, err := c.get("/boards/me")
	if err != nil {
		return nil, err
	}

	var boards []BoardRead
	if err := json.Unmarshal(body, &boards); err != nil {
		return nil, err
	}
	return boards, nil
}

// CreateBoard 새 노트 작성
func (c *Client) CreateBoard(board BoardCreate) (*BoardRead, error) {
	payload, err := json.Marshal(board)
	if err != nil {
		return nil, err
	}

	body, err := c.post("/boards", "application/json", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	var br BoardRead
	if err := json.Unmarshal(body, &br); err != nil {
		return nil, err
	}
	return &br, nil
}

// GetBoard 지정된 ID의 단일 노트 상세 조회
func (c *Client) GetBoard(id int) (*BoardRead, error) {
	endpoint := fmt.Sprintf("/boards/%d", id)
	body, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var br BoardRead
	if err := json.Unmarshal(body, &br); err != nil {
		return nil, err
	}
	return &br, nil
}

// UpdateBoard 기존 노트 수정
func (c *Client) UpdateBoard(id int, update BoardUpdate) (*BoardRead, error) {
	payload, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/boards/%d", id)
	body, err := c.patchReq(endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	var br BoardRead
	if err := json.Unmarshal(body, &br); err != nil {
		return nil, err
	}
	return &br, nil
}

// DeleteBoard ID로 노트 삭제
func (c *Client) DeleteBoard(id int) error {
	endpoint := fmt.Sprintf("/boards/%d", id)
	_, err := c.deleteReq(endpoint)
	return err
}
