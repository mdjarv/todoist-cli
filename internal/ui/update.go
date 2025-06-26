package ui

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mdjarv/todoist-cli/internal/todoist"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.mode == "new-task" {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				if content := m.taskInput.Value(); content != "" {
					options := todoist.CreateTaskOptions{Content: content}
					return m, func() tea.Msg { return createTaskMsg{options: options} }
				}
				// If empty, just do nothing
				return m, nil
			case tea.KeyEsc:
				m.mode = "tasks"
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.taskInput, cmd = m.taskInput.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Table.SetHeight(msg.Height - 4)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "a":
			m.mode = "new-task"
			return m, nil
		case "enter":
			selectedTask := m.allTasks[m.Table.Cursor()]
			m.updating[selectedTask.ID] = true
			return m, func() tea.Msg {
				return toggleDoneMsg{taskID: selectedTask.ID}
			}
		case "f":
			m.showDone = !m.showDone
			var filteredTasks []todoist.Task
			for _, task := range m.allTasks {
				if m.showDone || !task.Checked {
					filteredTasks = append(filteredTasks, task)
				}
			}
			rows := make([]table.Row, len(filteredTasks))
			for i, task := range filteredTasks {
				rows[i] = taskToRow(task, m.ProjectNames, m.updating[task.ID], m.Spinner)
			}
			m.Table.SetRows(rows)
			return m, nil
		case "r":
			// Prevent parallel reloads
			if m.Loading {
				return m, nil
			}
			m.Loading = true
			return m, tea.Batch(m.Spinner.Tick, reloadCmd(m.Client))
		}
	case ReloadMsg:
		m.Loading = false
		m.allTasks = msg.Tasks
		m.ProjectNames = msg.ProjectNames

		sortTasks(m.allTasks)

		var filteredTasks []todoist.Task
		for _, task := range m.allTasks {
			if m.showDone || !task.Checked {
				filteredTasks = append(filteredTasks, task)
			}
		}

		rows := make([]table.Row, len(filteredTasks))
		for i, task := range filteredTasks {
			rows[i] = taskToRow(task, msg.ProjectNames, m.updating[task.ID], m.Spinner)
		}
		m.Table.SetRows(rows)
	case toggleDoneMsg:
		return m, func() tea.Msg {
			var taskToUpdate todoist.Task
			for _, task := range m.allTasks {
				if task.ID == msg.taskID {
					taskToUpdate = task
					break
				}
			}

			var err error
			if taskToUpdate.Checked {
				err = m.Client.ReopenTask(context.Background(), msg.taskID)
			} else {
				err = m.Client.CloseTask(context.Background(), msg.taskID)
			}

			if err != nil {
				// Handle error, maybe show a message to the user
				return nil
			}

			return taskUpdatedMsg{taskID: msg.taskID}
		}
	case taskUpdatedMsg:
		m.updating[msg.taskID] = false
		for i, task := range m.allTasks {
			if task.ID == msg.taskID {
				m.allTasks[i].Checked = !m.allTasks[i].Checked
				break
			}
		}

		var filteredTasks []todoist.Task
		for _, task := range m.allTasks {
			if m.showDone || !task.Checked {
				filteredTasks = append(filteredTasks, task)
			}
		}

		rows := make([]table.Row, len(filteredTasks))
		for i, task := range filteredTasks {
			rows[i] = taskToRow(task, m.ProjectNames, m.updating[task.ID], m.Spinner)
		}
		m.Table.SetRows(rows)

		return m, nil
			case createTaskMsg:
		err := m.Client.CreateTask(context.Background(), msg.options)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error creating task:", err)
			return m, nil
		}
		m.taskInput.SetValue("") // Clear the input after creating task
		m.mode = "tasks" // Switch back to tasks view
		return m, reloadCmd(m.Client)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.Spinner, cmd = m.Spinner.Update(msg)
		// This is where we'll update the rows that are currently updating
		rows := m.Table.Rows()
		for i, row := range rows {
			// This is a bit of a hack, but we can get the task ID from the row
			// This is not ideal, but it works for now
			// A better solution would be to have a direct mapping from row to task ID
			// but that would require more significant changes
			// For now, we'll just find the task by content
			for _, task := range m.allTasks {
				if task.Content == row[1] && m.updating[task.ID] {
					rows[i] = taskToRow(task, m.ProjectNames, true, m.Spinner)
				}
			}
		}
		m.Table.SetRows(rows)
		return m, cmd
	}
	// Always delegate update to the Bubble Tea table so navigation and selection work
	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

