package nlp

import (
	"github.com/Vedant9500/WTF/internal/utils"
)

func (pq *ProcessedQuery) getRelevantActions() []string {
	if len(pq.Actions) > 0 {
		// Limit to 2-3 most relevant actions to avoid noise
		actionLimit := utils.Min(len(pq.Actions), 3)
		return pq.Actions[:actionLimit]
	}
	return nil
}

func (pq *ProcessedQuery) getRelevantTargets() []string {
	if len(pq.Targets) > 0 {
		// Limit to 2-3 most relevant targets to avoid noise
		targetLimit := utils.Min(len(pq.Targets), 3)
		return pq.Targets[:targetLimit]
	}
	return nil
}

func (pq *ProcessedQuery) getIntentKeywords() []string {
	switch pq.Intent {
	case IntentFind:
		return []string{"search", "find"}
	case IntentView:
		return []string{"cat", "view", "show"}
	case IntentCreate:
		return []string{"create", "make"}
	case IntentDelete:
		return []string{"delete", "remove"}
	case IntentInstall:
		return []string{"install", "setup"}
	case IntentModify:
		return []string{"configure", "change"}
	}
	return nil
}
