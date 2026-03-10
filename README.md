# 📝 Note CLI

터미널에서 빠르게 노트를 작성하고 관리할 수 있는 CLI 도구입니다.  
원격 API 서버에 노트를 저장하며, JWT 기반 인증을 통해 사용자별 노트를 안전하게 관리합니다.

---

## 주요 기능

- **노트 CRUD**: 노트 생성, 조회, 수정, 삭제
- **카테고리 분류**: Work / Personal / Idea / Other
- **외부 에디터 지원**: `$EDITOR` 환경변수에 설정된 편집기로 노트 내용 작성 (기본값: `vim`)
- **인터랙티브 UI**: [Charmbracelet](https://charm.sh) 라이브러리 기반의 TUI 폼
- **JWT 인증**: 액세스 토큰 만료 시 자동 재발급
- **로컬 설정 저장**: `~/.note_cli.json`에 토큰 정보 저장

---

## 요구 사항

- [Go](https://golang.org/) 1.21 이상
- `vim`, `nano` 등의 외부 편집기 (또는 `$EDITOR` 환경변수 설정)
- 노트 데이터를 저장할 API 서버 (`http://127.0.0.1:8880/api/v1`)

> API 서버 구성에 대한 자세한 내용은 [CLIENT_API_GUIDE.md](./CLIENT_API_GUIDE.md)를 참고하세요.

---

## 설치

```bash
# 저장소 클론
git clone https://github.com/wkqco33/note_cli.git
cd note_cli

# 빌드 (./note 바이너리 생성)
make build

# 또는 Go bin 경로에 설치 (~/.go/bin 또는 $GOPATH/bin)
make install
```

---

## 빠른 시작

```bash
# 1. 계정 등록
note register

# 2. 로그인
note login

# 3. 노트 추가
note add

# 4. 노트 목록 확인
note list
```

---

## 사용법

### 계정 관리

#### 회원가입
```bash
note register
```
이름, 이메일, 비밀번호를 입력하는 인터랙티브 폼이 표시됩니다.

#### 로그인
```bash
note login
```
이메일과 비밀번호를 입력하면 인증 토큰이 `~/.note_cli.json`에 저장됩니다.

---

### 노트 관리

#### 노트 목록 조회
```bash
note list
```
사용자의 모든 노트를 테이블 형식으로 출력합니다.

```
ID   TITLE          CATEGORY   UPDATED
1    회의 메모        Work       2026-03-10
2    아이디어 정리     Idea       2026-03-09
```

#### 노트 추가
```bash
note add
```
제목과 카테고리를 입력하는 폼이 나타난 후, 외부 편집기(기본: `vim`)가 열려 내용을 작성합니다.

카테고리 선택지:
- `Work`
- `Personal`
- `Idea`
- `Other`

#### 노트 조회
```bash
note view [ID]
```

예시:
```bash
note view 1
```

#### 노트 수정
```bash
note edit [ID]
```
기존 노트의 제목, 카테고리, 내용을 수정할 수 있습니다.

예시:
```bash
note edit 1
```

#### 노트 삭제
```bash
note delete [ID]
```

예시:
```bash
note delete 1
```

---

## 편집기 설정

노트 내용 작성 시 시스템의 기본 편집기를 사용합니다.  
원하는 편집기를 `$EDITOR` 환경변수로 지정할 수 있습니다.

```bash
# ~/.bashrc 또는 ~/.zshrc에 추가
export EDITOR=nano
```

---

## 설정 파일

로그인 후 `~/.note_cli.json` 파일이 자동으로 생성되며, 인증 토큰이 저장됩니다.

```json
{
  "access_token": "...",
  "refresh_token": "..."
}
```

- **액세스 토큰**: 유효 기간 30분, 만료 시 자동 재발급
- **리프레시 토큰**: 유효 기간 7일

---

## 개발

### 빌드 및 관련 명령어

| 명령어 | 설명 |
|---|---|
| `make build` | `./note` 바이너리 빌드 |
| `make install` | Go bin 경로에 설치 |
| `make run` | 빌드 후 실행 (도움말 표시) |
| `make test` | 단위 테스트 실행 |
| `make fmt` | 코드 포맷팅 |
| `make clean` | 빌드 결과물 삭제 |
| `make help` | 사용 가능한 명령어 목록 표시 |

### 테스트 실행

```bash
make test
# 또는
go test ./...
```

---

## 프로젝트 구조

```
note_cli/
├── main.go              # 진입점
├── go.mod               # Go 모듈 정의
├── Makefile             # 빌드/테스트 자동화
├── CLIENT_API_GUIDE.md  # API 명세서 (한국어)
├── api/
│   ├── client.go        # HTTP 클라이언트 (자동 토큰 갱신 포함)
│   ├── auth.go          # 인증 관련 API 호출
│   ├── board.go         # 노트 CRUD API 호출
│   ├── models.go        # 데이터 구조체 정의
│   └── api_test.go      # 단위 테스트
├── cmd/
│   ├── root.go          # 루트 명령어 설정 (Cobra)
│   ├── register.go      # register 명령어
│   ├── login.go         # login 명령어
│   ├── add.go           # add 명령어
│   ├── list.go          # list 명령어
│   ├── view.go          # view 명령어
│   ├── edit.go          # edit 명령어
│   └── delete.go        # delete 명령어
├── config/
│   └── config.go        # 설정 파일 로드/저장
└── tui/
    └── editor.go        # 외부 편집기 연동
```

---

## 주요 의존성

| 패키지 | 용도 |
|---|---|
| [cobra](https://github.com/spf13/cobra) | CLI 명령어 프레임워크 |
| [huh](https://github.com/charmbracelet/huh) | 인터랙티브 TUI 폼 |
| [bubbletea](https://github.com/charmbracelet/bubbletea) | 터미널 UI 프레임워크 |
| [lipgloss](https://github.com/charmbracelet/lipgloss) | 터미널 스타일링 |

---

## 라이선스

이 프로젝트의 라이선스 정보는 저장소를 확인하세요.
