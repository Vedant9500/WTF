package database

import (
	"testing"
)

func TestSearchWithPipelineOptions(t *testing.T) {
	// Create test database
	db := &Database{
		Commands: []Command{
			{
				Command:          "ls -la",
				Description:      "list files",
				Keywords:         []string{"ls", "list"},
				Pipeline:         false,
				CommandLower:     "ls -la",
				DescriptionLower: "list files",
				KeywordsLower:    []string{"ls", "list"},
			},
			{
				Command:          "find . -name '*.txt' | grep test | head -5",
				Description:      "find text files with test in name",
				Keywords:         []string{"find", "search", "text"},
				Pipeline:         true,
				CommandLower:     "find . -name '*.txt' | grep test | head -5",
				DescriptionLower: "find text files with test in name",
				KeywordsLower:    []string{"find", "search", "text"},
			},
			{
				Command:          "cat file.log | grep ERROR | tail -10",
				Description:      "show last 10 errors from log",
				Keywords:         []string{"log", "error", "tail"},
				Pipeline:         false, // Not marked as pipeline but has pipe
				CommandLower:     "cat file.log | grep error | tail -10",
				DescriptionLower: "show last 10 errors from log",
				KeywordsLower:    []string{"log", "error", "tail"},
			},
			{
				Command:          "ps aux",
				Description:      "show processes",
				Keywords:         []string{"process", "ps"},
				Pipeline:         false,
				CommandLower:     "ps aux",
				DescriptionLower: "show processes",
				KeywordsLower:    []string{"process", "ps"},
			},
		},
	}

	// Test pipeline-only search
	options := SearchOptions{
		Limit:        10,
		PipelineOnly: true,
	}

	results := db.SearchWithPipelineOptions("grep", options)

	// Should only return pipeline commands
	expectedPipelineCommands := 2 // Commands with Pipeline=true or containing "|"
	if len(results) != expectedPipelineCommands {
		t.Errorf("Expected %d pipeline commands, got %d", expectedPipelineCommands, len(results))
	}

	// Verify returned commands are pipelines
	for _, result := range results {
		if !result.Command.Pipeline && !isPipelineCommand(result.Command.Command) {
			t.Errorf("Non-pipeline command returned: %s", result.Command.Command)
		}
	}
}

func TestSearchWithPipelineBoost(t *testing.T) {
	// Create test database
	db := &Database{
		Commands: []Command{
			{
				Command:          "grep test file.txt",
				Description:      "search for test in file",
				Keywords:         []string{"grep", "search"},
				Pipeline:         false,
				CommandLower:     "grep test file.txt",
				DescriptionLower: "search for test in file",
				KeywordsLower:    []string{"grep", "search"},
			},
			{
				Command:          "cat file.txt | grep test | head -5",
				Description:      "pipeline search for test",
				Keywords:         []string{"grep", "search", "pipeline"},
				Pipeline:         true,
				CommandLower:     "cat file.txt | grep test | head -5",
				DescriptionLower: "pipeline search for test",
				KeywordsLower:    []string{"grep", "search", "pipeline"},
			},
		},
	}

	// Test without pipeline boost
	optionsNormal := SearchOptions{
		Limit: 10,
	}
	// Get baseline scores
	_ = db.SearchWithPipelineOptions("grep test", optionsNormal)

	// Test with pipeline boost
	optionsBoost := SearchOptions{
		Limit:         10,
		PipelineBoost: 2.0,
	}
	resultsBoost := db.SearchWithPipelineOptions("grep test", optionsBoost)

	// Pipeline command should rank higher with boost
	if len(resultsBoost) >= 2 {
		// Find the pipeline command in results
		var pipelineScore, normalScore float64
		for _, result := range resultsBoost {
			if result.Command.Pipeline {
				pipelineScore = result.Score
			} else {
				normalScore = result.Score
			}
		}

		if pipelineScore <= normalScore {
			t.Errorf("Pipeline boost not working: pipeline score %f <= normal score %f", pipelineScore, normalScore)
		}
	}
}

func TestIsPipelineCommand(t *testing.T) {
	tests := []struct {
		command  string
		expected bool
	}{
		{"ls -la", false},
		{"cat file | grep test", true},
		{"find . -name '*.txt' | head -5", true},
		{"echo 'test' && echo 'done'", true},
		{"cat file >> output.txt", true},
		{"grep test file.txt", false},
		{"command with pipe in description", true},
		{"PIPE uppercase test", true},
	}

	for _, test := range tests {
		result := isPipelineCommand(test.command)
		if result != test.expected {
			t.Errorf("isPipelineCommand(%q) = %v, expected %v", test.command, result, test.expected)
		}
	}
}
