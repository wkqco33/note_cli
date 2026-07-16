package api

// TokenResponse 인증 엔드포인트에서 반환된 JWT 토큰 정보
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

// UserCreate 신규 사용자 등록 페이로드
type UserCreate struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// BoardCreate 새 노트 생성 페이로드
type BoardCreate struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Category string   `json:"category"`
	Images   []string `json:"images,omitempty"`
}

// BoardUpdate 기존 노트 수정 페이로드
type BoardUpdate struct {
	Title    *string   `json:"title,omitempty"`
	Content  *string   `json:"content,omitempty"`
	Category *string   `json:"category,omitempty"`
	Images   *[]string `json:"images,omitempty"`
}

// BoardRead API 응답을 통해 반환된 노트 정보
type BoardRead struct {
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	Category  string   `json:"category"`
	Images    []string `json:"images"`
	OwnerID   int      `json:"owner_id"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// APIError 서버에서 반환한 오류 구조
type APIError struct {
	Detail string `json:"detail"`
}

func (e *APIError) Error() string {
	return e.Detail
}

// apiErrorEnvelope {"error":{"code":...,"message":...}} 형태의 서버 오류 응답
type apiErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// FileRead API 응답을 통해 반환된 파일 정보
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
