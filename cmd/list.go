package cmd

import (
	"fmt"
	"os"

	"github.com/mdjarv/todoist-cli/internal/auth"
	"github.com/mdjarv/todoist-cli/internal/tasks"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := auth.LoadCredentials()
		if err != nil {
			fmt.Printf("Error loading credentials: %v\n", err)
			os.Exit(1)
		}
		// Get first page with default limit (50)
		tasks, err := tasks.ListTasks(creds.AccessToken, nil)
		if err != nil {
			fmt.Printf("Error listing tasks: %v\n", err)
			os.Exit(1)
		}

		for _, task := range tasks.Results {
			fmt.Printf("Task ID: %s\n", task.ID)
			fmt.Printf("Content: %s\n", task.Content)
			fmt.Printf("Project ID: %s\n", task.ProjectID)
			fmt.Printf("Section ID: %s\n", task.SectionID)
			fmt.Printf("Priority: %d\n", task.Priority)
			fmt.Printf("Due Date: %v\n", task.Due)
			fmt.Println("-----------------------------")
		}
		return nil

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
