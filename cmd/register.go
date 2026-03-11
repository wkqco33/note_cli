package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/config"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "새 사용자 계정 등록",
	Run: func(cmd *cobra.Command, args []string) {
		var name, email, password string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Name").
					Value(&name).
					Validate(func(str string) error {
						if str == "" {
							return fmt.Errorf("name is required")
						}
						return nil
					}),
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
			fmt.Println("Registration cancelled.")
			return
		}

		cfg, _ := config.Load()
		client := api.NewClient(cfg)
		
		user := api.UserCreate{
			Name:     name,
			Email:    email,
			Password: password,
		}
		
		err = client.Register(user)
		if err != nil {
			fmt.Printf("Registration failed: %v\n", err)
			return
		}

		fmt.Println("Successfully registered! You can now run 'note login'.")
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
