package database

// Command represents a single command entry in the database
type Command struct {
	Command     string   `yaml:"command"`
	Description string   `yaml:"description"`
	Keywords    []string `yaml:"keywords"`
	Niche       string   `yaml:"niche,omitempty"`
	Platform    []string `yaml:"platform,omitempty"`
	Pipeline    bool     `yaml:"pipeline"`
}

// Database holds all commands and provides search functionality
type Database struct {
	Commands []Command `yaml:"-"`
}