package integration

import (
	"bytes"
	"kv-storage/internal/store"
	"os"
	"testing"
)

func TestStore(t *testing.T) {
	db, err := store.NewStore("test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove("test.db")

	db.Set("test", []byte("test"))

	entry, err := db.Get("test")
	if err != nil {
		t.Fatal(err)
	}
	if entry.Key != "test" || !bytes.Equal(entry.Value, []byte("test")) {
		t.Errorf("expected key/value 'test', got %v", entry)
	}
}

func TestStoreCorrupted(t *testing.T) {
	filename := "test_corrupt.db"
	db, err := store.NewStore(filename)
	if err != nil {
		t.Fatal(err)
	}

	db.Set("test", []byte("test"))

	db.Close()

	err = os.Truncate(filename, 0)
	if err != nil {
		t.Fatal(err)
	}

	db, err = store.NewStore(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove(filename)
	_, err = db.Get("test")

	if err == nil {
		t.Error("expected error after corruption, got nil")
	}
}
