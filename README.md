# 📝 Note CLI

> **Note**: 오픈소스로 공개하기 위한 문서가 준비되어 있습니다. 기여는 [CONTRIBUTING.md](./CONTRIBUTING.md)를, 보안 취약점 신고는 [SECURITY.md](./SECURITY.md)를 참고하세요.

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![CI](https://github.com/wkqco33/note_cli/actions/workflows/ci.yml/badge.svg)](https://github.com/wkqco33/note_cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/wkqco33/note_cli)](https://github.com/wkqco33/note_cli/releases)

터미널에서 빠르게 노트를 작성하고 관리할 수 있는 CLI 도구입니다.  
기본적으로 로컬 SQLite 데이터베이스를 사용하며, 필요하면 원격 API 서버에 연결해 사용자별 노트를 관리할 수 있습니다.

---

## 주요 기능

- **노트 CRUD**: 노트 생성, 조회, 수정, 삭제
- **카테고리 분류**: Work / Personal / Idea / Other
- **외부 에디터 지원**: `$EDITOR` 환경변수에 설정된 편집기로 노트 내용 작성 (기본값: `vim`)
- **인터랙티브 UI**: [Charmbracelet](https://charm.sh) 라이브러리 기반의 TUI 폼 및 프로그래스 바(Progress Bar) 지원
- **강력한 검색**: 노트 제목, 내용, 첨부파일 이름 기반 통합 검색 및 로컬 임베딩 기반 자연어 검색
- **대용량 파일 지원**: 스트리밍 전송 기반으로 최대 500MB까지 파일 업로드/다운로드 지원
- **JWT 인증**: 액세스 토큰 만료 시 자동 재발급
- **저장 모드 선택**: 로컬 SQLite 또는 원격 API 사용
- **주문형 로컬 설정**: `~/.config/note_cli/config.yaml`에 저장 모드, 호스트/포트 및 인증 토큰 저장
- **LLM 노트 보조**: 노트 개선안, 할 일 목록, 텍스트 첨부파일 요약 (Ollama / OpenAI)

---

## 요구 사항

- [Go](https://golang.org/) 1.26 이상 (go.mod의 `go` 지시자 참고)
- `vim`, `nano` 등의 외부 편집기 (또는 `$EDITOR` 환경변수 설정)
- 원격 모드 사용 시 노트 데이터를 저장할 API 서버 (`http://127.0.0.1:8880/api/v1`)

> API 서버 구성에 대한 자세한 내용은 [CLIENT_API_GUIDE.md](./CLIENT_API_GUIDE.md)를 참고하세요.

---

## 설치

```bash
# 저장소 클론
git clone https://github.com/wkqco33/note_cli.git
cd note_cli

# 빌드 (./ncli 바이너리 생성)
task build

# 또는 /usr/local/bin에 설치
task install
```

---

## 빠른 시작

```bash
# 로컬 모드에서는 API 서버나 계정 없이 바로 사용할 수 있습니다.
# 1. 노트 추가
ncli add

# 2. 노트 목록 확인
ncli list
```

로컬 데이터베이스는 `~/.local/share/note_cli/notes.db`에, 첨부파일은 같은 경로의 `files/` 디렉토리에 저장됩니다.

원격 API를 사용하려면 저장 모드를 변경한 뒤 회원가입과 로그인을 실행합니다.

```bash
ncli config set mode remote
ncli register
ncli login
ncli list
```

로컬 모드로 돌아오려면 다음을 실행합니다.

```bash
ncli config set mode local
```

로컬과 원격 데이터는 자동으로 동기화되지 않습니다. 저장소 간 데이터 이동은 `export`와 `import`를 사용하세요.

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

#### LLM 기능

LLM 기능은 기본적으로 로컬 Ollama를 사용합니다. Ollama를 실행하고 모델을 준비한 뒤 사용할 수 있습니다.

```bash
ollama serve
ollama pull llama3.2
ncli ai status
```

노트 개선안은 기본적으로 출력만 하며, `--apply`를 지정해야 저장합니다.

```bash
ncli ai improve 12
ncli ai improve
ncli ai improve 12 --apply
```

노트에서 명시된 할 일을 추출할 수 있습니다. `--create-note`를 지정하면 결과를 새 `Other` 카테고리 노트로 저장합니다.

```bash
ncli ai todos 12
ncli ai todos
ncli ai todos 12 --create-note
ncli ai summarize-file 3
ncli ai summarize-file
```

`ai improve`, `ai todos`, `ai summarize-file`은 ID를 생략하면 TUI에서 대상 노트 또는 첨부파일을 선택합니다. ID를 직접 지정할 수도 있습니다. `summarize-file`은 첨부파일 ID를 사용합니다. 현재 텍스트 기반 파일(`txt`, `md`, `csv`, `json`, `log`)을 지원하며, 분석 대상 파일은 2MB 이하입니다.

로컬 노트는 임베딩 인덱스를 생성한 뒤 자연어 검색을 사용할 수 있습니다.

```bash
ncli ai index
ncli ai index 12
ncli ai search "지난 회의에서 API 일정이 어떻게 결정됐지"
```

`ai index`는 전체 노트를 인덱싱하고, ID를 지정하면 해당 노트만 인덱싱합니다. 노트 본문이 변경되거나 임베딩 모델을 바꾸면 해당 노트만 다시 인덱싱됩니다. 임베딩 인덱싱과 자연어 검색은 현재 로컬 SQLite 모드에서 지원합니다.

OpenAI를 사용하려면 다음과 같이 설정합니다. API 키는 기존 설정의 비밀값 저장 방식으로 암호화됩니다.

```bash
ncli config set llm.provider openai
ncli config set llm.model gpt-4o-mini
ncli config set llm.base_url https://api.openai.com/v1
ncli config set llm.api_key sk-...
```

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
mode: local
host: 127.0.0.1
port: 8880
access_token: "enc:..."
refresh_token: "enc:..."
auto_login: false
username: "user@example.com"
password: "enc:..."
llm:
  provider: ollama
  model: llama3.2
  base_url: "http://127.0.0.1:11434/v1"
  embedding_model: nomic-embed-text
  timeout_seconds: 120
  # OpenAI 사용 시 설정. 실제 저장 시 enc: 형식으로 암호화됩니다.
  api_key: "enc:..."
```

> **보안 참고**: 액세스 토큰, 리프레시 토큰, 자동 로그인 비밀번호는 평문이 아니라 **`enc:` 접두사와 함께 플랫폼별 암호화**되어 저장됩니다 (Windows는 DPAPI, 그 외는 AES-256-GCM). 설정 파일 권한은 `0600`으로 생성됩니다. 보안 취약점 신고는 [SECURITY.md](./SECURITY.md)를 참고하세요.

- **mode**: 저장 모드 (`local` 또는 `remote`, 기본값 `local`)
- **host / port**: API 서버 주소를 변경할 때 수정합니다. 기본값은 `127.0.0.1` 및 `8880` 입니다.
- **auto_login**: 켜면 로그인 시 계정 정보를 저장해 인증 만료 시 자동 재로그인합니다.
- **액세스 토큰**: 유효 기간 30분, 만료 시 자동 재발급
- **리프레시 토큰**: 유효 기간 7일
- **llm.provider**: LLM 제공자 (`ollama` 또는 `openai`, 기본값 `ollama`)
- **llm.model**: 요약·개선·할 일 추출에 사용할 모델 (기본값 `llama3.2`)
- **llm.base_url**: OpenAI 호환 API 주소 (기본값 `http://127.0.0.1:11434/v1`)
- **llm.embedding_model**: 자연어 검색용 임베딩 모델 (기본값 `nomic-embed-text`)
- **llm.timeout_seconds**: LLM 요청 제한 시간 (기본값 120초)
- **llm.api_key**: OpenAI API 키. 설정 파일에는 암호화되어 저장됩니다.

LLM 관련 설정은 다음 명령으로 수정할 수 있습니다.

```bash
ncli config set llm.provider ollama
ncli config set llm.model llama3.2
ncli config set llm.base_url http://127.0.0.1:11434/v1
ncli config set llm.embedding_model nomic-embed-text
ncli config set llm.api_key sk-...
```

`llm.api_key`는 OpenAI를 사용할 때만 필요합니다. Ollama는 로컬 서버를 사용하므로 API 키가 필요하지 않습니다.

저장소의 `config.yaml.example` 파일을 참고하여 설정 파일을 직접 생성할 수도 있습니다.

---

## 개발

### 빌드 및 관련 명령어

이 저장소는 [Taskfile](https://taskfile.dev) 기반으로 빌드/테스트 자동화를 제공합니다. `task` 명령이 필요합니다 (`go install github.com/go-task/task/v3/cmd/task@latest`).

| 명령어           | 설명                         |
| ---------------- | ---------------------------- |
| `task build`     | `./ncli` 바이너리 빌드       |
| `task install`   | /usr/local/bin에 설치        |
| `task uninstall` | /usr/local/bin에서 제거      |
| `task run`       | 빌드 후 실행 (도움말 표시)   |
| `task test`      | 단위 테스트 실행             |
| `task fmt`       | 코드 포맷팅                  |
| `task clean`     | 빌드 결과물 삭제             |
| `task help`      | 사용 가능한 명령어 목록 표시 |

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
├── .github/workflows/            # CI 및 태그 기반 크로스플랫폼 릴리스
├── api/
│   ├── client.go                 # HTTP 클라이언트 (자동 토큰 갱신, 타임아웃)
│   ├── auth.go                   # 인증 관련 API 호출
│   ├── board.go                  # 노트 CRUD API 호출
│   ├── file.go                   # 파일 업로드/다운로드/삭제 (스트리밍)
│   ├── models.go                 # 데이터 구조체 및 APIError 정의
│   ├── decode.go                 # JSON 역직렬화 공용 헬퍼
│   ├── api_test.go               # 모델 직렬화 단위 테스트
│   └── client_test.go            # 401 재시도/스트림/본문 재전송 단위 테스트
├── cmd/
│   ├── root.go                   # 루트 명령어 (Cobra)
│   ├── store.go                  # 원격/로컬 공통 NoteStore 인터페이스
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
│   └── *_test.go                 # 커맨드 단위 테스트
├── config/
│   ├── config.go                 # 설정 파일 로드/저장 (~/.config/note_cli/config.yaml)
│   ├── secure.go                 # 비밀값 암호화/복호화 (플랫폼별 위임)
│   ├── secure_fallback.go        # 비 Windows AES-256-GCM 암호화
│   ├── secure_windows.go         # Windows DPAPI 암호화
│   ├── config_test.go            # 설정 단위 테스트
│   └── buildinfo/                # ldflags 주입용 빌드 정보
├── local/
│   ├── store.go                  # 로컬 SQLite 노트 저장소
│   └── store_test.go             # 로컬 저장소 단위 테스트
├── tui/
│   └── editor.go                 # 외부 편집기 연동
└── utils/
    └── logger.go                 # slog 기반 디버그 로거
```

---

## 주요 의존성

| 패키지                                                   | 용도                         |
| -------------------------------------------------------- | ---------------------------- |
| [wcli](https://github.com/wkqco33/wcli)                  | CLI 명령어 프레임워크        |
| [huh](https://github.com/charmbracelet/huh)              | 인터랙티브 TUI 폼            |
| [lipgloss](https://github.com/charmbracelet/lipgloss)    | 터미널 스타일링              |
| [glamour](https://github.com/charmbracelet/glamour)      | Markdown 렌더링              |
| [progressbar/v3](https://github.com/schollz/progressbar) | 업로드/다운로드 진행률 바    |
| [humanize](https://github.com/dustin/go-humanize)        | 바이트 단위 사람 친화적 표현 |
| [yaml.v3](https://github.com/go-yaml/yaml)               | 설정 파일 직렬화             |
| [modernc.org/sqlite](https://modernc.org/sqlite)         | 로컬 SQLite 데이터베이스     |
| [tdraw](https://github.com/wkqco33/tdraw)                | 터미널 이미지 렌더링         |

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
