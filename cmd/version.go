package cmd

import (
	"fmt"
	"note_cli/config/buildinfo"

	"github.com/wkqco33/wcli"
)

var versionCmd = &wcli.Command{
	Use:   "version",
	Short: "노트 CLI 버전 정보 출력",
	Run: func(ctx *wcli.Context) error {
		fmt.Printf("Note CLI 버전: %s\n", buildinfo.Version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
