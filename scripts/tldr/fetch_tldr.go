package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// TldrCommand represents a command parsed from tldr pages
type TldrCommand struct {
	Command     string   `json:"command"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	Category    string   `json:"category"`
	Pipeline    bool     `json:"pipeline"`
	Source      string   `json:"source"`
}

// GitHubAPIResponse represents GitHub API response for directory listing
type GitHubAPIResponse []struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

// TldrFetcher handles fetching and parsing tldr pages
type TldrFetcher struct {
	baseURL     string
	apiURL      string
	outputFile  string
	maxPages    int
	concurrency int
	categories  map[string]string
	client      *http.Client
}

// NewTldrFetcher creates a new fetcher instance
func NewTldrFetcher(outputFile string, maxPages, concurrency int) *TldrFetcher {
	return &TldrFetcher{
		baseURL:     "https://raw.githubusercontent.com/tldr-pages/tldr/main/pages",
		apiURL:      "https://api.github.com/repos/tldr-pages/tldr/contents/pages",
		outputFile:  outputFile,
		maxPages:    maxPages,
		concurrency: concurrency,
		client:      &http.Client{Timeout: 15 * time.Second}, // Reduced timeout
		categories: map[string]string{
			"common":  "general",
			"linux":   "system",
			"osx":     "system",
			"windows": "system",
			"android": "mobile",
			"freebsd": "system",
			"netbsd":  "system",
			"openbsd": "system",
			"sunos":   "system",
		},
	}
}

// fetchDirectoryListing gets the list of directories from GitHub API
func (f *TldrFetcher) fetchDirectoryListing(path string) ([]string, error) {
	url := f.apiURL
	if path != "" {
		url = fmt.Sprintf("%s/%s", f.apiURL, path)
	}

	fmt.Printf("🌐 Fetching directory listing from: %s\n", url)

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch directory listing: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var response GitHubAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	var directories []string
	for _, item := range response {
		if item.Type == "dir" {
			directories = append(directories, item.Name)
		}
	}

	return directories, nil
}

// fetchPageListing gets the list of .md files from a pages directory
func (f *TldrFetcher) fetchPageListing(pagesDir string) ([]string, error) {
	url := fmt.Sprintf("%s/%s", f.apiURL, pagesDir)

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page listing: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d for %s", resp.StatusCode, pagesDir)
	}

	var response GitHubAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	var pages []string
	for _, item := range response {
		if item.Type == "file" && strings.HasSuffix(item.Name, ".md") && !strings.Contains(item.Name, "%") {
			pages = append(pages, item.Name)
		}
	}

	return pages, nil
}

// fetchTldrPage downloads and parses a single tldr page
func (f *TldrFetcher) fetchTldrPage(pagesDir, filename string) (*TldrCommand, error) {
	// Determine platform/category from directory name
	platform := strings.TrimPrefix(pagesDir, "pages")
	if platform == "" {
		platform = "common"
	} else {
		platform = strings.TrimPrefix(platform, ".")
	}

	category, exists := f.categories[platform]
	if !exists {
		category = "other"
	}

	// Construct download URL
	url := fmt.Sprintf("%s/%s/%s", f.baseURL, pagesDir, filename)

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page %s: %w", filename, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch page %s: status %d", filename, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read page content: %w", err)
	}

	return f.parseTldrPage(string(content), filename, category)
}

// parseTldrPage parses the markdown content of a tldr page
func (f *TldrFetcher) parseTldrPage(content, filename, category string) (*TldrCommand, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 {
		return nil, fmt.Errorf("invalid tldr page format")
	}

	// Extract command name from filename
	commandName := strings.TrimSuffix(filename, ".md")

	// Extract description from the first few lines
	var description strings.Builder
	var commands []string

	// Regex patterns for parsing
	descLineRegex := regexp.MustCompile(`^>\s*(.+)$`)
	commandRegex := regexp.MustCompile("^`(.+)`$")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip the title line (starts with #)
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Description lines start with >
		if match := descLineRegex.FindStringSubmatch(line); match != nil {
			if !strings.Contains(match[1], "More information:") && !strings.Contains(match[1], "<http") {
				if description.Len() > 0 {
					description.WriteString(" ")
				}
				description.WriteString(match[1])
			}
			continue
		}

		// Command examples are in backticks
		if match := commandRegex.FindStringSubmatch(line); match != nil {
			command := match[1]
			// Clean up command (remove placeholders and simplify)
			command = f.cleanCommand(command)
			if command != "" {
				commands = append(commands, command)
			}
		}
	}

	// If no commands found, return nil
	if len(commands) == 0 {
		return nil, fmt.Errorf("no commands found in page %s", filename)
	}

	// Take the first, most basic command
	mainCommand := commands[0]

	// Generate description
	desc := description.String()
	if desc == "" {
		desc = fmt.Sprintf("%s command", commandName)
	}

	// Generate keywords
	keywords := f.generateKeywords(commandName, desc, mainCommand)

	// Check if it's a pipeline command
	isPipeline := f.isPipelineCommand(mainCommand)

	return &TldrCommand{
		Command:     mainCommand,
		Description: desc,
		Keywords:    keywords,
		Category:    category,
		Pipeline:    isPipeline,
		Source:      "tldr-pages",
	}, nil
}

// cleanCommand removes placeholders and simplifies commands
func (f *TldrFetcher) cleanCommand(command string) string {
	// Remove placeholder syntax {{...}}
	placeholderRegex := regexp.MustCompile(`\{\{[^}]+\}\}`)
	command = placeholderRegex.ReplaceAllString(command, "")

	// Remove extra spaces
	spaceRegex := regexp.MustCompile(`\s+`)
	command = spaceRegex.ReplaceAllString(command, " ")

	// Trim and return
	return strings.TrimSpace(command)
}

// generateKeywords creates relevant keywords for a command
func (f *TldrFetcher) generateKeywords(commandName, description, command string) []string {
	keywords := []string{commandName}

	// Add words from description
	descWords := strings.Fields(strings.ToLower(description))
	for _, word := range descWords {
		// Clean word and add if it's meaningful
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 2 && !contains(keywords, word) {
			keywords = append(keywords, word)
		}
	}

	// Add command parts
	commandParts := strings.Fields(command)
	for _, part := range commandParts {
		if len(part) > 1 && !strings.HasPrefix(part, "-") && !contains(keywords, part) {
			keywords = append(keywords, part)
		}
	}

	// Limit keywords to avoid bloat
	if len(keywords) > 8 {
		keywords = keywords[:8]
	}

	return keywords
}

// isPipelineCommand checks if a command contains pipeline operators
func (f *TldrFetcher) isPipelineCommand(command string) bool {
	pipelineIndicators := []string{"|", "&&", "||", ";", ">", ">>"}
	for _, indicator := range pipelineIndicators {
		if strings.Contains(command, indicator) {
			return true
		}
	}
	return false
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// fetchAllCommands fetches commands from tldr-pages using concurrent workers
func (f *TldrFetcher) fetchAllCommands() ([]TldrCommand, error) {
	fmt.Println("🚀 Starting optimized tldr-pages fetch...")

	// Get list of page directories
	directories, err := f.fetchDirectoryListing("")
	if err != nil {
		return nil, fmt.Errorf("failed to get directory listing: %w", err)
	}

	fmt.Printf("📁 Found %d page directories\n", len(directories))

	// Collect all page information first
	type PageJob struct {
		Dir      string
		Filename string
	}

	var allJobs []PageJob
	for _, dir := range directories {
		fmt.Printf("📖 Collecting pages from: %s\n", dir)

		pages, err := f.fetchPageListing(dir)
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to get pages from %s: %v\n", dir, err)
			continue
		}

		fmt.Printf("   📄 Found %d pages\n", len(pages))
		for _, page := range pages {
			allJobs = append(allJobs, PageJob{Dir: dir, Filename: page})
		}
	}

	// Limit jobs if needed
	if len(allJobs) > f.maxPages {
		allJobs = allJobs[:f.maxPages]
		fmt.Printf("⚡ Limited to %d pages\n", f.maxPages)
	}

	fmt.Printf("🔄 Processing %d pages with %d workers...\n", len(allJobs), f.concurrency)

	// Create channels for work distribution
	jobChan := make(chan PageJob, len(allJobs))
	resultChan := make(chan *TldrCommand, len(allJobs))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < f.concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobChan {
				if cmd, err := f.fetchTldrPage(job.Dir, job.Filename); err == nil {
					resultChan <- cmd
				} else {
					// Send nil for failed jobs
					resultChan <- nil
				}
			}
		}(i)
	}

	// Send all jobs to workers
	for _, job := range allJobs {
		jobChan <- job
	}
	close(jobChan)

	// Collect results
	var commands []TldrCommand
	go func() {
		processed := 0
		for i := 0; i < len(allJobs); i++ {
			cmd := <-resultChan
			if cmd != nil {
				commands = append(commands, *cmd)
			}
			processed++

			if processed%50 == 0 {
				fmt.Printf("   ✅ Processed %d/%d pages...\n", processed, len(allJobs))
			}
		}
	}()

	// Wait for all workers to finish
	wg.Wait()

	// Give result collector a moment to finish
	time.Sleep(200 * time.Millisecond)

	fmt.Printf("✅ Successfully fetched %d commands from tldr-pages\n", len(commands))
	return commands, nil
}

// saveToYAML saves commands to a YAML file
func (f *TldrFetcher) saveToYAML(commands []TldrCommand) error {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(f.outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(f.outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Write header
	header := fmt.Sprintf(`# WTF Command Database - TLDR Edition
# Fetched from tldr-pages on %s
# Total commands: %d
# Source: https://github.com/tldr-pages/tldr

commands:
`, time.Now().Format("2006-01-02 15:04:05"), len(commands))

	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write each command in YAML format
	for _, cmd := range commands {
		cmdYAML := fmt.Sprintf(`  - command: "%s"
    description: "%s"
    keywords: [%s]
    category: "%s"
    pipeline: %t
    source: "%s"
`,
			escapeYAMLString(cmd.Command),
			escapeYAMLString(cmd.Description),
			formatKeywords(cmd.Keywords),
			cmd.Category,
			cmd.Pipeline,
			cmd.Source)

		if _, err := file.WriteString(cmdYAML); err != nil {
			return fmt.Errorf("failed to write command: %w", err)
		}
	}

	return nil
}

// escapeYAMLString escapes quotes in YAML strings
func escapeYAMLString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

// formatKeywords formats keywords array for YAML
func formatKeywords(keywords []string) string {
	var quoted []string
	for _, kw := range keywords {
		quoted = append(quoted, fmt.Sprintf(`"%s"`, escapeYAMLString(kw)))
	}
	return strings.Join(quoted, ", ")
}

// run executes the fetching process
func (f *TldrFetcher) run() error {
	commands, err := f.fetchAllCommands()
	if err != nil {
		return err
	}

	if len(commands) == 0 {
		return fmt.Errorf("no commands were fetched")
	}

	if err := f.saveToYAML(commands); err != nil {
		return fmt.Errorf("failed to save commands: %w", err)
	}

	fmt.Printf("🎉 Successfully saved %d commands to %s\n", len(commands), f.outputFile)
	return nil
}

func main() {
	// Configuration
	outputFile := "../../assets/commands_tldr.yml"
	maxPages := 2000
	concurrency := 8 // Number of concurrent workers

	// Parse command line arguments
	args := os.Args[1:]
	for i, arg := range args {
		switch arg {
		case "--output", "-o":
			if i+1 < len(args) {
				outputFile = args[i+1]
			}
		case "--max", "-m":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &maxPages)
			}
		case "--workers", "-w":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &concurrency)
			}
		case "--help", "-h":
			fmt.Println("Usage: go run fetch_tldr.go [OPTIONS]")
			fmt.Println("Options:")
			fmt.Println("  --output, -o FILE     Output YAML file (default: ../../assets/commands_tldr.yml)")
			fmt.Println("  --max, -m NUMBER      Maximum pages to fetch (default: 2000)")
			fmt.Println("  --workers, -w NUMBER  Number of concurrent workers (default: 8)")
			fmt.Println("  --help, -h            Show this help")
			return
		}
	}

	// Create fetcher and run
	fetcher := NewTldrFetcher(outputFile, maxPages, concurrency)

	fmt.Printf("🎯 WTF TLDR Fetcher (Optimized)\n")
	fmt.Printf("📂 Output file: %s\n", outputFile)
	fmt.Printf("📊 Max pages: %d\n", maxPages)
	fmt.Printf("👥 Workers: %d\n", concurrency)
	fmt.Println()

	start := time.Now()
	if err := fetcher.run(); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(start)
	fmt.Printf("⏱️  Total time: %v\n", elapsed.Round(time.Second))
}
