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
    "images": []
  }
  ```

### 3.3 메모 상세 조회 / 수정 / 삭제

- **상세**: `GET /boards/{board_id}`
- **수정**: `PATCH /boards/{board_id}` (본인 것만 가능)
- **삭제**: `DELETE /boards/{board_id}` (본인 것만 가능)

---

## 4. 사용자 관리 (User)

### 4.1 회원가입

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

## 5. 추천 클라이언트 워크플로우

1. **초기 실행**: 로컬에 저장된 `access_token`이 있는지 확인.
2. **인증 확인**: `GET /boards/me`를 호출하여 토큰 유효성 검사.
3. **토큰 만료 처리 (401 Unauthorized)**:
   - 저장된 `refresh_token`이 있다면 `POST /auth/refresh` 호출.
   - 성공 시 토큰 갱신 후 원래 요청 재시도.
   - 실패 시 로그인 화면(`POST /auth/login`) 표시.
4. **데이터 동기화**: `GET /boards/me` 결과를 로컬 리스트로 출력.
5. **오프라인 모드 (권장)**: 작성한 메모를 로컬에 임시 저장 후 서버 연결 시 `POST` 요청 수행.

---

## 6. 주요 에러 코드

- `401 Unauthorized`: 토큰 만료 또는 잘못된 인증 정보.
- `403 Forbidden`: 본인이 작성하지 않은 메모에 대한 수정/삭제 시도.
- `404 Not Found`: 존재하지 않는 메모 ID 요청.
- `422 Unprocessable Entity`: 필수 필드 누락 또는 데이터 형식 오류.
