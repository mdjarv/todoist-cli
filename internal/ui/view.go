package ui

// View renders the UI
func (m Model) View() string {
	var s string
	var help string

	if m.Loading {
		help = m.Spinner.View() + " Loading... " + "Press q to quit, f to toggle completed tasks, enter to toggle done, a to add a new task."
	} else {
		help = "Press q to quit, f to toggle completed tasks, enter to toggle done, a to add a new task."
	}

	s = m.Table.View() + "\n\n" + help

	if m.mode == "new-task" {
		s = s + "\n\n" + m.taskInput.View()
	}

	return s
}

