package database

import (
	"os"
	"strings"

	"github.com/Vedant9500/WTF/internal/errors"

	"gopkg.in/yaml.v3"
)

// LoadDatabase loads commands from a YAML file
func LoadDatabase(filename string) (*Database, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.NewDatabaseError("read", filename, err)
	}

	var commands []Command
	if err := yaml.Unmarshal(data, &commands); err != nil {
		return nil, errors.NewDatabaseError("parse", filename, err)
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

	return &Database{
		Commands: commands,
	}, nil
}

// LoadDatabaseWithPersonal loads both main and personal database files
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
		// Other errors should be reported
		return nil, err
	}

	// Merge commands from both databases
	allCommands := make([]Command, 0, len(mainDB.Commands)+len(personalDB.Commands))
	allCommands = append(allCommands, mainDB.Commands...)
	allCommands = append(allCommands, personalDB.Commands...)

	return &Database{
		Commands: allCommands,
	}, nil
}

// Size returns the number of commands in the database
func (db *Database) Size() int {
	return len(db.Commands)
}
