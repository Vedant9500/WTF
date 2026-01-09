package cli

import (
	"fmt"
	"os"

	"github.com/Vedant9500/WTF/internal/config"
	"github.com/Vedant9500/WTF/internal/recovery"
	"github.com/Vedant9500/WTF/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive Terminal User Interface",
	Long:  `Start WTF in interactive mode with a rich Terminal User Interface (TUI).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg := config.DefaultConfig()
		
		// Load database
		dbFilePath := cfg.GetDatabasePath()
		personalDBPath := cfg.GetPersonalDatabasePath()
		
		dbRecovery := recovery.NewDatabaseRecovery(recovery.DefaultRetryConfig())
		db, err := dbRecovery.LoadDatabaseWithFallback(dbFilePath, personalDBPath)
		if err != nil {
			fmt.Printf("Error loading database: %v\n", err)
			os.Exit(1)
		}

		// Initial query if provided
		initialQuery := ""
		if len(args) > 0 {
			initialQuery = args[0]
		}

		// Initialize TUI model
		model := tui.NewModel(db, initialQuery)
		// TODO: Set initial query if implemented in model

		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error starting TUI: %v\n", err)
			os.Exit(1)
		}
	},
}
