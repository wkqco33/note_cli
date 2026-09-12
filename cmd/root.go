package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"note_cli/config/buildinfo"
	"note_cli/utils"

	"github.com/wkqco33/wcli"
)

const (
	// exitCodeError 일반 오류 종료 코드
	exitCodeError = 1
	// exitCodeActionRequired 대화형 입력이 필요하지만 프롬프트를 띄울 수 없는 경우.
	// 에이전트/스크립트가 실제 실패와 "플래그를 더 지정해야 함"을 구분할 수 있게 한다.
	exitCodeActionRequired = 2
)

var (
	debugMode bool
	// assumeYes 확인 프롬프트를 자동 승인한다 (--yes/-y)
	assumeYes bool
	// noInput 모든 대화형 프롬프트를 금지한다 (--no-input)
	noInput bool
	// noColorMode 색상 출력을 강제로 끈다 (--no-color)
	noColorMode bool
	// noPagerMode 긴 출력의 페이저 사용을 끈다 (--no-pager)
	noPagerMode bool
)

// issueURL 버그/피드백 접수 경로 (도움말에 안내)
const issueURL = "https://github.com/wkqco33/note_cli/issues"

// rootGlobalFlags 서브커맨드보다 앞에 와도 허용하는 루트 전역 플래그 이름.
// wcli는 서브커맨드 라우팅을 먼저 하므로, 앞에 오면 뒤로 옮겨준다.
var rootGlobalFlags = map[string]bool{
	"--debug":    true,
	"--yes":      true,
	"-y":         true,
	"--no-input": true,
	"--quiet":    true,
	"-q":         true,
	"--no-color": true,
	"--no-pager": true,
}

// isRootGlobalFlag 이름이 루트 전역 플래그인지 판단한다 (--flag=value 형식 포함).
func isRootGlobalFlag(arg string) bool {
	name := arg
	if idx := strings.IndexByte(arg, '='); idx >= 0 {
		name = arg[:idx]
	}

	return rootGlobalFlags[name]
}

var rootCmd = &wcli.Command{
	Use:   "note_cli",
	Short: "간단한 CLI 노트 애플리케이션",
	Long: `Note CLI는 원격 API를 사용하는 빠르고 간단한 터미널 기반 노트 에디터입니다.

전역 플래그:
  --yes, -y     확인 프롬프트를 자동 승인합니다 (비대화형/CI용)
  --no-input    모든 대화형 프롬프트를 금지합니다 (필요한 값을 플래그로 지정)
  --quiet, -q   진행/상태 메시지를 출력하지 않습니다
  --no-color    색상 출력을 끕니다 (NO_COLOR 환경변수와 동일)
  --no-pager    긴 출력을 페이저로 넘기지 않습니다 (PAGER 환경변수로 지정)
  --debug       디버그 출력 활성화

대화형 프롬프트는 stdin이 터미널일 때만 표시됩니다. 파이프/CI/에이전트 환경에서는
프롬프트 대신 필요한 플래그를 안내하는 오류와 함께 종료 코드 2로 종료합니다.

문서: https://github.com/wkqco33/note_cli#readme
버그 신고: ` + issueURL + `
`,
	// 에러는 Execute()에서 출력하므로 라이브러리의 중복 인쇄를 억제
	SilenceErrors: true,
	PersistentPreRun: func(ctx *wcli.Context) error {
		// DEBUG 환경변수는 --debug와 동일하게 동작한다 (일반 목적 환경변수 관례).
		utils.DebugMode = debugMode || os.Getenv("DEBUG") != ""
		utils.Quiet = quietMode
		utils.SetupLogger()
		if noColorMode {
			// lipgloss/glamour는 NO_COLOR를 참조해 색상을 끈다.
			_ = os.Setenv("NO_COLOR", "1")
		}
		if utils.DebugMode {
			utils.Debugln("디버그 모드가 활성화되었습니다.")
		}
		return nil
	},
}

// Execute 설정된 모든 자식 명령을 실행하고 플래그 값을 설정
func Execute() {
	if err := rootCmd.Execute(normalizeRootArgs(os.Args[1:])); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, ErrInteractionRequired) {
			os.Exit(exitCodeActionRequired)
		}
		os.Exit(exitCodeError)
	}
}

// normalizeRootArgs는 wcli가 루트 플래그 뒤의 서브커맨드를 라우팅하도록 보정한다.
func normalizeRootArgs(args []string) []string {
	commandIndex := -1
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			commandIndex = i
			break
		}
	}
	if commandIndex <= 0 {
		return args
	}

	var rootFlags []string
	for _, arg := range args[:commandIndex] {
		if !isRootGlobalFlag(arg) {
			return args
		}
		rootFlags = append(rootFlags, arg)
	}
	if len(rootFlags) == 0 {
		return args
	}

	result := make([]string, 0, len(args))
	result = append(result, args[commandIndex:]...)
	result = append(result, rootFlags...)
	return result
}

func init() {
	// 도움말에 실제 실행 파일 이름이 표시되도록 조정 (예: ncli.exe → ncli)
	rootCmd.Use = binaryName()

	// 최상위 명령어 전역 플래그 정의
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", "", false, "디버그 출력 활성화")
	rootCmd.PersistentFlags().BoolVar(&assumeYes, "yes", "y", false, "확인 프롬프트를 자동 승인합니다 (비대화형/CI용)")
	rootCmd.PersistentFlags().BoolVar(&noInput, "no-input", "", false, "모든 대화형 프롬프트를 금지합니다 (필요 시 플래그로 입력)")
	rootCmd.PersistentFlags().BoolVar(&quietMode, "quiet", "q", false, "진행/상태 메시지를 출력하지 않습니다")
	rootCmd.PersistentFlags().BoolVar(&noColorMode, "no-color", "", false, "색상 출력을 끕니다")
	rootCmd.PersistentFlags().BoolVar(&noPagerMode, "no-pager", "", false, "긴 출력을 페이저로 넘기지 않습니다")

	// --version 표준 플래그 지원 (wcli가 Version 설정 시 자동 감지)
	rootCmd.Version = buildinfo.Version
}
