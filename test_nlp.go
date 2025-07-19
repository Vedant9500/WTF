package main

import (
	"fmt"
	"github.com/Vedant9500/WTF/internal/nlp"
)

func main() {
	p := nlp.NewQueryProcessor()
	
	queries := []string{
		"create directory",
		"make new folder",
		"mkdir",
	}
	
	for _, query := range queries {
		fmt.Printf("\n=== Testing: %s ===\n", query)
		q := p.ProcessQuery(query)
		fmt.Printf("Intent: %s\n", q.Intent)
		fmt.Printf("Actions: %v\n", q.Actions)
		fmt.Printf("Targets: %v\n", q.Targets)
		fmt.Printf("Keywords: %v\n", q.Keywords)
		fmt.Printf("Enhanced: %v\n", q.GetEnhancedKeywords())
	}
}