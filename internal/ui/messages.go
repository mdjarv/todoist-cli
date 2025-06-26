package ui

import "github.com/mdjarv/todoist-cli/internal/todoist"

// ReloadMsg is sent after fetching new data
type ReloadMsg struct {
	Tasks        []todoist.Task
	ProjectNames map[string]string
}

type toggleDoneMsg struct {
	taskID string
}

type taskUpdatedMsg struct {
	taskID string
}

type createTaskMsg struct {
	options todoist.CreateTaskOptions
}
