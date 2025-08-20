// Package database provides command database management and search functionality.
//
// This package implements the core data structures and search algorithms for the WTF
// command discovery tool. It includes:
//   - Command data models with performance optimizations
//   - Advanced search algorithms with scoring and ranking
//   - Context-aware search with platform filtering
//   - Fuzzy search capabilities for typo tolerance
//   - Natural language processing integration
//
// The Database type is the main entry point for all search operations.
package database

// Command represents a single command entry in the database with metadata and
// performance optimizations for search operations.
//
// The Command struct includes both the original YAML fields and cached lowercased
// versions for improved search performance. The cached fields are automatically
// populated during database loading and should not be set manually.
type Command struct {
	// Command is the actual shell command or command template
	Command string `yaml:"command"`

	// Description provides a human-readable explanation of what the command does
	Description string `yaml:"description"`

	// Keywords are searchable terms that help users find this command
	Keywords []string `yaml:"keywords"`

	// Tags provide additional categorization for the command
	Tags []string `yaml:"tags,omitempty"`

	// Niche specifies the domain or specialty area (e.g., "git", "docker", "system")
	Niche string `yaml:"niche,omitempty"`

	// Platform specifies which operating systems support this command
	// Common values: "linux", "macos", "windows", "cross-platform"
	Platform []string `yaml:"platform,omitempty"`

	// Pipeline indicates whether this command is commonly used in pipelines
	Pipeline bool `yaml:"pipeline"`

	// Cached lowercased fields for performance optimization during search
	// These fields are automatically populated and should not be set manually
	CommandLower     string   `yaml:"-"`
	DescriptionLower string   `yaml:"-"`
	KeywordsLower    []string `yaml:"-"`
	TagsLower        []string `yaml:"-"`
}

// Database holds all commands and provides comprehensive search functionality.
//
// The Database struct is the main interface for command storage and retrieval.
// It supports various search modes including:
//   - Basic keyword search
//   - Context-aware search with platform filtering
//   - Fuzzy search for typo tolerance
//   - Natural language processing
//   - Pipeline-specific search
//
// Example usage:
//
//	db := &Database{}
//	err := db.LoadFromFile("commands.yml")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	results := db.Search("compress files", 5)
//	for _, result := range results {
//		fmt.Printf("%s: %s\n", result.Command.Command, result.Command.Description)
//	}
type Database struct {
	// Commands contains all loaded command entries
	Commands []Command `yaml:"-"`
	// uIndex is the optional universal inverted index for scalable search
	uIndex *universalIndex `yaml:"-"`
}
