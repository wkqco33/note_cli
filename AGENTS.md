# note_cli 프로젝트 개발 가이드

이 파일은 이 저장소에서 작업하는 모든 에이전트(AI, 개발자)가 동일한 개발 방식을 따르도록 하기 위한 지침입니다. 이 프로젝트에 기여하기 전에 반드시 읽어주세요.

## 프로젝트 개요

`note_cli`는 로컬 SQLite 또는 원격 API(`https://github.com/wkqco33/tdraw` 백엔드)를 사용하는 빠르고 간단한 터미널 기반 노트 에디터입니다.

- 언어: **Go** (`go.mod`의 `go` 지시문 참고)
- CLI 프레임워크: `github.com/wkqco33/wcli`
- TUI/폼: `github.com/charmbracelet/huh`, `bubbletea`, `glamour`, `lipgloss`
- 도구: `task` (Taskfile.yml), `go vet`, `gofmt` (GO 스타일)

## 디렉토리 구조

```
.
├── api/    # 백엔드 API 클라이언트 (HTTP 요청, 응답 파싱, 모델)
├── cmd/    # wcli 커맨드 정의 + 커맨드 전용 헬퍼/로직
├── config/ # 설정 로드/저장, 비밀값 관리
├── local/  # 로컬 SQLite 노트 저장소
├── tui/    # 터미널 UI (편집기 진입 등)
├── utils/  # 로거 등 공통 유틸
└── main.go # 진입점 (cmd.Execute() 호출)
```

## 개발 철학

- **탈객체지향, 데이터 중심, 성능 최우선**입니다.
- 복잡한 객체지향 설계/추상화/레이어보다 **간단한 함수형 프로그래밍과 데이터 중심 설계**를 선호합니다.
- 데이터 구조는 `struct`로 스키마를 명확히 정의하고 타입을 드러냅니다.
- I/O나 외부 의존성(wcli 실행, huh 폼, 네트워크)과 **순수 로직(파싱, 필터링, 문자열 처리)을 분리**해서 테스트하기 쉽게 만듭니다. 순수 로직은 `func`로 추출하고, `Run` 안에는 오케스트레이션만 남깁니다.
  - 예: `printBoardTable`의 실제 표 생성은 순수 함수 `buildBoardTable`로 분리
  - 예: `clean` 커맨드의 검색/삭제 로직은 `findCleanCacheFiles`, `cleanCacheFiles`로 분리
- **테스트를 위한 인터페이스/훅은 최소로만 추가**합니다 (예: `utils.SetLoggerHandler`).

## 코딩 규칙

- Go 표준 스타일(`gofmt`) 준수. `go vet` 경고가 없어야 합니다.
- **주석은 필요한 곳에만 간결하게** 한국어로 작성합니다.
- 오류는 `fmt.Errorf("... : %w", err)` 패턴으로 맥락을 붙여 감싸서 반환합니다.
- 사용자 대상 출력 문구는 한국어를 사용합니다.
- 하드코딩된 값을 상수로 추출합니다 (예: `maxAttachedFileSize`).
- 패키지 단위로 테스트 파일을 만듭니다 (`*_test.go`).

## 대화형 입력(TUI) 규칙 (중요)

`huh`/`tui` 기반 입력은 **stdin이 터미널일 때만** 허용합니다. 파이프·CI·에이전트 환경에서 조용히 취소되면 자동화가 성공으로 오인하므로, 다음 규칙을 지킵니다.

- 새 입력/선택/확인을 추가할 때는 `cmd/interactive.go`의 헬퍼(`requirePrompt`, `askConfirm`, `runNoteForm`)를 사용합니다. `huh`를 직접 호출하지 않습니다.
- 프롬프트를 띄울 수 없을 때는 `ErrInteractionRequired`(→ `interactionRequired(hint)`)를 반환해 **종료 코드 2**로 실패하게 합니다. 안내 문구에 쓸 플래그를 반드시 포함합니다(`--yes`, `--title` 등).
- 사람이 직접 취소한 경우(`huh.ErrUserAborted`)만 `handlePromptError(err, cancelMsg)`로 안내 문구를 출력하고 정상 종료(0)합니다.
- 입력이 필요한 커맨드는 **반드시 플래그/인자로 대체 경로**를 제공합니다(예: `--title`, `--yes`, `--password-stdin`). 프롬프트를 필수로 만들지 않습니다.
- TTY 판단은 `stdinIsTerminal`·`promptConfirm` 훅을 통해 테스트에서 교체합니다(실제 TTY 없이 단위 테스트).

## 출력 스트림 규칙

CLI Guidelines(clig.dev)에 맞춰 stdout과 stderr를 구분합니다.

- **stdout**: 명령의 실제 결과(노트 표, `--format json/yaml`, 노트 본문, 최종 결과 한 줄).
- **stderr**: 진행/상태 안내와 경고, 오류. `statusf`/`statusln`(`cmd/output.go`)을 사용합니다.
- 진행 메시지는 `fmt.Printf`로 stdout에 직접 쓰지 않습니다. 파이프 소비자가 오염됩니다.
- `-q/--quiet`이면 `statusf`/`statusln`과 진행률 표시줄(`newProgressBar`)이 출력을 생략합니다.
- 진행률 표시줄은 `progressBarVisible`로 quiet/비TTY를 검사해 CI 로그에 애니메이션이 남지 않게 합니다.
- 색상은 비TTY·`NO_COLOR`·`--no-color`에서 자동 비활성화됩니다(`resolveMarkdownStyle`, `os.Setenv("NO_COLOR")`).
- 표준 플래그 이름을 사용합니다: `--json`, `--quiet/-q`, `--no-color`, `--no-input`, `--version`, `--debug`.

### 페이저와 진행 피드백

- 사람이 읽는 긴 텍스트(`list`/`search` text)는 `writePaged`로 페이저에 넘깁니다. 페이징은 **stdout이 TTY일 때만** 하며 `PAGER`(기본 `less -FIRX`), `quiet`, `--no-pager`, JSON/YAML 출력에서는 하지 않습니다.
- 네트워크 호출 등 느릴 수 있는 작업은 `utils.WithSpinner`로 감싸 100ms 이내에 피드백을 줍니다. `WithSpinner`는 `spinnerDelay`(150ms) 후에만 표시해 빠른 작업의 깜짝임을 피하고, `Quiet`/비TTY에서는 아무것도 출력하지 않습니다.
- 스피너 코어는 `withSpinner(message, out, delay, fn)`로 분리되어 TTY 없이 단위 테스트합니다.

## TDD 방식 (중요)

주요 기능 추가/변경은 **테스트 주도 개발(TDD)** 방식으로 진행합니다:

1. **테스트 먼저 작성** — 검증할 동작을 `*_test.go`에 작성합니다.
2. **실패 확인** — 해당 기능이 없어 빌드/실패하는지 확인합니다.
3. **구현** — 테스트를 통과하는 최소 구현을 작성합니다.
4. **통과 확인** — 테스트가 모두 통과하는지 확인합니다.

테스트 원칙:

- **독립적**: 다른 테스트나 외부 서비스에 의존하지 않습니다.
- **고속**: 밀리초 단위로 끝나도록 단위 테스트 중심.
- 외부 의존성(HTTP, 홈 디렉토리, 환경변수)은 격리합니다.
  - HTTP는 `net/http/httptest` 사용 (예: `api/client_test.go`의 `newTestClient`).
  - 홈 디렉토리는 `t.TempDir()` + `t.Setenv("HOME", ...)`로 격리.
  - 환경변수는 `t.Setenv` 사용.
- 테이블 드리브 테이스(입력/기대치를 구조체 배열로)를 선호합니다.
- 커버리지가 낮은 부분(특히 `cmd`)은 커맨드의 순수 로직을 함수로 분리해 테스트로 커버합니다.

## 빌드/테스트 명령

```bash
# 테스트 (race + coverage)
go test -race -cover ./...

# vet (정적 분석)
go vet ./...

# 포맷
gofmt -l ./...   # 변경 필요 파일 확인
gofmt -w <file>  # 포맷 적용

# task 사용 시 (Taskfile.yml)
task test    # go test -v -race -cover ./...
task build   # ncli 바이너리 빌드
task fmt     # go fmt ./...
```

## 주의사항

- **`go mod tidy`로 go.mod/go.sum이 불필요하게 변경되는 것을 주의** — 작업에 필요한 변경이 아니면 되돌리세요 (`git checkout -- go.mod go.sum`).
- CI 워크플로우(`.github/workflows/release.yml`)는 릴리스 시 `go vet` + `go test -race`를 실행합니다. 릴리스 전에 로컬에서 동일하게 검증하세요.
- 커밋 메시지는 한국어로 간결하게, 변경 내용을 요약합니다.
