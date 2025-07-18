package database

import (
	"os"

	"cmd-finder/internal/errors"
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

	return &Database{
		Commands: commands,
	}, nil
}

// Size returns the number of commands in the database
func (db *Database) Size() int {
	return len(db.Commands)
}