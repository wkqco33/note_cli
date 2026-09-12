# 📝 Note CLI

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![CI](https://github.com/wkqco33/note_cli/actions/workflows/ci.yml/badge.svg)](https://github.com/wkqco33/note_cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/wkqco33/note_cli)](https://github.com/wkqco33/note_cli/releases)

터미널에서 빠르게 노트를 작성하고 관리하는 CLI 도구입니다. 기본적으로 로컬 SQLite 데이터베이스를 사용하며, 필요하면 원격 API 서버에 연결해 사용자별 노트를 관리할 수 있습니다.

| 문서 | 설명 |
| --- | --- |
| [문서 인덱스](./docs/README.md) | 전체 문서 목록 |
| [기여 가이드](./docs/CONTRIBUTING.md) | 개발 환경, 이슈/PR 절차 |
| [CLI 가이드라인 준수](./docs/CLI_GUIDELINES.md) | 입력·출력·종료 코드 규칙 |
| [API 사용 가이드](./docs/CLIENT_API_GUIDE.md) | 서버 API 명세 |
| [개발 가이드](./AGENTS.md) | 개발 철학, 코딩 규칙, TDD (에이전트·개발자 공용) |
| [보안 정책](./docs/SECURITY.md) | 취약점 신고 절차 |
| [행동 강령](./docs/CODE_OF_CONDUCT.md) | 커뮤니티 기준 |

---

## 주요 기능

- **노트 CRUD**: 노트 생성, 조회, 수정, 삭제
- **카테고리 분류**: Work / Personal / Idea / Other
- **외부 에디터 지원**: `$EDITOR` 환경변수에 설정된 편집기로 노트 내용 작성 (기본값: `vim`)
- **인터랙티브 UI**: [Charmbracelet](https://charm.sh) 라이브러리 기반의 TUI 폼 및 진행률 표시줄
- **강력한 검색**: 노트 제목·내용·첨부파일 이름 기반 통합 검색, 로컬 임베딩 기반 자연어 검색
- **대용량 파일 지원**: 스트리밍 전송 기반으로 최대 500MB까지 파일 업로드/다운로드
- **JWT 인증**: 액세스 토큰 만료 시 자동 재발급
- **저장 모드 선택**: 로컬 SQLite 또는 원격 API 사용
- **LLM 노트 보조**: 노트 개선안, 할 일 목록, 텍스트 첨부파일 요약 (Ollama / OpenAI)
- **자동화 친화**: `--json`, `--yes`, `--no-input` 등 비대화형 실행 지원

---

## 요구 사항

- [Go](https://golang.org/) 1.26 이상 (`go.mod`의 `go` 지시문 참고)
- `vim`, `nano` 등의 외부 편집기 (또는 `$EDITOR` 환경변수 설정)
- 원격 모드 사용 시 노트 데이터를 저장할 API 서버

> API 서버 구성에 대한 자세한 내용은 [API 사용 가이드](./docs/CLIENT_API_GUIDE.md)를 참고하세요.

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

빌드·테스트 명령과 프로젝트 구조는 [기여 가이드](./docs/CONTRIBUTING.md)를 참고하세요.

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

로컬 모드로 돌아가려면 다음을 실행합니다.

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

이름, 이메일, 비밀번호를 입력하는 인터랙티브 폼이 표시됩니다. 비대화형 환경에서는 플래그로 지정합니다.

```bash
ncli register --name 홍길동 --email user@example.com --password-file ~/.secrets/pw
cat pw.txt | ncli register --name 홍길동 --email user@example.com --password-stdin
```

#### 로그인

```bash
ncli login
```

이메일과 비밀번호를 입력하면 인증 토큰이 설정 파일에 저장됩니다. 비대화형 환경에서는 플래그로 지정할 수 있습니다.
비밀번호는 `--password-file` 또는 `--password-stdin`을 권장합니다. `--password`는 프로세스 목록과 셸 히스토리에 노출될 수 있어 사용 시 경고를 출력합니다.

```bash
ncli login --email user@example.com --password-file ~/.secrets/pw
cat pw.txt | ncli login --email user@example.com --password-stdin
```

---

### 노트 관리

#### 노트 목록 조회

```bash
ncli list
```

사용자의 모든 노트를 테이블 형식으로 출력합니다.

```text
ID   TITLE          CATEGORY   UPDATED
1    회의 메모        Work       2026-03-10
2    아이디어 정리     Idea       2026-03-09
```

#### 노트 추가

```bash
ncli add
```

제목과 카테고리를 입력하는 폼이 나타난 후, 외부 편집기(기본: `vim`)가 열려 내용을 작성합니다. 카테고리 선택지는 `Work`, `Personal`, `Idea`, `Other`입니다.

플래그를 지정하면 해당 항목의 폼/편집기 입력을 건너뜁니다. 모든 플래그를 지정하면 대화형 입력 없이 즉시 생성되므로 에이전트나 스크립트에서 유용합니다.

```bash
# 제목/카테고리/내용을 한 번에 지정 (비대화형)
ncli add --title "회의 메모" --content-file meeting.md --category work

# stdin으로 내용 전달
echo "할 일 목록" | ncli add -t "할 일" -c -

# 제목만 지정 → 내용은 편집기에서 작성
ncli add --title "아이디어"

# 카테고리 미지정 시 other로 생성
ncli add -t "메모" -c "내용"
```

플래그: `--title/-t`, `--content/-c` (`-`면 stdin), `--content-file`, `--category`, `--file/-f`

> 비대화형 환경에서 카테고리를 지정하지 않으면 기본값 `other`를 사용하며 프롬프트를 띄우지 않습니다. 제목과 내용은 반드시 플래그로 지정해야 합니다.

#### 노트 조회

```bash
ncli view [ID]
```

ID를 입력하지 않으면 TUI 목록에서 조회할 노트를 선택할 수 있습니다. 조회 시 노트의 원본 첨부파일 이름과 정보가 함께 표시됩니다.

```bash
ncli view
ncli view 1
ncli view 1 --no-image   # 첨부 이미지를 터미널에 렌더링하지 않음
```

#### 노트 수정

```bash
ncli edit [ID]
```

기존 노트의 제목, 카테고리, 내용, 첨부파일을 수정할 수 있습니다. ID를 생략하면 TUI 화면에서 수정할 노트를 고를 수 있습니다.
`add`와 동일하게 `--title/-t`, `--content/-c`, `--content-file`, `--category` 플래그를 지원하며, 지정하지 않은 필드는 기존 값이 유지됩니다.

```bash
# 제목만 변경 (내용/카테고리는 기존 값 유지)
ncli edit 3 --title "수정된 제목"

# 파일의 내용으로 교체
ncli edit 3 --content-file new_content.md

# stdin으로 내용 교체
echo "새 내용" | ncli edit 3 -c -
```

#### 첨부파일 다운로드

```bash
ncli download [FILE_ID]
```

첨부파일을 현재 디렉토리에 원본 파일명으로 다운로드합니다. ID를 생략하면 전체 파일 목록에서 선택합니다.

#### 검색

```bash
ncli search [flags]
```

`--title`(제목), `--content`(내용), `--file`(첨부파일명)을 조합해 노트를 검색합니다. 아무 플래그도 지정하지 않으면 TUI에서 검색 조건을 입력합니다.

```bash
ncli search --title 회의
ncli search --content "API 일정"
ncli search --file report
ncli search --title 회의 --json
```

#### 노트 및 첨부파일 삭제

```bash
ncli delete [ID]
```

노트 문서 전체를 지우거나, 첨부파일만 개별적으로 삭제할 수 있습니다. ID 없이 실행하면 TUI에서 삭제 대상을 확인한 뒤 선택합니다 (`--file`로 파일 삭제 모드 강제).

```bash
ncli delete
ncli delete 1
ncli delete --file
ncli delete 1 --yes        # 확인 없이 삭제 (비대화형/CI)
ncli delete --file 2 --yes # 첨부파일 삭제
```

삭제는 항상 확인 단계를 거칩니다. 비대화형 환경에서는 `--yes`가 없으면 종료 코드 2로 실패합니다.

---

### 백업과 복원

```bash
ncli export [PATH]   # 노트와 첨부파일을 ZIP으로 백업 (기본: note_backup_<날짜>.zip)
ncli import PATH     # ZIP 백업을 서버/로컬 저장소로 복원
```

```bash
ncli export
ncli export ./backup.zip
ncli import ./backup.zip
ncli import ./backup.zip --clean --yes   # 기존 데이터를 모두 삭제한 뒤 복원
```

`import --clean`은 기존 노트와 파일을 모두 삭제하는 파괴적 작업이므로 확인을 거칩니다. 비대화형 환경에서는 `--yes`가 필요합니다.

---

### LLM 기능

LLM 기능은 기본적으로 로컬 Ollama를 사용합니다. Ollama를 실행하고 모델을 준비한 뒤 사용할 수 있습니다.

LLM이 생성하는 제목, 요약, 할 일, 변경 사항과 주의 사항은 한국어를 기준으로 출력합니다. 코드, 명령어, URL, 토큰, 파일명과 같은 식별 가능한 값은 원문을 유지합니다.

```bash
ollama serve
ollama pull llama3.2
ncli ai status
```

노트 개선안은 기본적으로 출력만 하며, `--apply`를 지정해야 저장합니다.

```bash
ncli ai improve 12
ncli ai improve
ncli ai improve 12 --comment "제목을 간결하게 작성하고 설명을 보강해줘"
ncli ai improve 12 --apply
```

노트에서 명시된 할 일을 추출할 수 있습니다. `--create-note`를 지정하면 결과를 새 `Other` 카테고리 노트로 저장합니다.

```bash
ncli ai todos 12
ncli ai todos 12 --create-note
ncli ai summarize-file 3
```

사용자 요청에 따라 LLM이 노트를 생성해 바로 저장할 수 있습니다. 인자 없이 실행하면 요청을 폼에서 입력받습니다. LLM이 제목, 본문(Markdown), 카테고리를 제안하며 `--category`로 카테고리를 강제할 수 있습니다.

```bash
ncli ai create "어제 회의 내용을 회의록으로 정리해줘"
ncli ai create --category work "오늘 한 일 정리"
ncli ai create --dry-run "아이디어 브레인스톤 노트"   # 저장하지 않고 미리보기만
```

`ai improve`, `ai todos`, `ai summarize-file`은 ID를 생략하면 TUI에서 대상 노트 또는 첨부파일을 선택합니다. `summarize-file`은 첨부파일 ID를 사용하며, 현재 텍스트 기반 파일(`txt`, `md`, `csv`, `json`, `log`)과 2MB 이하 파일을 지원합니다.

로컬 노트는 임베딩 인덱스를 생성한 뒤 자연어 검색을 사용할 수 있습니다.

```bash
ncli ai index
ncli ai index 12
ncli ai search "지난 회의에서 API 일정이 어떻게 결정됐지"
```

`ai index`는 전체 노트를 인덱싱하고, ID를 지정하면 해당 노트만 인덱싱합니다. 노트 본문이 변경되거나 임베딩 모델을 바꾸면 해당 노트만 다시 인덱싱됩니다. 임베딩 인덱싱과 자연어 검색은 현재 로컬 SQLite 모드에서 지원합니다.

OpenAI를 사용하려면 다음과 같이 설정합니다. API 키는 설정의 비밀값 저장 방식으로 암호화됩니다.

```bash
ncli config set llm.provider openai
ncli config set llm.model gpt-4o-mini
ncli config set llm.base_url https://api.openai.com/v1
ncli config set llm.api_key sk-...
```

---

### 구조화 출력 (`--format` / `--json`)

`list`, `search`, `view`는 `--format` 플래그로 출력 형식을 지정할 수 있습니다. `json` 또는 `yaml`을 지정하면 색상·표 렌더링 없이 데이터 그대로 출력되므로 스크립트나 에이전트 파이프라인에서 유용합니다.

```bash
ncli list --format json
ncli list --json            # --format json 단축
ncli list --format yaml
ncli search --title 회의 --json
ncli view 3 --json
```

- 미지정 시 사람이 읽는 형식(text)으로 출력됩니다.
- `--json`은 `--format json`과 동일하며, 둘 다 지정하면 `--json`이 우선합니다.
- `json`/`yaml` 출력 시 이미지 렌더링이나 안내 문구 없이 순수 데이터만 출력됩니다.
- 진행·상태 안내는 stderr, 실제 결과(표/JSON/YAML/노트 본문)는 stdout으로 출력되어 파이프를 오염시키지 않습니다.
- 지원하지 않는 형식을 지정하면 오류와 함께 가능한 값이 안내됩니다.

---

### 비대화형 실행 (`--yes` / `--no-input`)

ncli는 `stdin`이 터미널이 아니면(파이프, 리다이렉트, CI, 에이전트) 대화형 프롬프트를 띄우지 않습니다.
이때 입력이 필요하면 조용히 취소되지 않고, 필요한 플래그를 안내하는 오류와 함께 종료 코드 2로 종료합니다.

```bash
# 프롬프트를 띄우지 않고 즉시 실패 (종료 코드 2)
echo "" | ncli delete 3
# 대화형 입력이 필요합니다: 삭제하려면 --yes/-y 플래그를 지정하세요

# 확인 자동 승인
ncli delete 3 --yes

# 모든 프롬프트 금지 (필요한 값은 플래그로)
ncli add -t "메모" -c "내용" --no-input
```

ID를 생략하는 `view`, `download`, `delete`, `ai improve`, `ai todos`, `ai summarize-file`은 비대화형 환경에서 ID를 인자로 요구합니다.

---

### 전역 플래그와 환경변수

전역 플래그는 서브커맨드 앞뒤 어디에나 지정할 수 있습니다.

| 플래그 | 설명 |
| --- | --- |
| `--yes`, `-y` | 확인 프롬프트(삭제, `import --clean` 등)를 자동 승인합니다. |
| `--no-input` | 모든 대화형 프롬프트를 금지합니다. 값을 지정하지 않으면 오류로 안내합니다. |
| `--quiet`, `-q` | 진행·상태 안내와 진행률 표시줄을 출력하지 않습니다. |
| `--no-color` | 색상 출력을 끕니다 (`NO_COLOR`와 동일). |
| `--no-pager` | 긴 출력을 페이저로 넘기지 않습니다. |
| `--version` | 버전을 출력합니다 (`ncli version`과 동일). |
| `--help`, `-h` | 도움말을 출력합니다. `ncli help <command>`로 서브커맨드 도움말도 볼 수 있습니다. |
| `--debug` | 디버그 로그를 stderr로 출력합니다 (`DEBUG`와 동일). |

환경변수:

| 변수 | 설명 |
| --- | --- |
| `EDITOR` | 노트 내용 편집에 사용할 편집기 (기본: `vim`, Windows는 `notepad`). |
| `PAGER` | 긴 출력에 사용할 페이저 (기본: `less -FIRX`, `cat`/`none`이면 사용 안 함). |
| `NO_COLOR` | 색상 출력을 끕니다 (`--no-color`와 동일). |
| `DEBUG` | 디버그 로그를 출력합니다 (`--debug`와 동일). |
| `XDG_CONFIG_HOME` | 설정 파일 위치 변경 (기본: `~/.config`). |
| `XDG_DATA_HOME` | 로컬 데이터 위치 변경 (기본: `~/.local/share`). |
| `TMPDIR` | 임시 파일 위치 (캐시 이미지 등). |

---

### 종료 코드

| 코드 | 의미 |
| --- | --- |
| `0` | 성공 |
| `1` | 실행 오류 (검증 실패, 네트워크 오류 등) |
| `2` | 대화형 입력이 필요하지만 프롬프트를 띄울 수 없음 (플래그로 입력) |

> 대화형 사용자가 ESC/Ctrl+C로 취소하면 종료 코드 0으로 끝납니다.

---

### 기타 명령

```bash
ncli config                  # 현재 설정 조회
ncli config set <key> <value>  # 설정 변경 (mode, host, port, auto_login, llm.*)
ncli clean                   # view 실행 중 생성된 임시 캐시 이미지 정리
ncli version                 # 버전 출력
ncli help [command]          # 전체 또는 특정 커맨드 도움말
```

---

## 편집기 설정

노트 내용 작성 시 시스템의 기본 편집기를 사용합니다. 원하는 편집기를 `$EDITOR` 환경변수로 지정할 수 있습니다.

```bash
# ~/.bashrc 또는 ~/.zshrc에 추가
export EDITOR=nano
```

---

## 설정 파일

앱을 처음 실행하거나 로그인하면 설정 파일이 자동으로 생성됩니다.
기본 위치는 `~/.config/note_cli/config.yaml`이며, `XDG_CONFIG_HOME`이 설정되면 `$XDG_CONFIG_HOME/note_cli/config.yaml`을 사용합니다.
로컬 데이터베이스는 기본 `~/.local/share/note_cli/notes.db`이고, `XDG_DATA_HOME`이 설정되면 `$XDG_DATA_HOME/note_cli/notes.db`를 사용합니다.

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

- **mode**: 저장 모드 (`local` 또는 `remote`, 기본값 `local`)
- **host / port**: API 서버 주소 (기본값 `127.0.0.1`, `8880`)
- **auto_login**: 켜면 로그인 시 계정 정보를 저장해 인증 만료 시 자동 재로그인합니다.
- **액세스 토큰**: 유효 기간 30분, 만료 시 자동 재발급
- **리프레시 토큰**: 유효 기간 7일
- **llm.provider**: LLM 제공자 (`ollama` 또는 `openai`, 기본값 `ollama`)
- **llm.model**: 요약·개선·할 일 추출에 사용할 모델 (기본값 `llama3.2`)
- **llm.base_url**: OpenAI 호환 API 주소 (기본값 `http://127.0.0.1:11434/v1`)
- **llm.embedding_model**: 자연어 검색용 임베딩 모델 (기본값 `nomic-embed-text`)
- **llm.timeout_seconds**: LLM 요청 제한 시간 (기본값 120초)
- **llm.api_key**: OpenAI API 키. 설정 파일에는 암호화되어 저장됩니다.

`llm.api_key`는 OpenAI를 사용할 때만 필요합니다. Ollama는 로컬 서버를 사용하므로 API 키가 필요하지 않습니다.
저장소의 `config.yaml.example` 파일을 참고해 설정 파일을 직접 생성할 수도 있습니다.

> **보안 참고**: 액세스 토큰, 리프레시 토큰, 자동 로그인 비밀번호는 평문이 아니라 **`enc:` 접두사와 함께 플랫폼별 암호화**되어 저장됩니다 (Windows는 DPAPI, 그 외는 AES-256-GCM). 설정 파일 권한은 `0600`으로 생성됩니다. 자세한 내용은 [보안 정책](./docs/SECURITY.md)을 참고하세요.

---

## CLI 가이드라인 준수

ncli는 [Command Line Interface Guidelines](https://clig.dev/)를 참고해 설계했습니다. 프롬프트는 TTY에서만 표시하고, 결과는 stdout·진행/오류는 stderr로 분리하며, 표준 플래그(`--json`, `--quiet`, `--no-color`, `--no-input`, `--version`, `--no-pager`)와 종료 코드(0/1/2)를 제공합니다.

자세한 내용은 [CLI 가이드라인 준수](./docs/CLI_GUIDELINES.md)를 참고하세요.

---

## 라이선스

이 프로젝트는 [MIT 라이선스](./LICENSE) 하에 배포됩니다.
