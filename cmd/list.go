package cmd

import (
	"fmt"
	"github.com/mdjarv/todoist-cli/internal/auth"
	"github.com/mdjarv/todoist-cli/internal/tasks"
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
			return fmt.Errorf("failed to load credentials, please login first")
		}

		client := todoist.NewClient(creds.AccessToken)
		return tasks.ListTasks(cmd.Context(), client)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
