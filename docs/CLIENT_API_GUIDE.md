# API 사용 가이드

`note_cli`가 사용하는 서버 API 명세입니다. 이 가이드는 서버 API만 다루며, 클라이언트가 기존 API를 조합해 구현한 기능은 마지막 섹션에서 설명합니다.

## 기본 설정

- **Base URL**: `http://<server-ip>:8880/api/v1`
- **Content-Type**: `application/json`
- **인증 방식**: `Authorization: Bearer <access_token>` 헤더

---

## 인증

### 로그인 및 토큰 발급

- **Endpoint**: `POST /auth/login`
- **Body (Form Data)**:
  - `username`: 사용자 이메일
  - `password`: 비밀번호
- **Response**: 새 액세스 토큰과 리프레시 토큰을 반환합니다.

  ```json
  {
    "access_token": "eyJhbG...",
    "refresh_token": "eyJhbG...",
    "token_type": "bearer"
  }
  ```

- **비고**: `access_token`은 30분, `refresh_token`은 7일간 유효합니다. 클라이언트는 두 토큰을 로컬에 안전하게 저장해야 합니다.

### 토큰 갱신

액세스 토큰이 만료되었을 때 비밀번호 입력 없이 세션을 연장합니다.

- **Endpoint**: `POST /auth/refresh`
- **Body (JSON)**:
  - `refresh_token`: 저장해 둔 리프레시 토큰 문자열
- **Response**: 새 `access_token`과 `refresh_token`을 반환합니다.

  ```json
  {
    "access_token": "eyJhbG...",
    "refresh_token": "eyJhbG...",
    "token_type": "bearer"
  }
  ```

- **비고**: 토큰이 URL에 노출되면 로그나 프록시에 유출될 수 있으므로 반드시 요청 본문으로 전송합니다.

---

## 노트 (Board API)

서버의 `Board` 기능을 노트 저장소로 활용합니다.

### 내 노트 목록 조회

현재 로그인한 사용자가 작성한 노트만 가져옵니다.

- **Endpoint**: `GET /boards/me`
- **Header**: `Authorization: Bearer <access_token>`
- **Response**: `List[BoardRead]` (최신순 정렬)

### 노트 작성

- **Endpoint**: `POST /boards`
- **Header**: `Authorization: Bearer <access_token>`
- **Body**:

  ```json
  {
    "title": "노트 제목",
    "content": "노트 내용",
    "category": "work",
    "images": ["http://<server-ip>:8880/static/uploads/filename.png"]
  }
  ```

- **비고**: `images` 필드에는 파일 업로드 API로 얻은 `url` 문자열 목록을 전달합니다.

### 노트 상세 조회·수정·삭제

- **상세**: `GET /boards/{board_id}`
- **수정**: `PATCH /boards/{board_id}` (본인 노트만 가능)
- **삭제**: `DELETE /boards/{board_id}` (본인 노트만 가능)

---

## 파일 (File API)

노트에 이미지나 파일을 첨부할 때 사용합니다.

### 파일 업로드

- **Endpoint**: `POST /files/upload`
- **Header**: `Authorization: Bearer <access_token>`
- **Body (Multipart/form-data)**:
  - `file`: 업로드할 파일 객체
- **Response**:

  ```json
  {
    "id": 1,
    "filename": "uuid_filename.png",
    "original_filename": "my_image.png",
    "file_size": 1024,
    "content_type": "image/png",
    "url": "http://<server-ip>:8880/static/uploads/uuid_filename.png",
    "user_id": 1,
    "created_at": "2024-03-11T10:00:00"
  }
  ```

- **활용**: 응답의 `url`을 노트 작성(`POST /boards`) 또는 수정(`PATCH /boards`) 시 `images` 목록에 포함해 전송합니다.

### 내 파일 목록 조회

- **Endpoint**: `GET /files`
- **Header**: `Authorization: Bearer <access_token>`
- **Response**: `List[FileRead]`

### 파일 다운로드

- **Endpoint**: `GET /files/download/{file_id}`
- **Header**: `Authorization: Bearer <access_token>`
- **Response**: 파일 바이너리 데이터 (원본 파일명으로 다운로드)

### 파일 삭제

- **Endpoint**: `DELETE /files/{file_id}`
- **Header**: `Authorization: Bearer <access_token>`
- **Response**: `204 No Content`

---

## 사용자 (User)

### 회원가입

CLI를 처음 사용하는 사용자를 등록합니다.

- **Endpoint**: `POST /users` (인증 불필요)
- **Body**:

  ```json
  {
    "name": "홍길동",
    "email": "user@example.com",
    "password": "securepassword123"
  }
  ```

---

## 클라이언트 워크플로우

1. **초기 실행**: 로컬에 저장된 `access_token`이 있는지 확인합니다.
2. **인증 확인**: `GET /boards/me`를 호출해 토큰 유효성을 검사합니다.
3. **파일 첨부**: `POST /files/upload`로 파일을 먼저 업로드해 `url`을 확보한 뒤, 그 `url`을 포함해 `POST /boards`를 호출합니다.
4. **토큰 만료(401 Unauthorized)**: `refresh_token`이 있으면 `POST /auth/refresh`를 호출하고, 성공 시 토큰을 갱신한 뒤 원래 요청을 재시도합니다. 실패하면 로그인(`POST /auth/login`)을 요청합니다.
5. **데이터 동기화**: `GET /boards/me` 결과를 로컬 목록으로 출력합니다.

---

## 주요 에러 코드

- `401 Unauthorized`: 토큰 만료 또는 잘못된 인증 정보
- `403 Forbidden`: 본인이 작성하지 않은 노트에 대한 수정·삭제 시도
- `404 Not Found`: 존재하지 않는 노트 ID 요청
- `422 Unprocessable Entity`: 필수 필드 누락 또는 데이터 형식 오류

---

## 클라이언트 사이드 기능

이 섹션의 기능은 서버에 전용 엔드포인트가 없고 클라이언트가 기존 API를 조합해 구현한 것입니다. 데이터가 많아지면 네트워크 비용이 선형으로 증가하므로 대량 데이터에서는 주의가 필요합니다.

- **검색(`search`)**: `GET /boards/me`(필요 시 `GET /files` 포함)로 전체를 내려받은 뒤 클라이언트 메모리에서 제목·내용·파일명 부분 일치로 필터링합니다.
- **백업 내보내기(`export`)**: `GET /boards/me`, `GET /files` 결과와 `GET /files/download/{id}` 스트림을 ZIP으로 묶어 로컬에 저장합니다.
- **복원(`import`)**: ZIP 내 `notes.json`/`files.json`을 파싱하고, 파일은 `POST /files/upload`로 재업로드한 뒤 새 URL로 노트를 `POST /boards`로 재생성합니다. `--clean`을 지정하면 기존 노트와 파일을 `DELETE`로 일괄 삭제한 뒤 복원합니다.
