package database

import (
	"os"
	"strings"

	"github.com/Vedant9500/WTF/internal/errors"

	"gopkg.in/yaml.v3"
)

// LoadDatabase loads commands from a YAML file and returns a populated Database.
//
// This function reads a YAML file containing an array of Command objects,
// parses the YAML content, and populates performance-optimized cache fields
// for faster search operations. The cache fields include lowercased versions
// of all searchable text fields.
//
// Parameters:
//   - filename: Path to the YAML file containing command definitions
//
// Returns:
//   - *Database: Populated database with all commands and cache fields
//   - error: Database loading error with user-friendly context
//
// The function handles various error conditions:
//   - File not found errors
//   - Permission denied errors
//   - YAML parsing errors
//   - Invalid command structure errors
func LoadDatabase(filename string) (*Database, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.NewDatabaseErrorWithContext("read", filename, err)
	}

	var commands []Command
	if err := yaml.Unmarshal(data, &commands); err != nil {
		return nil, errors.NewDatabaseErrorWithContext("parse", filename, err)
	}

	// Populate lowercased cache fields for performance
	for i := range commands {
		commands[i].CommandLower = strings.ToLower(commands[i].Command)
		commands[i].DescriptionLower = strings.ToLower(commands[i].Description)
		commands[i].KeywordsLower = make([]string, len(commands[i].Keywords))
		for j, kw := range commands[i].Keywords {
			commands[i].KeywordsLower[j] = strings.ToLower(kw)
		}
		commands[i].TagsLower = make([]string, len(commands[i].Tags))
		for j, tag := range commands[i].Tags {
			commands[i].TagsLower[j] = strings.ToLower(tag)
		}
	}

	db := &Database{Commands: commands}
	// Build universal index for scalable search
	db.BuildUniversalIndex()
	// Build TF-IDF searcher and command index for hybrid NLP reranking
	db.buildTFIDFSearcher()
	return db, nil
}

// LoadDatabaseWithPersonal loads both main and personal database files and merges them.
//
// This function loads the main command database and optionally merges it with
// a personal database file containing user-specific commands. The personal
// database is optional - if it doesn't exist, only the main database is loaded.
//
// Parameters:
//   - mainDBPath: Path to the main command database file (required)
//   - personalDBPath: Path to the personal command database file (optional)
//
// Returns:
//   - *Database: Combined database with commands from both files
//   - error: Loading error if the main database fails to load
//
// Behavior:
//   - Main database loading failure results in an error
//   - Personal database not found is silently ignored
//   - Personal database parsing errors are reported
//   - Commands from personal database are appended to main database commands
func LoadDatabaseWithPersonal(mainDBPath, personalDBPath string) (*Database, error) {
	// Load main database
	mainDB, err := LoadDatabase(mainDBPath)
	if err != nil {
		return nil, err
	}

	// Try to load personal database (it's OK if it doesn't exist)
	personalDB, err := LoadDatabase(personalDBPath)
	if err != nil {
		// If personal database doesn't exist, that's fine - just use main database
		if os.IsNotExist(err) {
			return mainDB, nil
		}
		// Check if it's a DatabaseError wrapping IsNotExist
		if dbErr, ok := err.(*errors.DatabaseError); ok {
			if os.IsNotExist(dbErr.Cause) {
				return mainDB, nil
			}
		}
		// Check if it's an AppError wrapping IsNotExist
		if appErr, ok := err.(*errors.AppError); ok {
			if appErr.Cause != nil && os.IsNotExist(appErr.Cause) {
				return mainDB, nil
			}
		}
		// Other errors should be reported
		return nil, err
	}

	// Merge commands from both databases
	allCommands := make([]Command, 0, len(mainDB.Commands)+len(personalDB.Commands))
	allCommands = append(allCommands, mainDB.Commands...)
	allCommands = append(allCommands, personalDB.Commands...)

	db := &Database{Commands: allCommands}
	// Build universal index for scalable search
	db.BuildUniversalIndex()
	// Build TF-IDF searcher and command index for hybrid NLP reranking
	db.buildTFIDFSearcher()
	return db, nil
}

// Size returns the total number of commands in the database.
//
// This method provides a quick way to get the count of all loaded commands,
// which is useful for metrics, debugging, and capacity planning.
func (db *Database) Size() int {
	return len(db.Commands)
}
