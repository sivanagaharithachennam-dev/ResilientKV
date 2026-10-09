package wal

import (
	"os"
	"testing"
)

func TestWALAppendAndRead(t *testing.T) {
	file := "test_wal.log"

	defer os.Remove(file)

	w, err := New(file)

	if err != nil {
		t.Fatalf("failed to create WAL: %v", err)
	}

	defer w.Close()

	err = w.Append("PUT name Yaswanthi")

	if err != nil {
		t.Fatalf("failed to append: %v", err)
	}

	err = w.Append("PUT city Guntur")

	if err != nil {
		t.Fatalf("failed to append: %v", err)
	}

	operations, err := w.ReadAll()

	if err != nil {
		t.Fatalf("failed to read WAL: %v", err)
	}

	if len(operations) != 2 {
		t.Fatalf("expected 2 operations, got %d", len(operations))
	}

	if operations[0] != "PUT name Yaswanthi" {
		t.Fatalf("unexpected first operation: %s", operations[0])
	}

	if operations[1] != "PUT city Guntur" {
		t.Fatalf("unexpected second operation: %s", operations[1])
	}
}
