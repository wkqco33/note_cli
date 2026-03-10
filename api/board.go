package api

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// GetBoards fetches all notes created by the current user
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

// CreateBoard creates a new note
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

// GetBoard fetches a single note by ID
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

// UpdateBoard modifies an existing note
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

// DeleteBoard deletes a note by ID
func (c *Client) DeleteBoard(id int) error {
	endpoint := fmt.Sprintf("/boards/%d", id)
	_, err := c.deleteReq(endpoint)
	return err
}
