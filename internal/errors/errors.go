package errors

import "fmt"

// DatabaseError represents database-related errors
type DatabaseError struct {
	Path  string
	Op    string
	Cause error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database %s failed for '%s': %v", e.Op, e.Path, e.Cause)
}

// Unwrap returns the underlying error for error chain support
func (e *DatabaseError) Unwrap() error {
	return e.Cause
}

// NewDatabaseError creates a new database error
func NewDatabaseError(op, path string, cause error) *DatabaseError {
	return &DatabaseError{
		Op:    op,
		Path:  path,
		Cause: cause,
	}
}

// SearchError represents search-related errors
type SearchError struct {
	Query string
	Cause error
}

func (e *SearchError) Error() string {
	return fmt.Sprintf("search failed for query '%s': %v", e.Query, e.Cause)
}

// Unwrap returns the underlying error for error chain support
func (e *SearchError) Unwrap() error {
	return e.Cause
}

// NewSearchError creates a new search error
func NewSearchError(query string, cause error) *SearchError {
	return &SearchError{
		Query: query,
		Cause: cause,
	}
}
