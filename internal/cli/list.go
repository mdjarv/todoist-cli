package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mdjarv/todoist-cli/internal/todoist"
)

// List displays a simple CLI list of tasks
func List(ctx context.Context, client todoist.Client, jsonOut bool) error {
	// Fetch tasks from Todoist API
	tasksResp, err := client.ListTasks(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch tasks: %w", err)
	}

	if len(tasksResp.Results) == 0 {
		fmt.Println("No tasks found.")
		return nil
	}

	// Display tasks in JSON if requested
	if jsonOut {
		importedJson, err := json.MarshalIndent(tasksResp.Results, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal tasks to JSON: %w", err)
		}
		fmt.Println(string(importedJson))
		return nil
	}

	// Display header
	fmt.Println("ID\tContent\tProject")
	for _, task := range tasksResp.Results {
		fmt.Printf("%s\t%s\t%s\n", task.ID, task.Content, task.ProjectID)
	}
	return nil
}
