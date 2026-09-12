package cmd

import (
	"github.com/wkqco33/wcli"
)

// helpCmd "ncli help [command]"을 지원한다.
// wcli는 `ncli <command> --help`만 제공하므로, 서브커맨드 도움말도 조회할 수 있게 보완한다.
var helpCmd = &wcli.Command{
	Use:   "help [command]",
	Short: "도움말 표시",
	Long: `도움말을 표시합니다.

인자가 없으면 전체 도움말을, 커맨드를 지정하면 해당 커맨드의 도움말을 출력합니다.

예시:
  ncli help
  ncli help add
  ncli help ai create`,
	Run: func(ctx *wcli.Context) error {
		// 대상 커맨드의 도움말을 재실행으로 출력한다 (--help는 ErrHelp로 정상 종료 처리됨).
		args := make([]string, 0, len(ctx.Args)+1)
		args = append(args, ctx.Args...)
		args = append(args, "--help")

		return rootCmd.Execute(args)
	},
}

func init() {
	rootCmd.AddCommand(helpCmd)
}
