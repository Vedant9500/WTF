package database

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadDatabase loads commands from a YAML file
func LoadDatabase(filename string) (*Database, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read database file: %w", err)
	}

	var commands []Command
	if err := yaml.Unmarshal(data, &commands); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filename, err)
	}

	return &Database{
		Commands: commands,
	}, nil
}

// Size returns the number of commands in the database
func (db *Database) Size() int {
	return len(db.Commands)
}