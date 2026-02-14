package integration

import (
	"bytes"
	"kv-storage/internal/entry"
	"testing"
)

func TestEntry(t *testing.T) {
	original := entry.Entry{
		Key:   "test",
		Value: []byte("test"),
	}
	decoded := entry.Entry{}

	encoded := original.Encode()
	decoded.Decode(encoded)

	if original.Key != decoded.Key || !bytes.Equal(original.Value, decoded.Value) {
		t.Errorf("expected %v, got %v", original, decoded)
	}
}
