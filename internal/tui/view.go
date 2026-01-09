package tui

import (
	"fmt"
)

func (m Model) View() string {
	var s string

	switch m.state {
	case StateInput:
		s += "🔍 What's The Function (WTF) - TUI Mode\n\n"
		s += "Enter your query:\n"
		s += "> " + m.query + "█\n\n" // Cursor simulation
		s += "(Press Enter to search, Esc to quit)"

	case StateSearching:
		s += "🔍 Finding the best commands for you...\n"

	case StateBrowsing:
		s += fmt.Sprintf("Found %d results for '%s' (Press q to search again):\n\n", len(m.results), m.query)

		// Calculate viewport
		start := 0
		end := len(m.results)
		if end > m.height-5 {
			end = m.height - 5
		}

		for i := start; i < end; i++ {
			cursor := " " // no cursor
			if m.cursor == i {
				cursor = ">" // cursor
			}

			res := m.results[i]
			title := res.Command.Command
			desc := res.Command.Description
			
			// Highlight selected
			if m.cursor == i {
				title = fmt.Sprintf("\033[1m%s\033[0m", title) // Bold
				s += fmt.Sprintf("%s %s\n   \033[36m%s\033[0m\n", cursor, title, desc)
			} else {
				s += fmt.Sprintf("%s %s\n   %s\n", cursor, title, desc)
			}
			s += "\n"
		}
		
		s += "\n(Use arrow keys to navigate, Enter to select, q to back)"

	case StateError:
		s += fmt.Sprintf("Error: %v\n\nPress q to try again.", m.err)
	}

	return s
}
