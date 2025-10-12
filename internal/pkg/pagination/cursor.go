package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// Cursor represents a pagination cursor
type Cursor struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
}

// EncodeCursor encodes a cursor to a base64 string
func EncodeCursor(cursor Cursor) (string, error) {
	data, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor: %w", err)
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// DecodeCursor decodes a base64 cursor string
func DecodeCursor(encoded string) (Cursor, error) {
	var cursor Cursor
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return cursor, fmt.Errorf("failed to decode cursor: %w", err)
	}

	if err := json.Unmarshal(data, &cursor); err != nil {
		return cursor, fmt.Errorf("failed to unmarshal cursor: %w", err)
	}

	return cursor, nil
}
