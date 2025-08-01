package testutil

import (
	"github.com/Vedant9500/WTF/internal/database"
)

// TestDataSets provides various predefined test data sets
type TestDataSets struct{}

// NewTestDataSets creates a new test data sets provider
func NewTestDataSets() *TestDataSets {
	return &TestDataSets{}
}

// GetGitCommands returns Git-related commands for testing
func (tds *TestDataSets) GetGitCommands() []database.Command {
	return []database.Command{
		{
			Command:     "git init",
			Description: "initialize a new git repository",
			Keywords:    []string{"git", "init", "repository", "version-control"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "development",
		},
		{
			Command:     "git clone <url>",
			Description: "clone a remote repository",
			Keywords:    []string{"git", "clone", "remote", "repository"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "development",
		},
		{
			Command:     "git add .",
			Description: "add all files to staging area",
			Keywords:    []string{"git", "add", "staging", "files"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "development",
		},
		{
			Command:     "git commit -m 'message'",
			Description: "commit changes with message",
			Keywords:    []string{"git", "commit", "message", "save"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "development",
		},
		{
			Command:     "git push origin main",
			Description: "push changes to remote repository",
			Keywords:    []string{"git", "push", "remote", "upload"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "development",
		},
		{
			Command:     "git pull origin main",
			Description: "pull changes from remote repository",
			Keywords:    []string{"git", "pull", "remote", "download", "sync"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "development",
		},
		{
			Command:     "git status",
			Description: "show working tree status",
			Keywords:    []string{"git", "status", "changes", "working-tree"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "development",
		},
		{
			Command:     "git log --oneline",
			Description: "show commit history in one line format",
			Keywords:    []string{"git", "log", "history", "commits"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "development",
		},
	}
}

// GetFileOperationCommands returns file operation commands for testing
func (tds *TestDataSets) GetFileOperationCommands() []database.Command {
	return []database.Command{
		{
			Command:     "ls -la",
			Description: "list files with detailed information",
			Keywords:    []string{"ls", "list", "files", "directory"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "find . -name '*.txt'",
			Description: "find text files in current directory",
			Keywords:    []string{"find", "search", "files", "text"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "grep -r 'pattern' .",
			Description: "search for pattern in files recursively",
			Keywords:    []string{"grep", "search", "pattern", "text"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "cp source destination",
			Description: "copy files or directories",
			Keywords:    []string{"cp", "copy", "files", "duplicate"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "mv source destination",
			Description: "move or rename files",
			Keywords:    []string{"mv", "move", "rename", "files"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "rm -rf directory",
			Description: "remove directory and contents recursively",
			Keywords:    []string{"rm", "remove", "delete", "directory"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "mkdir -p path/to/directory",
			Description: "create directory with parent directories",
			Keywords:    []string{"mkdir", "create", "directory", "folder"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "touch filename",
			Description: "create empty file or update timestamp",
			Keywords:    []string{"touch", "create", "file", "timestamp"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "system",
		},
	}
}

// GetArchiveCommands returns archive/compression commands for testing
func (tds *TestDataSets) GetArchiveCommands() []database.Command {
	return []database.Command{
		{
			Command:     "tar -czf archive.tar.gz .",
			Description: "create compressed tar archive",
			Keywords:    []string{"tar", "compress", "archive", "gzip"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "tar -xzf archive.tar.gz",
			Description: "extract compressed tar archive",
			Keywords:    []string{"tar", "extract", "archive", "uncompress"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "zip -r archive.zip directory",
			Description: "create zip archive of directory",
			Keywords:    []string{"zip", "compress", "archive", "directory"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "unzip archive.zip",
			Description: "extract zip archive",
			Keywords:    []string{"unzip", "extract", "archive", "decompress"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "gzip filename",
			Description: "compress file with gzip",
			Keywords:    []string{"gzip", "compress", "file"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "gunzip filename.gz",
			Description: "decompress gzip file",
			Keywords:    []string{"gunzip", "decompress", "extract", "gzip"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "system",
		},
	}
}

// GetNetworkCommands returns network-related commands for testing
func (tds *TestDataSets) GetNetworkCommands() []database.Command {
	return []database.Command{
		{
			Command:     "curl -O https://example.com/file",
			Description: "download file from URL",
			Keywords:    []string{"curl", "download", "http", "url"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "network",
		},
		{
			Command:     "wget https://example.com/file",
			Description: "download file using wget",
			Keywords:    []string{"wget", "download", "http", "url"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "network",
		},
		{
			Command:     "ssh user@host",
			Description: "connect to remote host via SSH",
			Keywords:    []string{"ssh", "remote", "connect", "secure"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "network",
		},
		{
			Command:     "scp file user@host:/path",
			Description: "copy file to remote host",
			Keywords:    []string{"scp", "copy", "remote", "secure"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "network",
		},
		{
			Command:     "rsync -av source/ destination/",
			Description: "synchronize files and directories",
			Keywords:    []string{"rsync", "sync", "backup", "copy"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    false,
			Niche:       "network",
		},
		{
			Command:     "ping google.com",
			Description: "test network connectivity",
			Keywords:    []string{"ping", "network", "connectivity", "test"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "network",
		},
	}
}

// GetPipelineCommands returns pipeline commands for testing
func (tds *TestDataSets) GetPipelineCommands() []database.Command {
	return []database.Command{
		{
			Command:     "ps aux | grep process",
			Description: "find running processes",
			Keywords:    []string{"ps", "grep", "process", "running"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    true,
			Niche:       "system",
		},
		{
			Command:     "cat file.txt | sort | uniq",
			Description: "sort and remove duplicates from file",
			Keywords:    []string{"cat", "sort", "uniq", "duplicates"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    true,
			Niche:       "system",
		},
		{
			Command:     "ls -la | grep '^d'",
			Description: "list only directories",
			Keywords:    []string{"ls", "grep", "directories", "filter"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    true,
			Niche:       "system",
		},
		{
			Command:     "find . -name '*.log' | xargs rm",
			Description: "find and delete log files",
			Keywords:    []string{"find", "xargs", "rm", "log", "delete"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    true,
			Niche:       "system",
		},
		{
			Command:     "history | grep git",
			Description: "search command history for git commands",
			Keywords:    []string{"history", "grep", "git", "search"},
			Platform:    []string{"linux", "macos"},
			Pipeline:    true,
			Niche:       "system",
		},
	}
}

// GetWindowsCommands returns Windows-specific commands for testing
func (tds *TestDataSets) GetWindowsCommands() []database.Command {
	return []database.Command{
		{
			Command:     "dir /s",
			Description: "list files and directories recursively",
			Keywords:    []string{"dir", "list", "files", "recursive"},
			Platform:    []string{"windows"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "copy source destination",
			Description: "copy files in Windows",
			Keywords:    []string{"copy", "files", "windows"},
			Platform:    []string{"windows"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "del filename",
			Description: "delete file in Windows",
			Keywords:    []string{"del", "delete", "file", "windows"},
			Platform:    []string{"windows"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "md directory",
			Description: "create directory in Windows",
			Keywords:    []string{"md", "mkdir", "create", "directory"},
			Platform:    []string{"windows"},
			Pipeline:    false,
			Niche:       "system",
		},
		{
			Command:     "type filename",
			Description: "display file contents in Windows",
			Keywords:    []string{"type", "cat", "display", "file"},
			Platform:    []string{"windows"},
			Pipeline:    false,
			Niche:       "system",
		},
	}
}

// GetEdgeCaseCommands returns commands for edge case testing
func (tds *TestDataSets) GetEdgeCaseCommands() []database.Command {
	return []database.Command{
		{
			Command:     "a",
			Description: "single character command",
			Keywords:    []string{"short"},
			Platform:    []string{"linux"},
			Pipeline:    false,
			Niche:       "test",
		},
		{
			Command:     "command-with-many-hyphens-and-underscores_test",
			Description: "command with special characters",
			Keywords:    []string{"special", "characters", "hyphens", "underscores"},
			Platform:    []string{"linux", "macos", "windows"},
			Pipeline:    false,
			Niche:       "test",
		},
		{
			Command:     "UPPERCASE_COMMAND",
			Description: "UPPERCASE DESCRIPTION",
			Keywords:    []string{"UPPERCASE", "CASE", "TEST"},
			Platform:    []string{"linux"},
			Pipeline:    false,
			Niche:       "test",
		},
		{
			Command:     "command with spaces",
			Description: "command that contains spaces",
			Keywords:    []string{"spaces", "whitespace"},
			Platform:    []string{"linux"},
			Pipeline:    false,
			Niche:       "test",
		},
		{
			Command:     "",
			Description: "empty command for testing",
			Keywords:    []string{"empty", "test"},
			Platform:    []string{},
			Pipeline:    false,
			Niche:       "",
		},
	}
}

// GetAllTestCommands returns all test commands combined
func (tds *TestDataSets) GetAllTestCommands() []database.Command {
	var allCommands []database.Command
	
	allCommands = append(allCommands, tds.GetGitCommands()...)
	allCommands = append(allCommands, tds.GetFileOperationCommands()...)
	allCommands = append(allCommands, tds.GetArchiveCommands()...)
	allCommands = append(allCommands, tds.GetNetworkCommands()...)
	allCommands = append(allCommands, tds.GetPipelineCommands()...)
	allCommands = append(allCommands, tds.GetWindowsCommands()...)
	
	return allCommands
}

// GetTestCommandsByCategory returns commands filtered by category
func (tds *TestDataSets) GetTestCommandsByCategory(category string) []database.Command {
	switch category {
	case "git":
		return tds.GetGitCommands()
	case "file":
		return tds.GetFileOperationCommands()
	case "archive":
		return tds.GetArchiveCommands()
	case "network":
		return tds.GetNetworkCommands()
	case "pipeline":
		return tds.GetPipelineCommands()
	case "windows":
		return tds.GetWindowsCommands()
	case "edge":
		return tds.GetEdgeCaseCommands()
	default:
		return tds.GetAllTestCommands()
	}
}

// GetTestCommandsByPlatform returns commands filtered by platform
func (tds *TestDataSets) GetTestCommandsByPlatform(platform string) []database.Command {
	allCommands := tds.GetAllTestCommands()
	var filteredCommands []database.Command
	
	for _, cmd := range allCommands {
		if len(cmd.Platform) == 0 {
			// Commands with no platform specified work on all platforms
			filteredCommands = append(filteredCommands, cmd)
			continue
		}
		
		for _, p := range cmd.Platform {
			if p == platform {
				filteredCommands = append(filteredCommands, cmd)
				break
			}
		}
	}
	
	return filteredCommands
}