package database

// Command represents a single command entry in the database
type Command struct {
	Command     string   `yaml:"command"`
	Description string   `yaml:"description"`
	Keywords    []string `yaml:"keywords"`
	Niche       string   `yaml:"niche,omitempty"`
	Platform    []string `yaml:"platform,omitempty"`
	Pipeline    bool     `yaml:"pipeline"`

	// Cached lowercased fields for performance
	CommandLower     string   `yaml:"-"`
	DescriptionLower string   `yaml:"-"`
	KeywordsLower    []string `yaml:"-"`
}

// Database holds all commands and provides search functionality
type Database struct {
	Commands []Command `yaml:"-"`
}
