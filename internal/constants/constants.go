// Package constants defines application-wide constants.
package constants

import "time"

// Search scoring constants
const (
	ScoreDirectCommandMatch  = 15.0
	ScoreCommandMatch        = 10.0
	ScoreDescriptionMatch    = 6.0
	ScoreKeywordExactMatch   = 4.0
	ScoreKeywordPartialMatch = 1.0
	ScoreDomainSpecificMatch = 12.0

	// Intent boost multipliers
	IntentBoostMultiplier  = 2.0
	ActionBoostExact       = 1.5
	ActionBoostDescription = 1.3
	TargetBoostExact       = 1.4
	TargetBoostDescription = 1.2

	// Category boost multipliers
	CategoryBoostCompression        = 1.5
	CategoryBoostDirectory          = 1.5
	CategoryBoostSearch             = 1.3
	CategoryBoostDownload           = 1.4
	CategoryBoostSpecialCompression = 2.5
	CategoryBoostSearchPenalty      = 0.2
)

// Search defaults
const (
	DefaultSearchLimit    = 5
	DefaultFuzzyThreshold = -30
	DefaultMaxResults     = 5
	DefaultHistorySize    = 100
	
	// Search algorithm constants
	CrossPlatformPenalty     = 0.9
	ResultsBufferMultiplier  = 3
	FuzzySearchMultiplier    = 2
	MinWordLength           = 2
	FuzzyScoreThreshold     = 0.5
	FuzzyNormalizationBase  = 100.0
	NicheBoostFactor        = 0.2
	
	// Word scoring constants
	DirectCommandMatchScore = 15.0
	CommandMatchScore      = 10.0
	DomainSpecificScore    = 12.0
	DescriptionMatchScore  = 6.0
	KeywordExactScore      = 4.0
	KeywordPartialScore    = 1.0
	
	// Suggestion constants
	DefaultMaxSuggestions  = 3
	FuzzySuggestionThreshold = -20
)

// Cache settings
const (
	DefaultCacheTTL = 5 * time.Minute
)

// File size limits
const (
	MaxQueryLength = 1000 // Maximum query length in characters
)

// NLP processing constants
const (
	MaxSynonymsPerWord = 1 // Only use the best synonym to avoid query expansion explosion
	StopWordThreshold  = 2 // Minimum word length to avoid stop word filtering
)
