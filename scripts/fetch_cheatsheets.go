package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// GitHubFile represents a file from GitHub API
type GitHubFile struct {
	Name        string `json:"name"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
}

// Command represents a single command entry
type Command struct {
	Command     string   `yaml:"command"`
	Description string   `yaml:"description"`
	Keywords    []string `yaml:"keywords"`
	Niche       string   `yaml:"niche,omitempty"`
	Platform    []string `yaml:"platform,omitempty"`
	Pipeline    bool     `yaml:"pipeline"`
}

func main() {
	fmt.Println("Fetching cheatsheets from cheat/cheatsheets repository...")
	
	// Get list of files from GitHub API
	files, err := getCheatsheetFiles()
	if err != nil {
		fmt.Printf("Error fetching file list: %v\n", err)
		return
	}
	
	fmt.Printf("Found %d cheatsheet files\n", len(files))
	
	var allCommands []Command
	
	// Process each cheatsheet file
	for i, file := range files {
		if file.Type != "file" {
			continue
		}
		
		fmt.Printf("Processing %d/%d: %s\n", i+1, len(files), file.Name)
		
		commands, err := processCheatsheet(file.Name, file.DownloadURL)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", file.Name, err)
			continue
		}
		
		allCommands = append(allCommands, commands...)
		
		// Add small delay to be respectful to GitHub API
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Printf("Processed %d total commands\n", len(allCommands))
	
	// Save to YAML file
	err = saveCommandsToYAML(allCommands, "commands.yml")
	if err != nil {
		fmt.Printf("Error saving commands: %v\n", err)
		return
	}
	
	fmt.Println("Successfully created commands.yml")
}

func getCheatsheetFiles() ([]GitHubFile, error) {
	url := "https://api.github.com/repos/cheat/cheatsheets/contents"
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var files []GitHubFile
	err = json.NewDecoder(resp.Body).Decode(&files)
	if err != nil {
		return nil, err
	}
	
	return files, nil
}

func processCheatsheet(filename, downloadURL string) ([]Command, error) {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	return parseCheatsheet(filename, string(content))
}

func parseCheatsheet(filename, content string) ([]Command, error) {
	var commands []Command
	
	lines := strings.Split(content, "\n")
	var tags []string
	
	// Extract tags from frontmatter
	inFrontmatter := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}
		
		if inFrontmatter && strings.HasPrefix(line, "tags:") {
			// Parse tags: [ compression, archive ]
			tagMatch := regexp.MustCompile(`\[(.*?)\]`).FindStringSubmatch(line)
			if len(tagMatch) > 1 {
				tagStr := strings.ReplaceAll(tagMatch[1], " ", "")
				tags = strings.Split(tagStr, ",")
			}
		}
	}
	
	// Parse commands
	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentDescription string
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip frontmatter and empty lines
		if line == "---" || line == "" {
			continue
		}
		
		// Check if this is a description line
		if strings.HasPrefix(line, "# To ") {
			currentDescription = strings.TrimPrefix(line, "# To ")
			currentDescription = strings.TrimSuffix(currentDescription, ":")
			continue
		}
		
		// Check if this is a command line (not a comment)
		if !strings.HasPrefix(line, "#") && currentDescription != "" && line != "" {
			// Clean up the command
			command := cleanCommand(line)
			if command != "" {
				keywords := append([]string{filename}, tags...)
				keywords = append(keywords, extractKeywords(currentDescription)...)
				
				commands = append(commands, Command{
					Command:     command,
					Description: currentDescription,
					Keywords:    removeDuplicates(keywords),
					Niche:       determineNiche(filename, tags),
					Platform:    []string{"linux", "macos"},
					Pipeline:    strings.Contains(command, "|"),
				})
			}
			currentDescription = ""
		}
	}
	
	return commands, nil
}

func cleanCommand(command string) string {
	// Remove leading/trailing whitespace
	command = strings.TrimSpace(command)
	
	// Skip if it's still a comment or empty
	if strings.HasPrefix(command, "#") || command == "" {
		return ""
	}
	
	return command
}

func extractKeywords(description string) []string {
	// Simple keyword extraction from description
	words := strings.Fields(strings.ToLower(description))
	var keywords []string
	
	// Filter out common stop words and extract meaningful terms
	stopWords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
		"be": true, "by": true, "for": true, "from": true, "has": true, "he": true,
		"in": true, "is": true, "it": true, "its": true, "of": true, "on": true,
		"that": true, "the": true, "to": true, "was": true, "will": true, "with": true,
	}
	
	for _, word := range words {
		word = strings.Trim(word, ".,!?:;")
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

func determineNiche(filename string, tags []string) string {
	// Map common tools to niches
	niches := map[string]string{
		"git":        "version-control",
		"docker":     "containers",
		"tar":        "compression",
		"zip":        "compression",
		"find":       "filesystem",
		"grep":       "text-processing",
		"awk":        "text-processing",
		"sed":        "text-processing",
		"curl":       "networking",
		"wget":       "networking",
		"ssh":        "networking",
		"mysql":      "database",
		"postgres":   "database",
		"nginx":      "web-server",
		"apache":     "web-server",
	}
	
	if niche, exists := niches[filename]; exists {
		return niche
	}
	
	// Check tags
	for _, tag := range tags {
		if niche, exists := niches[tag]; exists {
			return niche
		}
	}
	
	return ""
}

func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

func saveCommandsToYAML(commands []Command, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Write YAML header
	file.WriteString("# CLI Command Database\n")
	file.WriteString("# Generated from cheat/cheatsheets repository\n\n")
	
	for _, cmd := range commands {
		file.WriteString("- command: \"" + escapeYAML(cmd.Command) + "\"\n")
		file.WriteString("  description: \"" + escapeYAML(cmd.Description) + "\"\n")
		
		// Properly format keywords array with quoted strings
		file.WriteString("  keywords: [")
		for i, keyword := range cmd.Keywords {
			if i > 0 {
				file.WriteString(", ")
			}
			file.WriteString("\"" + escapeYAML(keyword) + "\"")
		}
		file.WriteString("]\n")
		
		if cmd.Niche != "" {
			file.WriteString("  niche: \"" + cmd.Niche + "\"\n")
		}
		
		file.WriteString("  platform: [")
		for i, platform := range cmd.Platform {
			if i > 0 {
				file.WriteString(", ")
			}
			file.WriteString(platform)
		}
		file.WriteString("]\n")
		
		file.WriteString(fmt.Sprintf("  pipeline: %t\n", cmd.Pipeline))
		file.WriteString("\n")
	}
	
	return nil
}

func escapeYAML(s string) string {
	// Basic YAML string escaping
	s = strings.ReplaceAll(s, "\\", "\\\\") // Escape backslashes first
	s = strings.ReplaceAll(s, "\"", "\\\"") // Escape quotes
	s = strings.ReplaceAll(s, "\n", "\\n")  // Escape newlines
	s = strings.ReplaceAll(s, "\r", "\\r")  // Escape carriage returns
	s = strings.ReplaceAll(s, "\t", "\\t")  // Escape tabs
	return s
}