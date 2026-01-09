package tui

import (
	"github.com/Vedant9500/WTF/internal/database"
	tea "github.com/charmbracelet/bubbletea"
)

// AppState represents the current state of the TUI
type AppState int

const (
	StateInput AppState = iota
	StateSearching
	StateBrowsing
	StateDetail
	StateError
)

// Model holds the application state
type Model struct {
	state          AppState
	query          string
	results        []database.SearchResult
	cursor         int
	viewportOffset int
	err            error
	width          int
	height         int
	db             *database.Database
	loadingMsg     string
}

// NewModel creates a new TUI model
func NewModel(db *database.Database, initialQuery string) Model {
	m := Model{
		state:  StateInput,
		query:  initialQuery,
		db:     db,
		cursor: 0,
	}

	if initialQuery != "" {
		m.state = StateSearching
	}

	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, tea.EnterAltScreen)
	
	if m.query != "" {
		cmds = append(cmds, performSearch(m.db, m.query))
	}
	
	return tea.Batch(cmds...)
}
