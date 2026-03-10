package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/config"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to the Note API",
	Run: func(cmd *cobra.Command, args []string) {
		var email, password string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Email").
					Value(&email).
					Validate(func(str string) error {
						if str == "" {
							return fmt.Errorf("email is required")
						}
						return nil
					}),
				huh.NewInput().
					Title("Password").
					EchoMode(huh.EchoModePassword).
					Value(&password).
					Validate(func(str string) error {
						if str == "" {
							return fmt.Errorf("password is required")
						}
						return nil
					}),
			),
		)

		err := form.Run()
		if err != nil {
			fmt.Println("Login cancelled.")
			return
		}

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		client := api.NewClient(cfg)
		err = client.Login(email, password)
		if err != nil {
			fmt.Printf("Login failed: %v\n", err)
			return
		}

		fmt.Println("Successfully logged in!")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
