package cmd

import (
	"fmt"
	"os"

	"note_cli/utils"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "note_cli",
	Short: "간단한 CLI 노트 애플리케이션",
	Long:  `Note CLI는 원격 API를 사용하는 빠르고 간단한 터미널 기반 노트 에디터입니다.`,
	// 에러는 Execute()에서 출력하므로 Cobra의 중복 인쇄를 억제
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		debug, _ := cmd.Flags().GetBool("debug")
		utils.DebugMode = debug
		utils.SetupLogger()
		if utils.DebugMode {
			utils.Debugln("디버그 모드가 활성화되었습니다.")
		}
	},
}

// Execute 설정된 모든 자식 명령을 실행하고 플래그 값을 설정
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// 도움말에 실제 실행 파일 이름이 표시되도록 조정 (예: ncli.exe → ncli)
	rootCmd.Use = binaryName()

	// 최상위 명령어 전역 플래그 정의
	rootCmd.PersistentFlags().Bool("debug", false, "디버그 출력 활성화")
}
