package testutil

import (
	"testing"

	"github.com/Vedant9500/WTF/internal/database"
)

func TestTestUtilPackage(t *testing.T) {
	// Test that we can create test database
	testDB := NewTestDatabase()
	if testDB == nil {
		t.Fatal("NewTestDatabase returned nil")
	}

	// Test that we can create test fixtures
	fixtures := NewTestFixtures()
	if fixtures == nil {
		t.Fatal("NewTestFixtures returned nil")
	}

	// Test that we can create a minimal database
	db := testDB.CreateMinimalDB()
	if db == nil {
		t.Fatal("CreateMinimalDB returned nil")
	}

	if len(db.Commands) == 0 {
		t.Error("Expected minimal database to have commands")
	}

	// Test that we can get sample commands
	commands := fixtures.GetSampleCommands()
	if len(commands) == 0 {
		t.Error("Expected sample commands to be non-empty")
	}

	// Test that we can get test queries
	queries := fixtures.GetTestQueries()
	if len(queries) == 0 {
		t.Error("Expected test queries to be non-empty")
	}
}

func TestTestDataSets(t *testing.T) {
	dataSets := NewTestDataSets()
	if dataSets == nil {
		t.Fatal("NewTestDataSets returned nil")
	}

	// Test git commands
	gitCommands := dataSets.GetGitCommands()
	if len(gitCommands) == 0 {
		t.Error("Expected git commands to be non-empty")
	}

	// Test file operation commands
	fileCommands := dataSets.GetFileOperationCommands()
	if len(fileCommands) == 0 {
		t.Error("Expected file operation commands to be non-empty")
	}

	// Test all commands
	allCommands := dataSets.GetAllTestCommands()
	if len(allCommands) == 0 {
		t.Error("Expected all test commands to be non-empty")
	}

	// Verify all commands have required fields populated
	for i, cmd := range allCommands {
		if cmd.Command == "" {
			t.Errorf("Command %d has empty Command field", i)
		}
		if cmd.Description == "" {
			t.Errorf("Command %d has empty Description field", i)
		}
		if len(cmd.Keywords) == 0 {
			t.Errorf("Command %d has no keywords", i)
		}
	}
}

func TestTestQuerySets(t *testing.T) {
	querySets := NewTestQuerySets()
	if querySets == nil {
		t.Fatal("NewTestQuerySets returned nil")
	}

	// Test basic queries
	basicQueries := querySets.GetBasicQueries()
	if len(basicQueries) == 0 {
		t.Error("Expected basic queries to be non-empty")
	}

	// Test advanced queries
	advancedQueries := querySets.GetAdvancedQueries()
	if len(advancedQueries) == 0 {
		t.Error("Expected advanced queries to be non-empty")
	}

	// Test edge case queries
	edgeQueries := querySets.GetEdgeCaseQueries()
	if len(edgeQueries) == 0 {
		t.Error("Expected edge case queries to be non-empty")
	}

	// Verify all queries have required fields
	allQueries := querySets.GetAllTestQueries()
	for i, query := range allQueries {
		if query.Query == "" && query.ExpectedResults > 0 {
			t.Errorf("Query %d has empty Query field but expects results", i)
		}
		if query.MinScore > query.MaxScore {
			t.Errorf("Query %d has MinScore > MaxScore", i)
		}
	}
}

func TestTestDataGenerator(t *testing.T) {
	generator := NewTestDataGenerator()
	if generator == nil {
		t.Fatal("NewTestDataGenerator returned nil")
	}

	// Test generating random commands
	commands := generator.GenerateRandomCommands(10)
	if len(commands) != 10 {
		t.Errorf("Expected 10 commands, got %d", len(commands))
	}

	// Test generating edge case commands
	edgeCommands := generator.GenerateEdgeCaseCommands()
	if len(edgeCommands) == 0 {
		t.Error("Expected edge case commands to be non-empty")
	}

	// Test generating performance test commands
	perfCommands := generator.GeneratePerformanceTestCommands(5)
	if len(perfCommands) != 5 {
		t.Errorf("Expected 5 performance commands, got %d", len(perfCommands))
	}

	// Test generating test queries
	queries := generator.GenerateTestQueries(5)
	if len(queries) != 5 {
		t.Errorf("Expected 5 test queries, got %d", len(queries))
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test helper functions work without panicking
	testDB := NewTestDatabase()
	db := testDB.CreateLargeDB()

	// Test assertion helpers
	results := []database.SearchResult{
		{
			Command: &db.Commands[0],
			Score:   10.0,
		},
	}

	// These should not panic
	AssertResultCount(t, results, 1)
	AssertScoreRange(t, 10.0, 5.0, 15.0)
	AssertCommandExists(t, results, db.Commands[0].Command)
	AssertContainsKeywords(t, &db.Commands[0], db.Commands[0].Keywords[:1])
}

func TestDatabaseTestHelper(t *testing.T) {
	helper := NewDatabaseTestHelper()
	if helper == nil {
		t.Fatal("NewDatabaseTestHelper returned nil")
	}

	// Test creating different types of databases
	minimalDB := helper.CreateMinimalDatabase()
	if minimalDB == nil || len(minimalDB.Commands) == 0 {
		t.Error("Expected minimal database with commands")
	}

	largeDB := helper.CreateLargeDatabase()
	if largeDB == nil || len(largeDB.Commands) == 0 {
		t.Error("Expected large database with commands")
	}

	emptyDB := helper.CreateEmptyDatabase()
	if emptyDB == nil {
		t.Error("Expected empty database to be created")
	}
	if len(emptyDB.Commands) != 0 {
		t.Error("Expected empty database to have no commands")
	}
}

func TestFileHelper(t *testing.T) {
	fileHelper := NewFileHelper()
	if fileHelper == nil {
		t.Fatal("NewFileHelper returned nil")
	}

	// Test creating temp file
	tempFile := fileHelper.CreateTempFile("test content")
	if tempFile == "" {
		t.Error("Expected temp file path to be non-empty")
	}

	// Test creating temp dir
	tempDir := fileHelper.CreateTempDir()
	if tempDir == "" {
		t.Error("Expected temp dir path to be non-empty")
	}

	// Test cleanup (should not panic)
	fileHelper.Cleanup()
}

func TestPathHelper(t *testing.T) {
	pathHelper := NewPathHelper()
	if pathHelper == nil {
		t.Fatal("NewPathHelper returned nil")
	}

	// Test getting test data dir
	testDataDir := pathHelper.GetTestDataDir()
	if testDataDir == "" {
		t.Error("Expected test data dir to be non-empty")
	}

	// Test ensuring test data dir (should not panic)
	err := pathHelper.EnsureTestDataDir()
	if err != nil {
		t.Errorf("EnsureTestDataDir failed: %v", err)
	}
}
