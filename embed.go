package echoi18n

import (
	"embed"
)

// EmbedLoader loads message files from an embedded filesystem.
type EmbedLoader struct {
	FS embed.FS // Embedded filesystem.
}

// LoadMessage retrieves a file from the embedded filesystem.
// Returns the file content or an error.
func (e *EmbedLoader) LoadMessage(path string) ([]byte, error) {
	return e.FS.ReadFile(path)
}
