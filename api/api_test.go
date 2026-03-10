package api

import (
	"encoding/json"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	errBytes := []byte(`{"detail":"Not Found"}`)
	var apiErr APIError
	
	if err := json.Unmarshal(errBytes, &apiErr); err != nil {
		t.Fatalf("Failed to unmarshal APIError: %v", err)
	}
	
	if apiErr.Error() != "Not Found" {
		t.Errorf("Expected 'Not Found', got '%s'", apiErr.Error())
	}
}

func TestBoardRead_Unmarshal(t *testing.T) {
	data := []byte(`{
		"id": 1,
		"title": "Test Note",
		"content": "Content",
		"category": "work",
		"owner_id": 42
	}`)
	
	var board BoardRead
	if err := json.Unmarshal(data, &board); err != nil {
		t.Fatalf("Failed to unmarshal BoardRead: %v", err)
	}
	
	if board.ID != 1 || board.Title != "Test Note" || board.OwnerID != 42 {
		t.Errorf("Unmarshaled data mismatch: %+v", board)
	}
}
