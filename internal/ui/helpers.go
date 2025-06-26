package ui

import (
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/mdjarv/todoist-cli/internal/todoist"
)

// taskToRow converts a todoist.Task to a table row
func taskToRow(task todoist.Task, projectNames map[string]string, isUpdating bool, spin spinner.Model) table.Row {
	done := " "
	if isUpdating {
		done = spin.View()
	} else if task.Checked {
		done = "✓"
	}
	// Use project name if available, fallback to ID or Unknown
	project := "Unknown"
	if task.ProjectID != "" {
		if name, ok := projectNames[task.ProjectID]; ok && name != "" {
			project = name
		} else {
			project = task.ProjectID
		}
	}

	// Format due date
	due := "No due date"
	if task.Due != nil {
		if dateStr, ok := task.Due["date"].(string); ok && dateStr != "" {
			due = dateStr
		}
	}

	// Format labels
	labels := strings.Join(task.Labels, ", ")
	if labels == "" {
		labels = "No labels"
	}

	return table.Row{
		done,
		task.Content,
		project,
		due,
		labels,
	}
}

func sortTasks(tasks []todoist.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		dueI := time.Now().AddDate(100, 0, 0) // A date far in the future
		if tasks[i].Due != nil {
			if dateStr, ok := tasks[i].Due["date"].(string); ok {
				parsed, err := time.Parse("2006-01-02", dateStr)
				if err == nil {
					dueI = parsed
				}
			}
		}

		dueJ := time.Now().AddDate(100, 0, 0) // A date far in the future
		if tasks[j].Due != nil {
			if dateStr, ok := tasks[j].Due["date"].(string); ok {
				parsed, err := time.Parse("2006-01-02", dateStr)
				if err == nil {
					dueJ = parsed
				}
			}
		}

		return dueI.Before(dueJ)
	})
}
