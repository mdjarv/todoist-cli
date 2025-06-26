package cmd

import (
	"fmt"
	"os"

	"github.com/mdjarv/todoist-cli/internal/auth"
	"github.com/mdjarv/todoist-cli/internal/todoist"
	"github.com/mdjarv/todoist-cli/internal/ui"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "todoist",
	Short: "Todoist CLI",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := auth.LoadCredentials()
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to load credentials, please authenticate first")
			os.Exit(1)
		}

		client := todoist.NewClient(creds.AccessToken)
		return ui.Run(client)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
