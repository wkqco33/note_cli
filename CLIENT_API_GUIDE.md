# Client Note CLI App - API 사용 가이드

이 문서는 Client Note CLI 애플리케이션 개발을 위한 서버 API 명세 및 사용 가이드입니다.

## 1. 기본 설정

- **Base URL**: `http://<server-ip>:8880/api/v1`
- **Content-Type**: `application/json`
- **인증 방식**: `Authorization: Bearer <access_token>` (헤더에 포함)

---

## 2. 인증 및 로그인 (Auth)

사용자가 CLI 앱을 처음 실행하거나 세션이 만료되었을 때 사용합니다.

### 2.1 로그인 및 토큰 발급

- **Endpoint**: `POST /auth/login`
- **Body (Form Data)**:
  - `username`: 사용자 이메일
  - `password`: 비밀번호
- **Response**:

  ```json
  {
    "access_token": "eyJhbG...",
    "refresh_token": "eyJhbG...",
    "token_type": "bearer"
  }
  ```

- **비고**: `access_token`은 30분, `refresh_token`은 7일간 유효합니다. CLI 앱은 두 토큰을 로컬에 안전하게 저장해야 합니다.

### 2.2 토큰 갱신 (로그인 유지)

액세스 토큰이 만료되었을 때 비밀번호 입력 없이 세션을 연장합니다.

- **Endpoint**: `POST /auth/refresh`
- **Query Parameter**:
  - `refresh_token`: 저장해둔 리프레시 토큰 문자열
- **Response**: 새 `access_token`과 `refresh_token` 반환.

---

## 3. 메모 관리 (Board API 활용)

서버의 `Board` 기능을 메모 저장소로 활용합니다.

### 3.1 내 메모 목록 조회 (CLI 메인 화면)

현재 로그인한 사용자가 작성한 메모만 가져옵니다.

- **Endpoint**: `GET /boards/me`
- **Header**: `Authorization: Bearer <access_token>`
- **Response**: `List[BoardRead]` (최신순 정렬)

### 3.2 새 메모 작성

- **Endpoint**: `POST /boards`
- **Header**: `Authorization: Bearer <access_token>`
- **Body**:

  ```json
  {
    "title": "메모 제목",
    "content": "메모 내용",
    "category": "work",
    "images": ["http://<server-ip>:8880/static/uploads/filename.png"]
  }
  ```

- **비고**: `images` 필드에는 '4. 파일 업로드' API를 통해 얻은 파일의 `url` 문자열 리스트를 전달합니다.

### 3.3 메모 상세 조회 / 수정 / 삭제

- **상세**: `GET /boards/{board_id}`
- **수정**: `PATCH /boards/{board_id}` (본인 것만 가능)
- **삭제**: `DELETE /boards/{board_id}` (본인 것만 가능)

---

## 4. 파일 업로드 (File API)

메모에 이미지나 파일을 첨부하기 위해 사용합니다.

### 4.1 파일 업로드

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

- **활용**: 응답으로 받은 `url`을 메모 작성(`POST /boards`) 또는 수정(`PATCH /boards`) 시 `images` 리스트에 포함하여 전송합니다.

### 4.2 내 파일 목록 조회

- **Endpoint**: `GET /files`
- **Header**: `Authorization: Bearer <access_token>`
- **Response**: `List[FileRead]`

### 4.3 파일 다운로드

- **Endpoint**: `GET /files/download/{file_id}`
- **Header**: `Authorization: Bearer <access_token>`
- **Response**: 파일 바이너리 데이터 (원본 파일명으로 다운로드됨)

### 4.4 파일 삭제

- **Endpoint**: `DELETE /files/{file_id}`
- **Header**: `Authorization: Bearer <access_token>`
- **Response**: `204 No Content`

---

## 5. 사용자 관리 (User)

### 5.1 회원가입

CLI 앱을 처음 사용하는 사용자를 등록합니다.

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

## 6. 추천 클라이언트 워크플로우

1. **초기 실행**: 로컬에 저장된 `access_token`이 있는지 확인.
2. **인증 확인**: `GET /boards/me`를 호출하여 토큰 유효성 검사.
3. **파일 첨부 시**:
   - `POST /files/upload`로 파일을 먼저 업로드하고 `url`을 확보합니다.
   - 확보된 `url`을 포함하여 `POST /boards`를 호출합니다.
4. **토큰 만료 처리 (401 Unauthorized)**:
   - 저장된 `refresh_token`이 있다면 `POST /auth/refresh` 호출.
   - 성공 시 토큰 갱신 후 원래 요청 재시도.
   - 실패 시 로그인 화면(`POST /auth/login`) 표시.
5. **데이터 동기화**: `GET /boards/me` 결과를 로컬 리스트로 출력.
6. **오프라인 모드 (권장)**: 작성한 메모를 로컬에 임시 저장 후 서버 연결 시 `POST` 요청 수행.

---

## 7. 주요 에러 코드

- `401 Unauthorized`: 토큰 만료 또는 잘못된 인증 정보.
- `403 Forbidden`: 본인이 작성하지 않은 메모에 대한 수정/삭제 시도.
- `404 Not Found`: 존재하지 않는 메모 ID 요청.
- `422 Unprocessable Entity`: 필수 필드 누락 또는 데이터 형식 오류.

---

## 8. 클라이언트 사이드 기능 참고

본 가이드는 서버 API 명세만 다루며, 아래 기능은 서버에 전용 엔드포인트가 없고
클라이언트가 기존 API를 조합해 구현한 것이다. 따라서 데이터가 많아지면
네트워크 비용이 선형으로 증가하므로 대량 데이터에서는 주의가 필요하다.

- **검색(`search`)**: `GET /boards/me`(필요 시 `GET /files` 포함)로 전체를
  내려받은 뒤 클라이언트 메모리에서 제목/내용/파일명 부분 일치 필터링.
- **백업 내보내기(`export`)**: `GET /boards/me`, `GET /files` 결과와
  `GET /files/download/{id}` 스트림을 ZIP으로 묶어 로컬에 저장.
- **복원(`import`)**: ZIP 내 `notes.json`/`files.json`을 파싱 후 파일은
  `POST /files/upload`로 재업로드하고, 얻은 새 URL로 노트를 `POST /boards`
  로 재생성. `--clean` 시 기존 노트/파일을 `DELETE`로 일괄 삭제 후 복원.
