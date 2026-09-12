# CLI 가이드라인 준수

`ncli`는 [Command Line Interface Guidelines](https://clig.dev/)를 참고해 설계했습니다. 이 문서는 ncli가 지키는 규칙과 새 커맨드·플래그 추가 시의 점검 체크리스트를 정리합니다.

## 핵심 원칙과 구현

| 항목 | 규칙 | ncli 구현 |
| --- | --- | --- |
| 입력 | 프롬프트는 `stdin`이 TTY일 때만 표시하고, 모든 입력에 플래그/인자 대체 경로를 제공합니다. | `promptAllowed`, `requirePrompt`, `ErrInteractionRequired` |
| 확인 | 파괴적 작업은 확인을 거치되 비대화형에서 `--yes`로 실행할 수 있어야 합니다. | `askConfirm`, `--yes/-y` |
| 출력 | 결과는 stdout, 진행·상태·오류는 stderr로 분리합니다. | `statusf`/`statusln`, `printRenderedNotes` |
| 기계 판독 | `--json`/`--format` 같은 구조화 출력을 제공합니다. | `--format json/yaml`, `--json` |
| 종료 코드 | 성공 `0`, 오류 `1`, 입력 필요 `2`로 구분합니다. | `exitCodeError`, `exitCodeActionRequired` |
| 표준 플래그 | 관용적인 이름을 사용합니다. | `--help/-h`, `--version`, `--quiet/-q`, `--no-color`, `--no-input`, `--debug`, `--no-pager`, `--dry-run` |
| 터미널 표현 | 비TTY·`NO_COLOR`·`TERM=dumb`에서 색상·애니메이션을 끄고, 긴 사람용 출력은 TTY에서만 페이저로 넘깁니다. | `resolveMarkdownStyle`, `progressBarVisible`, `writePaged`, `utils.WithSpinner` |
| 설정 | XDG 규격을 따릅니다. | `XDG_CONFIG_HOME`, `XDG_DATA_HOME` |
| 시크릿 | 플래그로 직접 받지 않고 파일/stdin을 사용합니다. | `--password-file`, `--password-stdin` |
| 도움말 | 전체 경로, 예시, 문서/이슈 링크를 포함합니다. | `ncli add [flags]`, 각 커맨드 `Long` |
| 네트워크 | 요청 전 100ms 이내에 피드백을 주고 타임아웃을 둡니다. | `utils.WithSpinner`, HTTP/LLM 타임아웃 |
| 하위 호환 | 변경은 additive하게 유지하고, 불가피하면 deprecation 경고를 제공합니다. | — |

## 점검 체크리스트

새 커맨드나 플래그를 추가할 때 다음을 확인합니다.

- [ ] 프롬프트는 `stdin`이 TTY일 때만 표시하는가?
- [ ] 프롬프트로 받는 값에 플래그/인자 대체 경로가 있는가?
- [ ] 비TTY·`--no-input`에서 조용히 취소되지 않고 0이 아닌 코드로 실패하는가?
- [ ] 파괴적 작업에 확인이 있고, `--yes`로 비대화형 실행이 가능한가?
- [ ] 결과는 stdout, 진행·상태·오류는 stderr로 분리했는가?
- [ ] 스크립트용 `--json` 또는 `--format` 출력을 제공하는가?
- [ ] 종료 코드가 성공 `0` / 오류 `1` / 입력 필요 `2`를 따르는가?
- [ ] 표준 플래그 이름을 사용했는가?
- [ ] 비TTY·`NO_COLOR`에서 색상·애니메이션을 끄는가?
- [ ] 긴 사람용 출력은 TTY에서만 페이저로 넘기는가?
- [ ] 설정·시크릿이 XDG 규격과 파일/stdin 입력을 따르는가?
- [ ] 도움말에 전체 경로·예시·문서/이슈 링크가 있는가?
- [ ] 네트워크 호출에 피드백과 타임아웃이 있는가?

## 코드 수준 규칙

프롬프트와 출력 처리는 `cmd/interactive.go`, `cmd/output.go`, `cmd/pager.go`의 헬퍼를 사용합니다. 구체적인 구현 규칙은 [개발 가이드](../AGENTS.md)의 "대화형 입력(TUI) 규칙", "출력 스트림 규칙", "페이저와 진행 피드백"을 참고하세요.
