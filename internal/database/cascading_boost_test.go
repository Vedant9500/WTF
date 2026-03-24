package database

import (
	"testing"

	"github.com/Vedant9500/WTF/internal/constants"
	"github.com/Vedant9500/WTF/internal/nlp"
)

func TestCalculateBoostForCommand_GatesSingleWeakSignal(t *testing.T) {
	db := &Database{}
	cmd := &Command{
		Command:     "printf hello",
		Description: "print text",
		Keywords:    []string{"print", "text"},
	}
	ctx := boostContext{
		keywordTerms: []string{"print"},
		intent:       nlp.IntentGeneral,
	}

	boost := db.calculateBoostForCommand(cmd, ctx)
	if boost != 1.0 {
		t.Fatalf("expected no boost for single weak signal, got %f", boost)
	}
}

func TestCalculateBoostForCommand_AllowsSingleHintSignal(t *testing.T) {
	db := &Database{}
	cmd := &Command{
		Command:     "mkdir -p logs",
		Description: "create directories",
		Keywords:    []string{"directory", "create"},
	}
	ctx := boostContext{
		commandHints: []string{"mkdir"},
		intent:       nlp.IntentCreate,
	}

	boost := db.calculateBoostForCommand(cmd, ctx)
	if boost <= 1.0 {
		t.Fatalf("expected hint-based boost to apply, got %f", boost)
	}
}

func TestCalculateBoostForCommand_CapsMaxMultiplier(t *testing.T) {
	db := &Database{}
	cmd := &Command{
		Command:     "mkdir -p logs",
		Description: "create logs directory and setup files",
		Keywords:    []string{"create", "directory", "logs", "setup"},
	}
	ctx := boostContext{
		actionTerms:  []string{"create"},
		targetTerms:  []string{"directory"},
		keywordTerms: []string{"logs"},
		commandHints: []string{"mkdir"},
		contexts:     []string{"mkdir"},
		intent:       nlp.IntentCreate,
	}

	boost := db.calculateBoostForCommand(cmd, ctx)
	if boost > constants.CascadingMaxMultiplier {
		t.Fatalf("expected capped boost <= %f, got %f", constants.CascadingMaxMultiplier, boost)
	}
	if boost != constants.CascadingMaxMultiplier {
		t.Fatalf("expected boost to hit cap %f for dense signal match, got %f", constants.CascadingMaxMultiplier, boost)
	}
}

func TestShouldApplyCascadingBoost_MinSignals(t *testing.T) {
	if shouldApplyCascadingBoost(boostMatches{keyword: true}) {
		t.Fatalf("expected single keyword signal to be rejected")
	}

	if !shouldApplyCascadingBoost(boostMatches{keyword: true, target: true}) {
		t.Fatalf("expected two signals to be accepted")
	}
}
