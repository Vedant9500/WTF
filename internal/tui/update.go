package tui

import (
	"github.com/Vedant9500/WTF/internal/database"
	tea "github.com/charmbracelet/bubbletea"
)

// performSearch is a command to run the search in background
func performSearch(db *database.Database, query string) tea.Cmd {
	return func() tea.Msg {
		opts := database.SearchOptions{
			Limit:    20, // More results for TUI
			UseFuzzy: true,
			UseNLP:   true,
		}
		results := db.SearchUniversal(query, opts)
		return resultsMsg(results)
	}
}

type resultsMsg []database.SearchResult

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		switch m.state {
		case StateInput:
			switch msg.Type {
			case tea.KeyEnter:
				if m.query != "" {
					m.state = StateSearching
					return m, performSearch(m.db, m.query)
				}
			case tea.KeyEsc:
				return m, tea.Quit
			case tea.KeyBackspace:
				if len(m.query) > 0 {
					m.query = m.query[:len(m.query)-1]
				}
			case tea.KeyRunes:
				m.query += string(msg.Runes)
			case tea.KeySpace: // Handle space explicitly if not caught by Runes
				m.query += " "
			}

		case StateBrowsing:
			switch msg.String() {
			case "q", "esc":
				m.state = StateInput
				m.results = nil
				m.cursor = 0
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.results)-1 {
					m.cursor++
				}
			case "enter":
				// Could move to detail view, for now just toggle or something
				// m.state = StateDetail
			}
		}

	case resultsMsg:
		m.results = msg
		m.state = StateBrowsing
		m.cursor = 0

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}
