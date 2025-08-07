package testutil

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Vedant9500/WTF/internal/database"
)

// TestDataGenerator provides methods to generate test data for various scenarios
type TestDataGenerator struct {
	rand *rand.Rand
}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewTestDataGeneratorWithSeed creates a new test data generator with a specific seed
func NewTestDataGeneratorWithSeed(seed int64) *TestDataGenerator {
	return &TestDataGenerator{
		rand: rand.New(rand.NewSource(seed)),
	}
}

// GenerateRandomCommands generates random commands for stress testing
func (tdg *TestDataGenerator) GenerateRandomCommands(count int) []database.Command {
	commands := make([]database.Command, count)

	commandPrefixes := []string{
		"git", "find", "grep", "tar", "zip", "curl", "wget", "ssh", "scp", "rsync",
		"ls", "cp", "mv", "rm", "mkdir", "rmdir", "cat", "less", "more", "head",
		"tail", "sort", "uniq", "awk", "sed", "ps", "top", "htop", "kill", "chmod",
		"chown", "sudo", "ping", "nc", "nmap", "docker", "kubectl", "helm",
	}

	commandSuffixes := []string{
		"-l", "-a", "-r", "-f", "-v", "-h", "--help", "--version", "-i", "-o",
		"file.txt", "directory/", "*.log", "pattern", "user@host", "localhost",
		"--recursive", "--force", "--verbose", "--quiet", "--dry-run",
	}

	descriptions := []string{
		"manage files and directories", "search and filter content", "network operations",
		"version control operations", "system monitoring", "process management",
		"archive and compression", "text processing", "remote access", "configuration",
		"development tools", "container management", "security operations",
	}

	keywords := []string{
		"file", "directory", "search", "network", "git", "process", "system",
		"archive", "text", "remote", "config", "development", "container",
		"security", "monitor", "manage", "create", "delete", "copy", "move",
		"list", "show", "edit", "compress", "extract", "download", "upload",
	}

	platforms := [][]string{
		{"linux", "macos"},
		{"linux", "macos", "windows"},
		{"windows"},
		{"linux"},
		{"macos"},
	}

	niches := []string{
		"development", "system", "network", "security", "devops", "database",
		"web", "mobile", "desktop", "server", "cloud", "monitoring",
	}

	for i := 0; i < count; i++ {
		prefix := commandPrefixes[tdg.rand.Intn(len(commandPrefixes))]
		suffix := commandSuffixes[tdg.rand.Intn(len(commandSuffixes))]
		command := fmt.Sprintf("%s %s", prefix, suffix)

		description := descriptions[tdg.rand.Intn(len(descriptions))]

		// Generate 2-5 keywords
		numKeywords := 2 + tdg.rand.Intn(4)
		cmdKeywords := make([]string, numKeywords)
		for j := 0; j < numKeywords; j++ {
			cmdKeywords[j] = keywords[tdg.rand.Intn(len(keywords))]
		}

		platform := platforms[tdg.rand.Intn(len(platforms))]
		niche := niches[tdg.rand.Intn(len(niches))]
		pipeline := tdg.rand.Float32() < 0.2 // 20% chance of being a pipeline command

		commands[i] = database.Command{
			Command:     command,
			Description: description,
			Keywords:    cmdKeywords,
			Platform:    platform,
			Pipeline:    pipeline,
			Niche:       niche,
		}
	}

	return commands
}

// GenerateEdgeCaseCommands generates commands for edge case testing
func (tdg *TestDataGenerator) GenerateEdgeCaseCommands() []database.Command {
	return []database.Command{
		// Empty fields
		{
			Command:     "",
			Description: "",
			Keywords:    []string{},
			Platform:    []string{},
			Pipeline:    false,
			Niche:       "",
		},
		// Very short command
		{
			Command:     "a",
			Description: "single character command",
			Keywords:    []string{"short"},
			Platform:    []string{"linux"},
			Pipeline:    false,
			Niche:       "test",
		},
		// Very long command
		{
			Command:     strings.Repeat("very-long-command-name-", 10),
			Description: strings.Repeat("very long description ", 20),
			Keywords:    []string{strings.Repeat("long-keyword-", 5)},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "test",
		},
		// Special characters
		{
			Command:     "command!@#$%^&*()_+-=[]{}|;':\",./<>?",
			Description: "command with special characters",
			Keywords:    []string{"special", "characters", "symbols"},
			Platform:    []string{"linux"},
			Pipeline:    false,
			Niche:       "test",
		},
		// Unicode characters
		{
			Command:     "命令 コマンド команда",
			Description: "command with unicode characters",
			Keywords:    []string{"unicode", "international", "characters"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "test",
		},
		// Mixed case
		{
			Command:     "MiXeD CaSe CoMmAnD",
			Description: "MiXeD CaSe DeScRiPtIoN",
			Keywords:    []string{"MiXeD", "CaSe", "TeSt"},
			Platform:    []string{"windows"},
			Pipeline:    false,
			Niche:       "TeSt",
		},
		// Whitespace variations
		{
			Command:     "  command  with  extra  spaces  ",
			Description: "\tcommand\twith\ttabs\tand\tspaces\n",
			Keywords:    []string{" spaced ", "\ttabbed\t", "\nnewlined\n"},
			Platform:    []string{"linux"},
			Pipeline:    false,
			Niche:       " test ",
		},
		// Numbers and symbols
		{
			Command:     "command123 --option=value",
			Description: "command with numbers 123 and symbols",
			Keywords:    []string{"123", "numbers", "symbols", "options"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "numeric",
		},
		// Pipeline with complex structure
		{
			Command:     "cat file.txt | grep pattern | sort | uniq -c | head -10",
			Description: "complex pipeline with multiple stages",
			Keywords:    []string{"pipeline", "complex", "multi-stage"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    true,
			Niche:       "advanced",
		},
		// Command with quotes
		{
			Command:     `echo "hello world" | grep 'hello'`,
			Description: "command with single and double quotes",
			Keywords:    []string{"quotes", "echo", "grep"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    true,
			Niche:       "text",
		},
	}
}

// GeneratePerformanceTestCommands generates commands for performance testing
func (tdg *TestDataGenerator) GeneratePerformanceTestCommands(count int) []database.Command {
	commands := make([]database.Command, count)

	// Base patterns for realistic commands
	patterns := []struct {
		command     string
		description string
		keywords    []string
		platform    []string
		niche       string
	}{
		{
			command:     "git %s",
			description: "git %s operation",
			keywords:    []string{"git", "version-control"},
			platform:    []string{"linux", "macos", "windows"},
			niche:       "development",
		},
		{
			command:     "find %s",
			description: "find %s in filesystem",
			keywords:    []string{"find", "search", "filesystem"},
			platform:    []string{"linux", "macos"},
			niche:       "system",
		},
		{
			command:     "docker %s",
			description: "docker %s operation",
			keywords:    []string{"docker", "container"},
			platform:    []string{"linux", "macos", "windows"},
			niche:       "devops",
		},
		{
			command:     "kubectl %s",
			description: "kubernetes %s operation",
			keywords:    []string{"kubectl", "kubernetes", "k8s"},
			platform:    []string{"linux", "macos", "windows"},
			niche:       "devops",
		},
		{
			command:     "npm %s",
			description: "npm %s operation",
			keywords:    []string{"npm", "node", "javascript"},
			platform:    []string{"linux", "macos", "windows"},
			niche:       "development",
		},
	}

	operations := []string{
		"init", "clone", "add", "commit", "push", "pull", "status", "log",
		"branch", "checkout", "merge", "rebase", "reset", "stash", "tag",
		"remote", "fetch", "diff", "show", "config", "help", "version",
	}

	for i := 0; i < count; i++ {
		pattern := patterns[i%len(patterns)]
		operation := operations[i%len(operations)]

		commands[i] = database.Command{
			Command:     fmt.Sprintf(pattern.command, operation),
			Description: fmt.Sprintf(pattern.description, operation),
			Keywords:    append(pattern.keywords, operation),
			Platform:    pattern.platform,
			Pipeline:    false,
			Niche:       pattern.niche,
		}
	}

	return commands
}

// GenerateTestQueries generates test queries for various scenarios
func (tdg *TestDataGenerator) GenerateTestQueries(count int) []TestQuery {
	queries := make([]TestQuery, count)

	queryPatterns := []string{
		"git commit", "find files", "docker run", "kubectl get", "npm install",
		"search text", "compress archive", "download file", "upload data",
		"create directory", "delete file", "copy data", "move file",
		"list items", "show status", "edit config", "run command",
		"start service", "stop process", "restart application",
	}

	for i := 0; i < count; i++ {
		pattern := queryPatterns[i%len(queryPatterns)]

		queries[i] = TestQuery{
			Query:            pattern,
			ExpectedResults:  1 + tdg.rand.Intn(3),            // 1-3 results
			MinScore:         float64(tdg.rand.Intn(10) + 5),  // 5-14
			MaxScore:         float64(tdg.rand.Intn(20) + 30), // 30-49
			ShouldContain:    strings.Fields(pattern)[:1],     // First word
			ShouldNotContain: []string{},
		}
	}

	return queries
}

// GenerateStressTestData generates large amounts of test data for stress testing
func (tdg *TestDataGenerator) GenerateStressTestData(commandCount, queryCount int) ([]database.Command, []TestQuery) {
	commands := tdg.GeneratePerformanceTestCommands(commandCount)
	queries := tdg.GenerateTestQueries(queryCount)

	return commands, queries
}

// GenerateBenchmarkData generates data specifically for benchmarking
func (tdg *TestDataGenerator) GenerateBenchmarkData() ([]database.Command, []TestQuery) {
	// Generate a realistic dataset size
	commands := tdg.GeneratePerformanceTestCommands(1000)

	// Add some edge cases
	commands = append(commands, tdg.GenerateEdgeCaseCommands()...)

	// Generate queries that should find results
	queries := []TestQuery{
		{Query: "git", ExpectedResults: 20, MinScore: 4.0, MaxScore: 50.0},
		{Query: "find", ExpectedResults: 20, MinScore: 4.0, MaxScore: 50.0},
		{Query: "docker", ExpectedResults: 20, MinScore: 4.0, MaxScore: 50.0},
		{Query: "kubectl", ExpectedResults: 20, MinScore: 4.0, MaxScore: 50.0},
		{Query: "npm", ExpectedResults: 20, MinScore: 4.0, MaxScore: 50.0},
		{Query: "commit", ExpectedResults: 5, MinScore: 4.0, MaxScore: 30.0},
		{Query: "run", ExpectedResults: 5, MinScore: 4.0, MaxScore: 30.0},
		{Query: "install", ExpectedResults: 5, MinScore: 4.0, MaxScore: 30.0},
		{Query: "get", ExpectedResults: 5, MinScore: 4.0, MaxScore: 30.0},
		{Query: "files", ExpectedResults: 5, MinScore: 4.0, MaxScore: 30.0},
	}

	return commands, queries
}

// GenerateMemoryTestData generates data for memory usage testing
func (tdg *TestDataGenerator) GenerateMemoryTestData(size string) []database.Command {
	var count int
	switch size {
	case "small":
		count = 100
	case "medium":
		count = 1000
	case "large":
		count = 10000
	case "xlarge":
		count = 100000
	default:
		count = 1000
	}

	return tdg.GeneratePerformanceTestCommands(count)
}

// GenerateRealisticCommands generates realistic commands based on common CLI tools
func (tdg *TestDataGenerator) GenerateRealisticCommands() []database.Command {
	var commands []database.Command

	// Git commands
	gitOps := []string{"init", "clone", "add", "commit", "push", "pull", "status", "log", "branch", "checkout", "merge"}
	for _, op := range gitOps {
		commands = append(commands, database.Command{
			Command:     fmt.Sprintf("git %s", op),
			Description: fmt.Sprintf("git %s operation", op),
			Keywords:    []string{"git", "version-control", op},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "development",
		})
	}

	// Docker commands
	dockerOps := []string{"run", "build", "pull", "push", "ps", "images", "exec", "logs", "stop", "rm"}
	for _, op := range dockerOps {
		commands = append(commands, database.Command{
			Command:     fmt.Sprintf("docker %s", op),
			Description: fmt.Sprintf("docker %s operation", op),
			Keywords:    []string{"docker", "container", op},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "devops",
		})
	}

	// File operations
	fileOps := []struct {
		cmd, desc string
		keywords  []string
	}{
		{"ls -la", "list files with details", []string{"ls", "list", "files"}},
		{"find . -name", "find files by name", []string{"find", "search", "files"}},
		{"grep -r", "search text recursively", []string{"grep", "search", "text"}},
		{"cp -r", "copy files recursively", []string{"cp", "copy", "files"}},
		{"mv", "move or rename files", []string{"mv", "move", "rename"}},
		{"rm -rf", "remove files and directories", []string{"rm", "remove", "delete"}},
	}

	for _, op := range fileOps {
		commands = append(commands, database.Command{
			Command:     op.cmd,
			Description: op.desc,
			Keywords:    op.keywords,
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "system",
		})
	}

	return commands
}
