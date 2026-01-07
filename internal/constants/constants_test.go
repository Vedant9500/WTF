package constants

import (
	"testing"
	"time"
)

func TestSearchScoringConstants(t *testing.T) {
	// Test that scoring constants are reasonable values
	testCases := []struct {
		name     string
		value    float64
		minValue float64
		maxValue float64
	}{
		{"ScoreDirectCommandMatch", ScoreDirectCommandMatch, 10.0, 20.0},
		{"ScoreCommandMatch", ScoreCommandMatch, 5.0, 15.0},
		{"ScoreDescriptionMatch", ScoreDescriptionMatch, 3.0, 10.0},
		{"ScoreKeywordExactMatch", ScoreKeywordExactMatch, 2.0, 8.0},
		{"ScoreKeywordPartialMatch", ScoreKeywordPartialMatch, 0.5, 3.0},
		{"ScoreDomainSpecificMatch", ScoreDomainSpecificMatch, 8.0, 15.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value < tc.minValue || tc.value > tc.maxValue {
				t.Errorf("%s = %f, expected between %f and %f",
					tc.name, tc.value, tc.minValue, tc.maxValue)
			}
		})
	}
}

func TestIntentBoostConstants(t *testing.T) {
	// Test that intent boost multipliers are reasonable
	testCases := []struct {
		name     string
		value    float64
		minValue float64
		maxValue float64
	}{
		{"IntentBoostMultiplier", IntentBoostMultiplier, 1.5, 3.0},
		{"ActionBoostExact", ActionBoostExact, 1.2, 2.0},
		{"ActionBoostDescription", ActionBoostDescription, 1.1, 1.8},
		{"TargetBoostExact", TargetBoostExact, 1.2, 2.0},
		{"TargetBoostDescription", TargetBoostDescription, 1.1, 1.8},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value < tc.minValue || tc.value > tc.maxValue {
				t.Errorf("%s = %f, expected between %f and %f",
					tc.name, tc.value, tc.minValue, tc.maxValue)
			}
		})
	}
}

func TestCategoryBoostConstants(t *testing.T) {
	// Test category boost multipliers
	testCases := []struct {
		name     string
		value    float64
		minValue float64
		maxValue float64
	}{
		{"CategoryBoostCompression", CategoryBoostCompression, 1.2, 2.0},
		{"CategoryBoostDirectory", CategoryBoostDirectory, 1.2, 2.0},
		{"CategoryBoostSearch", CategoryBoostSearch, 1.1, 1.8},
		{"CategoryBoostDownload", CategoryBoostDownload, 1.2, 2.0},
		{"CategoryBoostSpecialCompression", CategoryBoostSpecialCompression, 2.0, 3.0},
		{"CategoryBoostSearchPenalty", CategoryBoostSearchPenalty, 0.1, 0.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value < tc.minValue || tc.value > tc.maxValue {
				t.Errorf("%s = %f, expected between %f and %f",
					tc.name, tc.value, tc.minValue, tc.maxValue)
			}
		})
	}
}

func TestSearchDefaults(t *testing.T) {
	// Test search default values
	if DefaultSearchLimit <= 0 {
		t.Errorf("DefaultSearchLimit = %d, expected positive value", DefaultSearchLimit)
	}

	if DefaultSearchLimit > 20 {
		t.Errorf("DefaultSearchLimit = %d, expected reasonable limit (≤20)", DefaultSearchLimit)
	}

	if DefaultFuzzyThreshold >= 0 {
		t.Errorf("DefaultFuzzyThreshold = %d, expected negative value", DefaultFuzzyThreshold)
	}

	if DefaultMaxResults <= 0 {
		t.Errorf("DefaultMaxResults = %d, expected positive value", DefaultMaxResults)
	}

	if DefaultHistorySize <= 0 {
		t.Errorf("DefaultHistorySize = %d, expected positive value", DefaultHistorySize)
	}
}

func TestCacheSettings(t *testing.T) {
	// Test cache TTL
	if DefaultCacheTTL <= 0 {
		t.Errorf("DefaultCacheTTL = %v, expected positive duration", DefaultCacheTTL)
	}

	if DefaultCacheTTL > time.Hour {
		t.Errorf("DefaultCacheTTL = %v, expected reasonable duration (≤1 hour)", DefaultCacheTTL)
	}

	// Verify it's actually 5 minutes as expected
	expectedTTL := 5 * time.Minute
	if DefaultCacheTTL != expectedTTL {
		t.Errorf("DefaultCacheTTL = %v, expected %v", DefaultCacheTTL, expectedTTL)
	}
}

func TestFileSizeLimits(t *testing.T) {
	// Test query length limit
	if MaxQueryLength <= 0 {
		t.Errorf("MaxQueryLength = %d, expected positive value", MaxQueryLength)
	}

	if MaxQueryLength < 100 {
		t.Errorf("MaxQueryLength = %d, expected at least 100 characters", MaxQueryLength)
	}

	if MaxQueryLength > 10000 {
		t.Errorf("MaxQueryLength = %d, expected reasonable limit (≤10000)", MaxQueryLength)
	}

	// Test minimum word length
	if MinWordLength <= 0 {
		t.Errorf("MinWordLength = %d, expected positive value", MinWordLength)
	}

	if MinWordLength > 5 {
		t.Errorf("MinWordLength = %d, expected reasonable minimum (≤5)", MinWordLength)
	}
}

func TestNLPConstants(t *testing.T) {
	// Test NLP processing constants
	if MaxSynonymsPerWord < 0 {
		t.Errorf("MaxSynonymsPerWord = %d, expected non-negative value", MaxSynonymsPerWord)
	}

	if MaxSynonymsPerWord > 5 {
		t.Errorf("MaxSynonymsPerWord = %d, expected reasonable limit (≤5)", MaxSynonymsPerWord)
	}

	if StopWordThreshold <= 0 {
		t.Errorf("StopWordThreshold = %d, expected positive value", StopWordThreshold)
	}

	if StopWordThreshold > 10 {
		t.Errorf("StopWordThreshold = %d, expected reasonable threshold (≤10)", StopWordThreshold)
	}
}

func TestConstantRelationships(t *testing.T) {
	// Test that constants have logical relationships

	// Direct command match should be higher than regular command match
	if ScoreDirectCommandMatch <= ScoreCommandMatch {
		t.Errorf("ScoreDirectCommandMatch (%f) should be > ScoreCommandMatch (%f)",
			ScoreDirectCommandMatch, ScoreCommandMatch)
	}

	// Command match should be higher than description match
	if ScoreCommandMatch <= ScoreDescriptionMatch {
		t.Errorf("ScoreCommandMatch (%f) should be > ScoreDescriptionMatch (%f)",
			ScoreCommandMatch, ScoreDescriptionMatch)
	}

	// Exact keyword match should be higher than partial match
	if ScoreKeywordExactMatch <= ScoreKeywordPartialMatch {
		t.Errorf("ScoreKeywordExactMatch (%f) should be > ScoreKeywordPartialMatch (%f)",
			ScoreKeywordExactMatch, ScoreKeywordPartialMatch)
	}

	// Intent boost should be higher than action boost
	if IntentBoostMultiplier <= ActionBoostExact {
		t.Errorf("IntentBoostMultiplier (%f) should be > ActionBoostExact (%f)",
			IntentBoostMultiplier, ActionBoostExact)
	}

	// Action boost exact should be higher than description boost
	if ActionBoostExact <= ActionBoostDescription {
		t.Errorf("ActionBoostExact (%f) should be > ActionBoostDescription (%f)",
			ActionBoostExact, ActionBoostDescription)
	}
}

func TestConstantTypes(_ *testing.T) {
	// Test that constants are of expected types

	var i int
	var d time.Duration

	// Scoring constants should be float64
	_ = float64(ScoreDirectCommandMatch)
	_ = float64(ScoreCommandMatch)
	_ = float64(ScoreDescriptionMatch)
	_ = float64(ScoreKeywordExactMatch)
	_ = float64(ScoreKeywordPartialMatch)
	_ = float64(ScoreDomainSpecificMatch)

	// Search defaults should be int
	i = DefaultSearchLimit
	i = DefaultFuzzyThreshold
	i = DefaultMaxResults
	i = DefaultHistorySize
	i = MaxQueryLength
	i = MinWordLength
	i = MaxSynonymsPerWord
	i = StopWordThreshold
	_ = i

	// Cache TTL should be time.Duration
	d = DefaultCacheTTL
	_ = d
}

func TestConstantValues(t *testing.T) {
	// Test specific expected values
	if DefaultSearchLimit != 5 {
		t.Errorf("DefaultSearchLimit = %d, expected 5", DefaultSearchLimit)
	}

	if DefaultMaxResults != 5 {
		t.Errorf("DefaultMaxResults = %d, expected 5", DefaultMaxResults)
	}

	if DefaultHistorySize != 100 {
		t.Errorf("DefaultHistorySize = %d, expected 100", DefaultHistorySize)
	}

	if MaxQueryLength != 1000 {
		t.Errorf("MaxQueryLength = %d, expected 1000", MaxQueryLength)
	}

	if MinWordLength != 2 {
		t.Errorf("MinWordLength = %d, expected 2", MinWordLength)
	}

	if MaxSynonymsPerWord != 1 {
		t.Errorf("MaxSynonymsPerWord = %d, expected 1", MaxSynonymsPerWord)
	}

	if StopWordThreshold != 2 {
		t.Errorf("StopWordThreshold = %d, expected 2", StopWordThreshold)
	}
}
