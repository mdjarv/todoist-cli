package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mdjarv/todoist-cli/internal/todoist"
)

// Model is the Bubble Tea model for the UI
// Extend this struct to support different modes/screens.
type Model struct {
	// Main table component from Bubble Tea
	Table table.Model
	// Mode could be "tasks", "projects", etc.
	mode string
	// Client for fetching data
	Client todoist.Client
	// Map project IDs to names
	ProjectNames map[string]string
	// Spinner for loading indication
	Spinner spinner.Model
	// Loading state
	Loading bool
	// All tasks from the API
	allTasks []todoist.Task
	// showDone toggles the filter for completed tasks
	showDone bool
	// updating is a map of task IDs to a boolean indicating if the task is being updated
	updating map[string]bool

	// New task input fields
	taskInput textinput.Model
	taskInputQuitting bool
}

// NewModel creates a new UI model initialized with default mode and table data
func NewModel(tasks []todoist.Task, projectNames map[string]string, client todoist.Client) Model {
	// Define table columns
	columns := []table.Column{
		{Title: "Done", Width: 5},
		{Title: "Task", Width: 40},
		{Title: "Project", Width: 20},
		{Title: "Due", Width: 12},
		{Title: "Labels", Width: 20},
	}

	// Sort tasks by due date
	sortTasks(tasks)

	// Filter out done tasks by default
	filteredTasks := []todoist.Task{}
	for _, task := range tasks {
		if !task.Checked {
			filteredTasks = append(filteredTasks, task)
		}
	}

	// Convert tasks to table rows
	rows := make([]table.Row, len(filteredTasks))
	for i, task := range filteredTasks {
		rows[i] = taskToRow(task, projectNames, false, spinner.Model{}) // No spinner initially
	}

	// Create table with styling
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	// Apply custom styling
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	// Initialize new task input
	ti := textinput.New()
	ti.Placeholder = "Task content"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return Model{
		mode:              "tasks",
		Table:             t,
		Client:            client,
		ProjectNames:      projectNames,
		Spinner:           sp,
		Loading:           false,
		allTasks:          tasks,
		showDone:          false,
		updating:          make(map[string]bool),
		taskInput:         ti,
		taskInputQuitting: false,
	}
}

// Init is called when the program starts
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.Spinner.Tick, textinput.Blink)
}

