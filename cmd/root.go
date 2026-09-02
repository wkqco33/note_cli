package cmd

import (
	"fmt"
	"os"
	"strings"

	"note_cli/utils"

	"github.com/wkqco33/wcli"
)

var debugMode bool

var rootCmd = &wcli.Command{
	Use:   "note_cli",
	Short: "간단한 CLI 노트 애플리케이션",
	Long:  `Note CLI는 원격 API를 사용하는 빠르고 간단한 터미널 기반 노트 에디터입니다.`,
	// 에러는 Execute()에서 출력하므로 라이브러리의 중복 인쇄를 억제
	SilenceErrors: true,
	PersistentPreRun: func(ctx *wcli.Context) error {
		utils.DebugMode = debugMode
		utils.SetupLogger()
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
		os.Exit(1)
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
		if arg != "--debug" && !strings.HasPrefix(arg, "--debug=") {
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
}
