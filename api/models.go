package api



// TokenResponse represents the JWT tokens returned from auth endpoints
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

// UserCreate represents the payload to register a new user
type UserCreate struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// BoardCreate represents the payload to create a new board/note
type BoardCreate struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Category string   `json:"category"`
	Images   []string `json:"images,omitempty"`
}

// BoardUpdate represents the payload to update an existing board/note
type BoardUpdate struct {
	Title    *string   `json:"title,omitempty"`
	Content  *string   `json:"content,omitempty"`
	Category *string   `json:"category,omitempty"`
	Images   *[]string `json:"images,omitempty"`
}

// BoardRead represents a board/note from the API responses
type BoardRead struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Category  string    `json:"category"`
	Images    []string  `json:"images"`
	OwnerID   int       `json:"owner_id"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

// APIError represents the error structure returned from the server
type APIError struct {
	Detail string `json:"detail"`
}

func (e *APIError) Error() string {
	return e.Detail
}

// FileRead represents a file from the API responses
type FileRead struct {
	ID               int    `json:"id"`
	Filename         string `json:"filename"`
	OriginalFilename string `json:"original_filename"`
	FileSize         int    `json:"file_size"`
	ContentType      string `json:"content_type"`
	URL              string `json:"url"`
	UserID           int    `json:"user_id"`
	CreatedAt        string `json:"created_at"`
}
