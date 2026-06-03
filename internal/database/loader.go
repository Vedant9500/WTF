package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Vedant9500/WTF/internal/errors"
	"github.com/Vedant9500/WTF/internal/nlp"

	"gopkg.in/yaml.v3"
)

var dbLoadCache = struct {
	sync.RWMutex
	items map[string]*Database
}{items: make(map[string]*Database)}

// LoadDatabase loads commands from a YAML file and returns a populated Database.
//
// This function reads a YAML file containing an array of Command objects,
// parses the YAML content, and populates performance-optimized cache fields
// for faster search operations. The cache fields include lowercased versions
// of all searchable text fields.
//
// Parameters:
//   - filename: Path to the YAML file containing command definitions
//
// Returns:
//   - *Database: Populated database with all commands and cache fields
//   - error: Database loading error with user-friendly context
//
// The function handles various error conditions:
//   - File not found errors
//   - Permission denied errors
//   - YAML parsing errors
//   - Invalid command structure errors
func LoadDatabase(filename string) (*Database, error) {
	if db, ok := getCachedDatabase(filename, ""); ok {
		return db, nil
	}

	commands, err := loadCommandsFromFile(filename)
	if err != nil {
		return nil, err
	}

	db := newDatabase(commands)
	if err := db.LoadEmbeddings(); err != nil {
		return nil, err
	}

	setCachedDatabase(filename, "", db)
	return db, nil
}

func loadCommandsFromFile(filename string) ([]Command, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.NewDatabaseErrorWithContext("read", filename, err)
	}

	var commands []Command
	if err := yaml.Unmarshal(data, &commands); err != nil {
		return nil, errors.NewDatabaseErrorWithContext("parse", filename, err)
	}

	populateLowercasedCommandFields(commands)
	return commands, nil
}

func populateLowercasedCommandFields(commands []Command) {
	// Populate lowercased cache fields for performance
	for i := range commands {
		commands[i].CommandLower = strings.ToLower(commands[i].Command)
		commands[i].DescriptionLower = strings.ToLower(commands[i].Description)
		commands[i].KeywordsLower = make([]string, len(commands[i].Keywords))
		for j, kw := range commands[i].Keywords {
			commands[i].KeywordsLower[j] = strings.ToLower(kw)
		}
		commands[i].TagsLower = make([]string, len(commands[i].Tags))
		for j, tag := range commands[i].Tags {
			commands[i].TagsLower[j] = strings.ToLower(tag)
		}
	}
}

func newDatabase(commands []Command) *Database {
	db := &Database{Commands: commands}
	// Build universal index for scalable search
	db.BuildUniversalIndex()
	// Build command index for fast pointer lookups; TF-IDF is built lazily when NLP reranking is requested.
	db.buildCommandIndex()
	// Initialize NLP processor for enhanced query processing
	db.nlpProcessor = nlp.NewQueryProcessor()
	return db
}

// LoadDatabaseWithPersonal loads both main and personal database files and merges them.
//
// This function loads the main command database and optionally merges it with
// a personal database file containing user-specific commands. The personal
// database is optional - if it doesn't exist, only the main database is loaded.
//
// Parameters:
//   - mainDBPath: Path to the main command database file (required)
//   - personalDBPath: Path to the personal command database file (optional)
//
// Returns:
//   - *Database: Combined database with commands from both files
//   - error: Loading error if the main database fails to load
//
// Behavior:
//   - Main database loading failure results in an error
//   - Personal database not found is silently ignored
//   - Personal database parsing errors are reported
//   - Commands from personal database are appended to main database commands
func LoadDatabaseWithPersonal(mainDBPath, personalDBPath string) (*Database, error) {
	if db, ok := getCachedDatabase(mainDBPath, personalDBPath); ok {
		return db, nil
	}

	// Load main database
	mainCommands, err := loadCommandsFromFile(mainDBPath)
	if err != nil {
		return nil, err
	}

	// Try to load personal database (it's OK if it doesn't exist)
	personalCommands, err := loadCommandsFromFile(personalDBPath)
	if err != nil {
		// If personal database doesn't exist, that's fine - just use main database
		if os.IsNotExist(err) {
			db := newDatabase(mainCommands)
			if e := db.LoadEmbeddings(); e != nil {
				return nil, e
			}
			setCachedDatabase(mainDBPath, personalDBPath, db)
			return db, nil
		}
		// Check if it's a DatabaseError wrapping IsNotExist
		if dbErr, ok := err.(*errors.DatabaseError); ok {
			if os.IsNotExist(dbErr.Cause) {
				db := newDatabase(mainCommands)
				if e := db.LoadEmbeddings(); e != nil {
					return nil, e
				}
				setCachedDatabase(mainDBPath, personalDBPath, db)
				return db, nil
			}
		}
		// Check if it's an AppError wrapping IsNotExist
		if appErr, ok := err.(*errors.AppError); ok {
			if appErr.Cause != nil && os.IsNotExist(appErr.Cause) {
				db := newDatabase(mainCommands)
				if e := db.LoadEmbeddings(); e != nil {
					return nil, e
				}
				setCachedDatabase(mainDBPath, personalDBPath, db)
				return db, nil
			}
		}
		// Other errors should be reported
		return nil, err
	}

	// Merge commands from both databases
	allCommands := make([]Command, 0, len(mainCommands)+len(personalCommands))
	allCommands = append(allCommands, mainCommands...)
	allCommands = append(allCommands, personalCommands...)

	db := newDatabase(allCommands)
	if err := db.LoadEmbeddings(); err != nil {
		return nil, err
	}
	setCachedDatabase(mainDBPath, personalDBPath, db)
	return db, nil
}

func getCachedDatabase(mainPath, personalPath string) (*Database, bool) {
	key, ok := databaseCacheKey(mainPath, personalPath)
	if !ok {
		return nil, false
	}
	dbLoadCache.RLock()
	defer dbLoadCache.RUnlock()
	db, ok := dbLoadCache.items[key]
	if !ok {
		return nil, false
	}
	return db.cloneForCacheHit(), true
}

func setCachedDatabase(mainPath, personalPath string, db *Database) {
	key, ok := databaseCacheKey(mainPath, personalPath)
	if !ok || db == nil {
		return
	}
	dbLoadCache.Lock()
	defer dbLoadCache.Unlock()
	dbLoadCache.items[key] = db
}

func databaseCacheKey(mainPath, personalPath string) (string, bool) {
	mainKey, ok := fileIdentity(mainPath, false)
	if !ok {
		return "", false
	}
	personalKey, ok := fileIdentity(personalPath, true)
	if !ok {
		return "", false
	}
	return mainKey + "|" + personalKey, true
}

func fileIdentity(path string, optional bool) (string, bool) {
	if path == "" {
		if optional {
			return "<none>", true
		}
		return "", false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", false
	}
	info, err := os.Stat(abs)
	if err != nil {
		if optional && os.IsNotExist(err) {
			return abs + ":missing", true
		}
		return "", false
	}
	return fmt.Sprintf("%s:%s:%d", abs, info.ModTime().UTC().Format("20060102150405.000000000"), info.Size()), true
}

func (db *Database) cloneForCacheHit() *Database {
	if db == nil {
		return nil
	}
	clone := &Database{
		Commands:               append([]Command(nil), db.Commands...),
		uIndex:                 db.uIndex,
		tfidf:                  db.tfidf,
		embeddingIndex:         db.embeddingIndex,
		enhancedEmbeddingIndex: db.enhancedEmbeddingIndex,
		sentenceIndex:          db.sentenceIndex,
		queryEmbedIndex:        db.queryEmbedIndex,
		embeddingSearcher:      db.embeddingSearcher,
		familyPriorIndex:       db.familyPriorIndex,
		nlpProcessor:           db.nlpProcessor,
	}
	clone.buildCommandIndex()
	return clone
}

// Size returns the total number of commands in the database.
//
// This method provides a quick way to get the count of all loaded commands,
// which is useful for metrics, debugging, and capacity planning.
func (db *Database) Size() int {
	return len(db.Commands)
}
