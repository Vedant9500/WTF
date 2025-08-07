package testutil

import (
	"strings"

	"github.com/Vedant9500/WTF/internal/database"
)

// CommandBuilder provides a fluent interface for building test commands
type CommandBuilder struct {
	command     string
	description string
	keywords    []string
	platform    []string
	pipeline    bool
	niche       string
}

// NewCommandBuilder creates a new command builder
func NewCommandBuilder() *CommandBuilder {
	return &CommandBuilder{
		keywords: make([]string, 0),
		platform: make([]string, 0),
	}
}

// Command sets the command string
func (cb *CommandBuilder) Command(cmd string) *CommandBuilder {
	cb.command = cmd
	return cb
}

// Description sets the description
func (cb *CommandBuilder) Description(desc string) *CommandBuilder {
	cb.description = desc
	return cb
}

// Keywords sets the keywords
func (cb *CommandBuilder) Keywords(keywords ...string) *CommandBuilder {
	cb.keywords = keywords
	return cb
}

// Platform sets the supported platforms
func (cb *CommandBuilder) Platform(platforms ...string) *CommandBuilder {
	cb.platform = platforms
	return cb
}

// Pipeline marks the command as a pipeline command
func (cb *CommandBuilder) Pipeline(isPipeline bool) *CommandBuilder {
	cb.pipeline = isPipeline
	return cb
}

// Niche sets the command niche
func (cb *CommandBuilder) Niche(niche string) *CommandBuilder {
	cb.niche = niche
	return cb
}

// Build creates the database.Command
func (cb *CommandBuilder) Build() database.Command {
	cmd := database.Command{
		Command:     cb.command,
		Description: cb.description,
		Keywords:    cb.keywords,
		Platform:    cb.platform,
		Pipeline:    cb.pipeline,
		Niche:       cb.niche,
	}

	// Populate cached fields
	cmd.CommandLower = strings.ToLower(cmd.Command)
	cmd.DescriptionLower = strings.ToLower(cmd.Description)
	cmd.KeywordsLower = make([]string, len(cmd.Keywords))
	for i, keyword := range cmd.Keywords {
		cmd.KeywordsLower[i] = strings.ToLower(keyword)
	}

	return cmd
}

// CommonPlatforms provides common platform combinations
var (
	AllPlatforms  = []string{"linux", "macos", "windows"}
	UnixPlatforms = []string{"linux", "macos"}
	WindowsOnly   = []string{"windows"}
	LinuxOnly     = []string{"linux"}
	MacOSOnly     = []string{"macos"}
)

// Predefined command builders for common scenarios
func GitCommand(subcommand, description string) *CommandBuilder {
	return NewCommandBuilder().
		Command("git "+subcommand).
		Description(description).
		Keywords("git", "version-control", subcommand).
		Platform(AllPlatforms...).
		Niche("development")
}

func FileCommand(cmd, description string) *CommandBuilder {
	return NewCommandBuilder().
		Command(cmd).
		Description(description).
		Keywords("file", "filesystem").
		Platform(UnixPlatforms...).
		Niche("system")
}

func NetworkCommand(cmd, description string) *CommandBuilder {
	return NewCommandBuilder().
		Command(cmd).
		Description(description).
		Keywords("network", "remote").
		Platform(AllPlatforms...).
		Niche("network")
}

func PipelineCommand(cmd, description string) *CommandBuilder {
	return NewCommandBuilder().
		Command(cmd).
		Description(description).
		Keywords("pipeline", "chain").
		Platform(UnixPlatforms...).
		Pipeline(true).
		Niche("system")
}
