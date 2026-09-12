# 기여 가이드

`note_cli`에 관심을 가져주셔서 감사합니다. 이 문서는 개발 환경과 기여 절차를 설명합니다.
개발 규칙(개발 철학, 코딩 규칙, TDD)은 [개발 가이드](../AGENTS.md)를 함께 참고하세요.

## 프로젝트 소개

`note_cli`는 로컬 SQLite 또는 원격 API를 사용하는 터미널 기반 노트 에디터입니다. 기능과 사용법은 [README](../README.md)를 참고하세요.

- 언어: Go (`go.mod`의 `go` 지시문 참고)
- CLI 프레임워크: [wcli](https://github.com/wkqco33/wcli)
- TUI/폼: [huh](https://github.com/charmbracelet/huh), [bubbletea](https://github.com/charmbracelet/bubbletea), [glamour](https://github.com/charmbracelet/glamour), [lipgloss](https://github.com/charmbracelet/lipgloss)
- 도구: `task` (Taskfile.yml), `go vet`, `gofmt`, `golangci-lint`

## 개발 환경

Go 1.26 이상이 필요합니다. 빌드 자동화는 [Taskfile](https://taskfile.dev)을 사용합니다 (`go install github.com/go-task/task/v3/cmd/task@latest`).

```bash
git clone https://github.com/wkqco33/note_cli.git
cd note_cli

task build     # ncli 바이너리 빌드
task test      # 테스트 (race + coverage)
```

## 프로젝트 구조

```text
note_cli/
├── main.go          # 진입점
├── go.mod, go.sum   # Go 모듈
├── Taskfile.yml     # 빌드/테스트 자동화 (task)
├── AGENTS.md        # 개발 가이드 (에이전트·개발자)
├── docs/            # 프로젝트 문서
├── .github/         # CI, 릴리스, 이슈/PR 템플릿
├── api/             # 백엔드 API 클라이언트 (요청, 응답 파싱, 모델)
├── attachments/     # 첨부파일 텍스트 추출
├── cmd/             # 커맨드 정의 + 커맨드 전용 로직/헬퍼
├── config/          # 설정 로드/저장, 비밀값 암호화, 빌드 정보
├── embedding/       # 임베딩 벡터 및 유사도 계산
├── llm/             # LLM 클라이언트와 프롬프트
├── local/           # 로컬 SQLite 노트 저장소
├── tui/             # 외부 편집기 연동 등 터미널 UI
└── utils/           # 로거, 스피너 등 공통 유틸
```

주요 `cmd/` 파일:

- `root.go`: 루트 커맨드와 전역 플래그(`--yes`, `--no-input`, `--quiet`, `--no-color`, `--no-pager`, `--debug`, `--version`)
- `interactive.go`: 프롬프트 정책과 확인 헬퍼
- `output.go`: stdout/stderr 출력 헬퍼와 진행률 표시줄
- `pager.go`: 페이저 실행
- `store.go`: 원격/로컬 공통 `NoteStore` 인터페이스
- `helpers.go`: 검증/포맷/TUI 선택 공용 헬퍼

## 주요 의존성

| 패키지 | 용도 |
| --- | --- |
| [wcli](https://github.com/wkqco33/wcli) | CLI 커맨드 프레임워크 |
| [huh](https://github.com/charmbracelet/huh) | 인터랙티브 TUI 폼 |
| [glamour](https://github.com/charmbracelet/glamour) | Markdown 렌더링 |
| [lipgloss](https://github.com/charmbracelet/lipgloss) | 터미널 스타일링 |
| [progressbar/v3](https://github.com/schollz/progressbar) | 업로드/다운로드 진행률 바 |
| [humanize](https://github.com/dustin/go-humanize) | 바이트 단위 표현 |
| [yaml.v3](https://github.com/go-yaml/yaml) | 설정 파일 직렬화 |
| [modernc.org/sqlite](https://modernc.org/sqlite) | 로컬 SQLite 데이터베이스 |
| [tdraw](https://github.com/wkqco33/tdraw) | 터미널 이미지 렌더링 |

## 검증 명령

PR을 보내기 전에 다음을 모두 통과해야 합니다.

```bash
go vet ./...
gofmt -l ./...          # 빈 출력이어야 합니다
go test -race ./...
golangci-lint run ./...
```

`task`를 사용할 때:

```bash
task test    # go test -v -race -cover ./...
task build   # ncli 바이너리 빌드
task fmt     # go fmt ./...
```

## PR 제출 절차

1. 저장소를 fork하고 `master`에서 작업 브랜치를 만듭니다.
2. 변경 전에 테스트를 먼저 작성합니다 (TDD). 자세한 내용은 [개발 가이드](../AGENTS.md)를 참고하세요.
3. 변경 후 위의 "검증 명령"을 모두 통과시킵니다.
4. 커밋 메시지를 작성하고 PR을 보냅니다. PR 설명에 변경 이유와 테스트 방법을 적어주세요.
5. PR 템플릿의 체크리스트를 채웁니다.

## 커밋 메시지 규칙

커밋 메시지는 한국어로 간결하게 작성하고 변경 내용을 요약합니다. 관례적으로 `type(scope): 요약` 형식을 사용합니다.

```text
feat(search): 내용 필터링 추가
fix(import): 임시 파일 누수 수정
docs: CLI 가이드라인 문서 통합
```

## 이슈 보고

버그를 발견하면 [버그 리포트 템플릿](https://github.com/wkqco33/note_cli/issues/new?template=bug_report.yml)으로 이슈를 만들어 주세요. 가능하면 다음을 포함해 주세요.

- 재현 단계
- 기대한 동작과 실제 동작
- 실행 환경(OS, Go 버전, 터미널)
- 관련 로그와 에러 메시지 (비밀값 제외)

기능 제안은 [기능 제안 템플릿](https://github.com/wkqco33/note_cli/issues/new?template=feature_request.yml)을 사용해 주세요.
보안 취약점은 이슈에 작성하지 말고 [보안 정책](./SECURITY.md)의 절차를 따라주세요.

## 주의사항

- `go mod tidy`로 `go.mod`/`go.sum`이 불필요하게 변경되지 않도록 주의하세요. 작업에 필요하지 않은 변경은 되돌립니다.
- CI(`.github/workflows/ci.yml`)는 `gofmt`, `go vet`, `golangci-lint`, `go test -race`를 실행합니다.
- 릴리스(`.github/workflows/release.yml`)는 `v*` 태그 푸시 시 `go vet` + `go test -race`를 실행한 뒤 크로스플랫폼 바이너리와 GitHub Release를 생성합니다.

## 관련 문서

- [개발 가이드](../AGENTS.md)
- [CLI 가이드라인 준수](./CLI_GUIDELINES.md)
- [API 사용 가이드](./CLIENT_API_GUIDE.md)
- [보안 정책](./SECURITY.md)
- [행동 강령](./CODE_OF_CONDUCT.md)
- [문서 인덱스](./README.md)
