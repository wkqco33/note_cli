package cmd

import (
	"fmt"
	"note_cli/config/buildinfo"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "노트 CLI 버전 정보 출력",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Note CLI 버전: %s\n", buildinfo.Version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
