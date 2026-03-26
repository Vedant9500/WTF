package nlp

import "strings"

var lemmaFixups = map[string]string{
	"configur": "configure",
	"delet":    "delete",
	"creat":    "create",
	"remov":    "remove",
	"modifi":   "modify",
	"updat":    "update",
	"generat":  "generate",
	"locat":    "locate",
	"runn":     "run",
	"stopp":    "stop",
	"mov":      "move",
	"copi":     "copy",
	"archiv":   "archive",
}

// normalizeToken applies conservative normalization to align common inflected
// forms with canonical keywords (e.g., "installing" -> "install").
func normalizeToken(token string) string {
	w := strings.ToLower(strings.TrimSpace(token))
	if len(w) < 4 {
		return w
	}

	if fixed, ok := lemmaFixups[w]; ok {
		return fixed
	}

	stem := w
	switch {
	case len(w) > 4 && strings.HasSuffix(w, "ies"):
		stem = w[:len(w)-3] + "y"
	case len(w) > 5 && strings.HasSuffix(w, "ing"):
		stem = trimTrailingDoubleConsonant(w[:len(w)-3])
	case len(w) > 4 && strings.HasSuffix(w, "ed"):
		stem = trimTrailingDoubleConsonant(w[:len(w)-2])
	case len(w) > 4 && strings.HasSuffix(w, "es"):
		stem = w[:len(w)-2]
	}

	if fixed, ok := lemmaFixups[stem]; ok {
		return fixed
	}

	return stem
}

func trimTrailingDoubleConsonant(s string) string {
	if len(s) < 2 {
		return s
	}
	last := s[len(s)-1]
	prev := s[len(s)-2]
	if last == prev && !isASCIIvowel(last) && last != 'l' {
		return s[:len(s)-1]
	}
	return s
}

func isASCIIvowel(b byte) bool {
	switch b {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	default:
		return false
	}
}
