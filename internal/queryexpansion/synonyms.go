// Package queryexpansion provides query expansion with domain-specific synonyms
// for improved semantic search recall.
package queryexpansion

import (
	"strings"
)

// DomainSynonyms maps common query terms to domain-specific synonyms
// for command-line tools and development workflows.
var DomainSynonyms = map[string][]string{
	// Git operations
	"undo":     {"reset", "revert", "rollback", "restore"},
	"revert":   {"undo", "reset", "rollback"},
	"commit":   {"commit", "stash"},
	"branch":   {"branch", "checkout", "switch"},
	"history":  {"log", "show", "history"},
	"compare":  {"diff", "compare"},
	"merge":    {"merge", "rebase"},

	// File operations
	"delete":   {"remove", "rm", "rmdir", "destroy"},
	"remove":   {"delete", "rm", "rmdir"},
	"create":   {"create", "make", "mkdir", "touch", "new"},
	"make":     {"create", "make", "mkdir", "build"},
	"list":     {"list", "ls", "show", "display", "dir"},
	"show":     {"show", "list", "display", "view", "cat", "ls"},
	"view":     {"view", "show", "display", "cat", "less", "more"},
	"display":  {"display", "show", "list", "view"},
	"copy":     {"copy", "cp", "duplicate"},
	"move":     {"move", "mv", "rename"},
	"rename":   {"rename", "mv", "move"},

	// Compression/Archive
	"compress": {"compress", "archive", "zip", "tar", "gzip", "pack"},
	"archive":  {"archive", "compress", "zip", "tar", "gzip"},
	"extract":  {"extract", "unzip", "unpack", "decompress", "gunzip"},
	"unzip":    {"unzip", "extract", "unpack"},
	"decompress": {"decompress", "extract", "unzip", "unpack"},

	// Docker
	"container": {"container", "docker", "exec"},
	"image":     {"image", "docker", "build"},
	"compose":   {"compose", "docker-compose"},

	// Network
	"download": {"download", "wget", "curl", "fetch", "get"},
	"fetch":    {"fetch", "download", "pull", "curl", "wget"},
	"port":     {"port", "netstat", "ss", "lsof"},
	"network":  {"network", "ip", "ifconfig"},

	// Search/Text processing
	"search":   {"search", "grep", "find", "locate", "ripgrep"},
	"find":     {"find", "search", "locate", "grep"},
	"replace":  {"replace", "sed", "substitute", "perl"},
	"sort":     {"sort", "order"},
	"unique":   {"unique", "uniq", "distinct"},

	// System/Process
	"process":  {"process", "ps", "top", "task"},
	"kill":     {"kill", "pkill", "killall", "terminate", "stop"},
	"stop":     {"stop", "kill", "terminate", "pkill"},
	"monitor":  {"monitor", "top", "htop", "watch"},
	"service":  {"service", "systemctl", "daemon"},

	// Package management
	"install":  {"install", "add", "pip", "apt", "npm", "brew"},
	"update":   {"update", "upgrade"},

	// Permissions
	"permission": {"permission", "chmod", "access", "rights"},
	"executable": {"executable", "chmod", "exec"},

	// SSH/Remote
	"remote":   {"remote", "ssh", "scp", "rsync"},
	"connect":  {"connect", "ssh"},
}

// ExpandQuery expands a query with domain-specific synonyms to improve recall.
// Returns the expanded query string with synonyms appended.
func ExpandQuery(query string) string {
	tokens := tokenize(query)
	expanded := make([]string, len(tokens))
	copy(expanded, tokens)

	// Add synonyms for each token
	for _, token := range tokens {
		if synonyms, exists := DomainSynonyms[token]; exists {
			expanded = append(expanded, synonyms...)
		}
	}

	return strings.Join(expanded, " ")
}

// ExpandQueryTerms expands a list of query terms with domain-specific synonyms.
// Returns deduplicated terms including original terms and their synonyms.
func ExpandQueryTerms(terms []string) []string {
	seen := make(map[string]bool)
	expanded := make([]string, 0, len(terms)*3)

	// Add original terms
	for _, term := range terms {
		termLower := strings.ToLower(term)
		if !seen[termLower] {
			seen[termLower] = true
			expanded = append(expanded, termLower)
		}
	}

	// Add synonyms
	for _, term := range terms {
		termLower := strings.ToLower(term)
		if synonyms, exists := DomainSynonyms[termLower]; exists {
			for _, synonym := range synonyms {
				synonymLower := strings.ToLower(synonym)
				if !seen[synonymLower] {
					seen[synonymLower] = true
					expanded = append(expanded, synonymLower)
				}
			}
		}
	}

	return expanded
}

// tokenize splits a query into lowercase tokens.
func tokenize(query string) []string {
	query = strings.ToLower(query)
	tokens := strings.FieldsFunc(query, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})

	// Filter out very short tokens (< 2 chars) except common commands
	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if len(token) >= 2 {
			result = append(result, token)
		}
	}

	return result
}
