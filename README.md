# 📝 Note CLI

> **Note**: 오픈소스로 공개하기 위한 문서가 준비되어 있습니다. 기여는 [CONTRIBUTING.md](./CONTRIBUTING.md)를, 보안 취약점 신고는 [SECURITY.md](./SECURITY.md)를 참고하세요.

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![CI](https://github.com/wkqco33/note_cli/actions/workflows/ci.yml/badge.svg)](https://github.com/wkqco33/note_cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/wkqco33/note_cli)](https://github.com/wkqco33/note_cli/releases)

터미널에서 빠르게 노트를 작성하고 관리할 수 있는 CLI 도구입니다.  
원격 API 서버에 노트를 저장하며, JWT 기반 인증을 통해 사용자별 노트를 안전하게 관리합니다.

---

## 주요 기능

- **노트 CRUD**: 노트 생성, 조회, 수정, 삭제
- **카테고리 분류**: Work / Personal / Idea / Other
- **외부 에디터 지원**: `$EDITOR` 환경변수에 설정된 편집기로 노트 내용 작성 (기본값: `vim`)
- **인터랙티브 UI**: [Charmbracelet](https://charm.sh) 라이브러리 기반의 TUI 폼 및 프로그래스 바(Progress Bar) 지원
- **강력한 검색**: 노트 제목, 내용, 첨부파일 이름 기반 통합 검색 기능
- **대용량 파일 지원**: 스트리밍 전송 기반으로 최대 500MB까지 파일 업로드/다운로드 지원
- **JWT 인증**: 액세스 토큰 만료 시 자동 재발급
- **주문형 로컬 설정**: `~/.config/note_cli/config.yaml`에 호스트/포트 및 인증 토큰 저장

---

## 요구 사항

- [Go](https://golang.org/) 1.26 이상 (go.mod의 `go` 지시자 참고)
- `vim`, `nano` 등의 외부 편집기 (또는 `$EDITOR` 환경변수 설정)
- 노트 데이터를 저장할 API 서버 (`http://127.0.0.1:8880/api/v1`)

> API 서버 구성에 대한 자세한 내용은 [CLIENT_API_GUIDE.md](./CLIENT_API_GUIDE.md)를 참고하세요.

---

## 설치

```bash
# 저장소 클론 (서브모듈 포함)
git clone --recurse-submodules https://github.com/wkqco33/note_cli.git
cd note_cli

# 이미 클론한 경우 서브모듈 초기화
git submodule update --init --recursive

# 빌드 (./ncli 바이너리 생성)
task build

# 또는 Go bin 경로에 설치 (~/.go/bin 또는 $GOPATH/bin)
task install
```

> **참고**: `view` 명령의 터미널 이미지 렌더링 기능은 [tdraw](https://github.com/wkqco33/tdraw) 서브모듈에
> 의존합니다. 서브모듈 없이 빌드하면 `go build`가 실패하므로 반드시 `--recurse-submodules`로
> 클론하거나 `git submodule update --init --recursive`를 실행하세요.

---

## 빠른 시작

```bash
# 1. 계정 등록
ncli register

# 2. 로그인
ncli login

# 3. 노트 추가
ncli add

# 4. 노트 목록 확인
ncli list
```

---

## 사용법

### 계정 관리

#### 회원가입

```bash
ncli register
```

이름, 이메일, 비밀번호를 입력하는 인터랙티브 폼이 표시됩니다.

#### 로그인

```bash
ncli login
```

이메일과 비밀번호를 입력하면 인증 토큰이 `~/.config/note_cli/config.yaml`에 저장됩니다.

---

### 노트 관리

#### 노트 목록 조회

```bash
ncli list
```

사용자의 모든 노트를 테이블 형식으로 출력합니다.

```bash
ID   TITLE          CATEGORY   UPDATED
1    회의 메모        Work       2026-03-10
2    아이디어 정리     Idea       2026-03-09
```

#### 노트 추가

```bash
ncli add
```

제목과 카테고리를 입력하는 폼이 나타난 후, 외부 편집기(기본: `vim`)가 열려 내용을 작성합니다.

카테고리 선택지:

- `Work`
- `Personal`
- `Idea`
- `Other`

#### 노트 조회

```bash
ncli view [ID]
```

ID를 입력하지 않을 경우 TUI 목록에서 조회할 노트를 선택할 수 있습니다. 조회 시 노트의 원본 첨부파일 이름과 정보가 함께 표시됩니다.

예시:

```bash
ncli view
ncli view 1
```

#### 노트 및 첨부파일 수정/다운로드/검색

```bash
ncli edit [ID]
ncli download [ID]
ncli search [flags]
```

- `edit`: 기존 노트의 제목, 카테고리, 내용, 첨부파일을 수정할 수 있습니다. ID 생략 시 TUI 화면에서 수정할 노트를 고를 수 있습니다.
- `download`: 노트의 첨부파일을 다운로드합니다. 파일 전송에는 진행률 표시줄(Progress bar)이 제공되며, ID 생략 시 전체 파일 목록을 TUI 기반으로 탐색하여 다운로드할 수 있습니다. 원본 파일명으로 저장됩니다.
- `search`: `--title (제목)`, `--content (내용)`, `--file (첨부파일명)` 플래그를 조합하여 노트를 빠르게 검색할 수 있습니다. 아무 플래그도 입력하지 않으면 대화형(TUI) 방식으로 검색 조건을 선택할 수 있습니다.

#### 노트 및 첨부파일 삭제

```bash
ncli delete [ID]
```

원하는 노트 문서 전체를 지우거나, 첨부파일만 개별적으로 선택해 삭제할 수 있습니다. ID 없이 명령어만 실행하면 대화형 UI(TUI)를 통해 안전하게 삭제 대상을 확인 후 선택할 수 있습니다 (`--file` 플래그로 파일 삭제 모드 강제 가능).

예시:

```bash
ncli delete
ncli delete 1
ncli delete --file
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

앱을 처음 실행하거나 로그인하면 `~/.config/note_cli/config.yaml` 파일이 자동으로 생성됩니다.

```yaml
host: 127.0.0.1
port: 8880
access_token: "enc:..."
refresh_token: "enc:..."
```

> **보안 참고**: 액세스 토큰, 리프레시 토큰, 자동 로그인 비밀번호는 평문이 아니라 **`enc:` 접두사와 함께 플랫폼별 암호화**되어 저장됩니다 (Windows는 DPAPI, 그 외는 AES-256-GCM). 설정 파일 권한은 `0600`으로 생성됩니다. 보안 취약점 신고는 [SECURITY.md](./SECURITY.md)를 참고하세요.

- **host / port**: API 서버 주소를 변경할 때 수정합니다. 기본값은 `127.0.0.1` 및 `8880` 입니다.
- **액세스 토큰**: 유효 기간 30분, 만료 시 자동 재발급
- **리프레시 토큰**: 유효 기간 7일

저장소의 `config.yaml.example` 파일을 참고하여 설정 파일을 직접 생성할 수도 있습니다.

### API 키 / 시크릿 키

릴리스 바이너리는 API 키와 시크릿 키가 **빌드 시점에 내장**되도록 빌드할 수 있습니다.  
그래서 GitHub Actions에서 생성한 바이너리는 별도 `.env` 없이 바로 실행할 수 있습니다.

로컬 개발이나 별도 빌드에서는 저장소 루트의 `.env` 파일 또는 셸 환경변수를 사용할 수 있습니다.

```bash
cp .env.example .env
```

`.env` 또는 셸 환경에 아래 값을 설정하세요.

```bash
NOTE_CLI_API_KEY="..."
NOTE_CLI_SECRET_KEY="..."
```

빌드 시점에 값이 주입되면 실행 시 `.env`는 필요 없습니다.  
실행 시 환경변수가 있으면 빌드타임 내장값보다 **우선 적용**되므로, 운영 환경이나 테스트 환경에서 override 용도로 사용할 수 있습니다.

---

## 개발

### 빌드 및 관련 명령어

이 저장소는 [Taskfile](https://taskfile.dev) 기반으로 빌드/테스트 자동화를 제공합니다. `task` 명령이 필요합니다 (`go install github.com/go-task/task/v3/cmd/task@latest`).

| 명령어 | 설명 |
| - | - |
| `task build` | `./ncli` 바이너리 빌드 |
| `task install` | Go bin 경로에 설치 |
| `task uninstall` | Go bin 경로에서 제거 |
| `task run` | 빌드 후 실행 (도움말 표시) |
| `task test` | 단위 테스트 실행 |
| `task fmt` | 코드 포맷팅 |
| `task clean` | 빌드 결과물 삭제 |
| `task help` | 사용 가능한 명령어 목록 표시 |

### 테스트 실행

```bash
task test
# 또는
go test ./...
```

---

## 프로젝트 구조

```bash
note_cli/
├── main.go                       # 진입점
├── go.mod                        # Go 모듈 정의
├── Taskfile.yml                  # 빌드/테스트 자동화 (task)
├── CLIENT_API_GUIDE.md           # API 명세서 (한국어)
├── .github/workflows/release.yml # 태그 기반 크로스플랫폼 릴리스
├── api/
│   ├── client.go                 # HTTP 클라이언트 (자동 토큰 갱신, 타임아웃, 시크릿 캐싱)
│   ├── auth.go                   # 인증 관련 API 호출
│   ├── board.go                  # 노트 CRUD API 호출
│   ├── file.go                   # 파일 업로드/다운로드/삭제 (스트리밍)
│   ├── models.go                 # 데이터 구조체 및 APIError 정의
│   ├── decode.go                 # JSON 역직렬화 공용 헬퍼
│   ├── api_test.go               # 모델 직렬화 단위 테스트
│   └── client_test.go            # 401 재시도/스트림/본문 재전송 단위 테스트
├── cmd/
│   ├── root.go                   # 루트 명령어 (Cobra, RunE 전파)
│   ├── helpers.go                # 검증/포맷/TUI 선택 공용 헬퍼
│   ├── register.go               # register 명령어
│   ├── login.go                  # login 명령어
│   ├── add.go                    # add 명령어
│   ├── list.go                   # list 명령어
│   ├── view.go                   # view 명령어 (이미지 렌더링)
│   ├── edit.go                   # edit 명령어
│   ├── delete.go                 # delete 명령어 (노트/파일)
│   ├── download.go               # download 명령어
│   ├── search.go                 # search 명령어 (제목/내용/파일명)
│   ├── export.go                 # export 백업 명령어
│   ├── import.go                 # import 복원 명령어
│   ├── config.go                 # config 명령어 (설정 조회/수정)
│   ├── clean.go                  # 임시 캐시 정리 명령어
│   ├── version.go                # version 명령어
│   ├── helpers_test.go           # 헬퍼 단위 테스트
│   ├── search_test.go            # 검색 필터링 단위 테스트
│   ├── export_test.go            # 백업 직렬화 단위 테스트
│   ├── import_test.go            # 복원 매핑/아카이브 탐색 단위 테스트
│   ├── clean_test.go             # 캐시 정리 단위 테스트
│   ├── config_test.go            # 설정 적용 단위 테스트
│   └── delete_test.go            # 삭제 대상 매핑 단위 테스트
├── config/
│   ├── config.go                 # 설정 파일 로드/저장 (~/.config/note_cli/config.yaml)
│   ├── secrets.go                # 런타임 시크릿 해석 (env > .env > 빌드타임 내장)
│   ├── secure.go                 # 비밀값 암호화/복호화 (플랫폼별 위임)
│   ├── secure_fallback.go        # 비 Windows AES-256-GCM 암호화
│   ├── secure_windows.go         # Windows DPAPI 암호화
│   ├── secrets_test.go           # 시크릿 해석 단위 테스트
│   └── buildinfo/                # ldflags 주입용 빌드 정보
├── tui/
│   └── editor.go                 # 외부 편집기 연동
├── utils/
│   └── logger.go                 # slog 기반 디버그 로거
└── tdraw/                        # 터미널 이미지 렌더링 (git submodule)
```

---

## 주요 의존성

| 패키지 | 용도 |
| - | - |
| [cobra](https://github.com/spf13/cobra) | CLI 명령어 프레임워크 |
| [huh](https://github.com/charmbracelet/huh) | 인터랙티브 TUI 폼 |
| [lipgloss](https://github.com/charmbracelet/lipgloss) | 터미널 스타일링 |
| [glamour](https://github.com/charmbracelet/glamour) | Markdown 렌더링 |
| [progressbar/v3](https://github.com/schollz/progressbar) | 업로드/다운로드 진행률 바 |
| [humanize](https://github.com/dustin/go-humanize) | 바이트 단위 사람 친화적 표현 |
| [godotenv](https://github.com/joho/godotenv) | `.env` 파일 로드 |
| [yaml.v3](https://github.com/go-yaml/yaml) | 설정 파일 직렬화 |
| [tdraw](https://github.com/wkqco33/tdraw) | 터미널 이미지 렌더링 (submodule) |

---

## 라이선스

이 프로젝트는 [MIT 라이선스](./LICENSE) 하에 배포됩니다.

---

## 기여

기여를 환영합니다. 자세한 개발 방식(TDD, 코딩 규칙, 커밋 규칙)과 PR 절차는 [CONTRIBUTING.md](./CONTRIBUTING.md)를 참고하세요.

- [버그 리포트](./.github/ISSUE_TEMPLATE/bug_report.yml)
- [기능 제안](./.github/ISSUE_TEMPLATE/feature_request.yml)
- [기여자 행동 강령](./CODE_OF_CONDUCT.md)
- [보안 정책](./SECURITY.md)
