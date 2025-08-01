package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestDatabaseError(t *testing.T) {
	cause := errors.New("file not found")
	dbErr := NewDatabaseError("load", "/path/to/db.yaml", cause)

	// Test error message format
	expectedMsg := "database load failed for '/path/to/db.yaml': file not found"
	if dbErr.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, dbErr.Error())
	}

	// Test fields
	if dbErr.Op != "load" {
		t.Errorf("Expected Op 'load', got '%s'", dbErr.Op)
	}

	if dbErr.Path != "/path/to/db.yaml" {
		t.Errorf("Expected Path '/path/to/db.yaml', got '%s'", dbErr.Path)
	}

	if dbErr.Cause != cause {
		t.Errorf("Expected Cause to be the original error")
	}
}

func TestDatabaseErrorUnwrap(t *testing.T) {
	cause := errors.New("original error")
	dbErr := NewDatabaseError("save", "/path", cause)

	// Test unwrapping
	unwrapped := dbErr.Unwrap()
	if unwrapped != cause {
		t.Errorf("Expected unwrapped error to be the original cause")
	}

	// Test with errors.Is
	if !errors.Is(dbErr, cause) {
		t.Error("Expected errors.Is to find the cause in the error chain")
	}
}

func TestSearchError(t *testing.T) {
	cause := errors.New("invalid query")
	searchErr := NewSearchError("test query", cause)

	// Test error message format
	expectedMsg := "search failed for query 'test query': invalid query"
	if searchErr.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, searchErr.Error())
	}

	// Test fields
	if searchErr.Query != "test query" {
		t.Errorf("Expected Query 'test query', got '%s'", searchErr.Query)
	}

	if searchErr.Cause != cause {
		t.Errorf("Expected Cause to be the original error")
	}
}

func TestSearchErrorUnwrap(t *testing.T) {
	cause := errors.New("original error")
	searchErr := NewSearchError("query", cause)

	// Test unwrapping
	unwrapped := searchErr.Unwrap()
	if unwrapped != cause {
		t.Errorf("Expected unwrapped error to be the original cause")
	}

	// Test with errors.Is
	if !errors.Is(searchErr, cause) {
		t.Error("Expected errors.Is to find the cause in the error chain")
	}
}

func TestErrorChaining(t *testing.T) {
	// Test error chaining with multiple levels
	originalErr := errors.New("root cause")
	dbErr := NewDatabaseError("load", "/path", originalErr)
	searchErr := NewSearchError("query", dbErr)

	// Test that we can find the original error through the chain
	if !errors.Is(searchErr, originalErr) {
		t.Error("Expected errors.Is to find the root cause through the error chain")
	}

	if !errors.Is(searchErr, dbErr) {
		t.Error("Expected errors.Is to find the database error in the chain")
	}
}

func TestErrorWithNilCause(t *testing.T) {
	// Test behavior with nil cause
	dbErr := NewDatabaseError("test", "/path", nil)
	
	expectedMsg := "database test failed for '/path': <nil>"
	if dbErr.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, dbErr.Error())
	}

	if dbErr.Unwrap() != nil {
		t.Error("Expected Unwrap() to return nil when cause is nil")
	}
}

func TestErrorWithEmptyFields(t *testing.T) {
	// Test with empty operation and path
	cause := errors.New("test error")
	dbErr := NewDatabaseError("", "", cause)

	expectedMsg := "database  failed for '': test error"
	if dbErr.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, dbErr.Error())
	}

	// Test with empty query
	searchErr := NewSearchError("", cause)
	expectedMsg = "search failed for query '': test error"
	if searchErr.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, searchErr.Error())
	}
}

func TestErrorFormatting(t *testing.T) {
	// Test error formatting with fmt verbs
	cause := errors.New("test cause")
	dbErr := NewDatabaseError("load", "/test/path", cause)

	// Test %s formatting
	formatted := fmt.Sprintf("%s", dbErr)
	expected := "database load failed for '/test/path': test cause"
	if formatted != expected {
		t.Errorf("Expected formatted string '%s', got '%s'", expected, formatted)
	}

	// Test %v formatting
	formatted = fmt.Sprintf("%v", dbErr)
	if formatted != expected {
		t.Errorf("Expected formatted string '%s', got '%s'", expected, formatted)
	}

	// Test %+v formatting (should be same as %v for our simple error types)
	formatted = fmt.Sprintf("%+v", dbErr)
	if formatted != expected {
		t.Errorf("Expected formatted string '%s', got '%s'", expected, formatted)
	}
}

func TestErrorTypeAssertion(t *testing.T) {
	cause := errors.New("test")
	dbErr := NewDatabaseError("load", "/path", cause)
	searchErr := NewSearchError("query", cause)

	// Test type assertions
	var err error

	err = dbErr
	if _, ok := err.(*DatabaseError); !ok {
		t.Error("Expected DatabaseError to be assertable to *DatabaseError")
	}

	err = searchErr
	if _, ok := err.(*SearchError); !ok {
		t.Error("Expected SearchError to be assertable to *SearchError")
	}

	// Test negative cases
	if _, ok := err.(*DatabaseError); ok {
		t.Error("Expected SearchError not to be assertable to *DatabaseError")
	}
}

func TestErrorEquality(t *testing.T) {
	cause1 := errors.New("cause1")
	cause2 := errors.New("cause2")

	dbErr1 := NewDatabaseError("load", "/path", cause1)
	dbErr2 := NewDatabaseError("load", "/path", cause1)
	dbErr3 := NewDatabaseError("save", "/path", cause1)
	dbErr4 := NewDatabaseError("load", "/other", cause1)
	dbErr5 := NewDatabaseError("load", "/path", cause2)

	// Test that errors with same content are not equal (different instances)
	if dbErr1 == dbErr2 {
		t.Error("Expected different error instances not to be equal")
	}

	// Test that errors.Is works correctly
	if !errors.Is(dbErr1, cause1) {
		t.Error("Expected errors.Is to find the cause")
	}

	if errors.Is(dbErr1, cause2) {
		t.Error("Expected errors.Is not to find different cause")
	}

	// Test field differences
	if dbErr1.Op == dbErr3.Op && dbErr1.Path == dbErr3.Path && dbErr1.Cause == dbErr3.Cause {
		t.Error("Expected dbErr3 to have different Op")
	}

	if dbErr1.Path == dbErr4.Path {
		t.Error("Expected dbErr4 to have different Path")
	}

	if dbErr1.Cause == dbErr5.Cause {
		t.Error("Expected dbErr5 to have different Cause")
	}
}

func TestComplexErrorScenarios(t *testing.T) {
	// Test realistic error scenarios
	
	// Database loading error
	fileErr := errors.New("permission denied")
	dbErr := NewDatabaseError("load", "/etc/wtf/commands.yaml", fileErr)
	
	// Search error wrapping database error
	searchErr := NewSearchError("git commit", dbErr)
	
	// Verify error messages
	expectedDbMsg := "database load failed for '/etc/wtf/commands.yaml': permission denied"
	if dbErr.Error() != expectedDbMsg {
		t.Errorf("Expected db error message '%s', got '%s'", expectedDbMsg, dbErr.Error())
	}
	
	expectedSearchMsg := "search failed for query 'git commit': " + expectedDbMsg
	if searchErr.Error() != expectedSearchMsg {
		t.Errorf("Expected search error message '%s', got '%s'", expectedSearchMsg, searchErr.Error())
	}
	
	// Verify error chain traversal
	if !errors.Is(searchErr, fileErr) {
		t.Error("Expected to find original file error in search error chain")
	}
	
	if !errors.Is(searchErr, dbErr) {
		t.Error("Expected to find database error in search error chain")
	}
}