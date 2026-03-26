package database

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func (db *Database) commandSnapshotHash(limit int) string {
	if db == nil || len(db.Commands) == 0 {
		return ""
	}
	if limit <= 0 || limit > len(db.Commands) {
		limit = len(db.Commands)
	}

	h := sha256.New()
	for i := 0; i < limit; i++ {
		cmd := db.Commands[i]
		writeNormalized(h, cmd.Command)
		h.Write([]byte{0x1f})
		writeNormalized(h, cmd.Description)
		h.Write([]byte{0x1f})

		for j, kw := range cmd.Keywords {
			if j > 0 {
				h.Write([]byte{0x1e})
			}
			writeNormalized(h, kw)
		}
		h.Write([]byte{0x1d})
	}

	return hex.EncodeToString(h.Sum(nil))
}

func writeNormalized(h interface{ Write([]byte) (int, error) }, value string) {
	normalized := strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
	_, _ = h.Write([]byte(normalized))
}
