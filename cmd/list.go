package cmd

import (
	"fmt"
	"os"

	"github.com/mdjarv/todoist-cli/internal/auth"
	"github.com/mdjarv/todoist-cli/internal/cli"
	"github.com/mdjarv/todoist-cli/internal/todoist"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := auth.LoadCredentials()
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to load credentials, please authenticate first")
			os.Exit(1)
		}

			jsonOut, _ := cmd.Flags().GetBool("json")
			client := todoist.NewClient(creds.AccessToken)
			return cli.List(cmd.Context(), client, jsonOut)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Add --json flag for outputting tasks as JSON
	listCmd.Flags().Bool("json", false, "Output tasks as JSON")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
