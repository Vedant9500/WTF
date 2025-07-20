// WTF (What's The Function) - Natural Language Command Discovery Tool
//
// WTF is a powerful command-line tool that helps users discover and understand
// terminal commands using natural language queries. It features:
//   - Natural language processing for intuitive queries
//   - Fuzzy search with typo tolerance
//   - Context-aware search based on project type
//   - Comprehensive command database with 3850+ commands
//   - Search history and analytics
//
// Usage:
//
//	wtf "how do I create a folder"
//	wtf search "compress files"
//	wtf history
//
// For more information, visit: https://github.com/Vedant9500/WTF
package main

import (
	"fmt"
	"os"

	"github.com/Vedant9500/WTF/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
