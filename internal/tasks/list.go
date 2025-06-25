package tasks

import (
	"context"
	"fmt"
	"github.com/mdjarv/todoist-cli/internal/todoist"
)

func ListTasks(ctx context.Context, client todoist.Client) error {
	tasksResponse, err := client.ListTasks(ctx, nil)
	if err != nil {
		return err
	}

	for _, task := range tasksResponse.Results {
		fmt.Printf("ID: %v | Content: %v | Due: %v\n", task.ID, task.Content, task.Due)
	}
	return nil
}
