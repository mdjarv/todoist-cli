package ui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdjarv/todoist-cli/internal/todoist"
)

// Run starts the Bubble Tea program, accepting a todoist.Client for fetching tasks
func Run(client todoist.Client) error {
	ctx := context.Background()

	projectsResp, err := client.ListProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch projects: %w", err)
	}

	tasksResp, err := client.ListTasks(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch tasks: %w", err)
	}

	// Build projectID to name map
	projectNames := make(map[string]string)
	for _, p := range projectsResp.Results {
		projectNames[p.ID] = p.Name
	}

	m := NewModel(tasksResp.Results, projectNames, client)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to start UI: %w", err)
	}
	return nil
}