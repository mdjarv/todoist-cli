package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdjarv/todoist-cli/internal/todoist"
)

// reloadCmd fetches projects and tasks, then returns ReloadMsg
func reloadCmd(client todoist.Client) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		projectsResp, err := client.ListProjects(ctx)
		if err != nil {
			return err
		}
		tasksResp, err := client.ListTasks(ctx, nil)
		if err != nil {
			return err
		}
		projectNames := make(map[string]string)
		for _, p := range projectsResp.Results {
			projectNames[p.ID] = p.Name
		}
		return ReloadMsg{
			Tasks:        tasksResp.Results,
			ProjectNames: projectNames,
		}
	}
}
