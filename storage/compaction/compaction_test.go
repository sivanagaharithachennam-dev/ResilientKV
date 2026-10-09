package compaction

import (
	"os"
	"path/filepath"
	"testing"

	"resilientkv/storage/sstable"
)

func TestCompactKeepsNewestValue(t *testing.T) {
	dataDir := t.TempDir()

	err := sstable.Write(
		filepath.Join(dataDir, "sstable-000001.sst"),
		[]sstable.Entry{
			{
				Key:   "name",
				Value: "Old",
			},
		},
	)

	if err != nil {
		t.Fatalf("failed to create first SSTable: %v", err)
	}

	err = sstable.Write(
		filepath.Join(dataDir, "sstable-000002.sst"),
		[]sstable.Entry{
			{
				Key:   "name",
				Value: "New",
			},
		},
	)

	if err != nil {
		t.Fatalf("failed to create second SSTable: %v", err)
	}

	output, err := Compact(dataDir, 3)

	if err != nil {
		t.Fatalf("compaction failed: %v", err)
	}

	value, exists, tombstone, err :=
		sstable.Get(output, "name")

	if err != nil {
		t.Fatalf("failed to read compacted SSTable: %v", err)
	}

	if !exists {
		t.Fatal("expected name to exist")
	}

	if tombstone {
		t.Fatal("expected name not to be a tombstone")
	}

	if value != "New" {
		t.Fatalf(
			"expected New, got %s",
			value,
		)
	}
}

func TestCompactPreservesTombstone(t *testing.T) {
	dataDir := t.TempDir()

	err := sstable.Write(
		filepath.Join(dataDir, "sstable-000001.sst"),
		[]sstable.Entry{
			{
				Key:   "user",
				Value: "Yaswanthi",
			},
		},
	)

	if err != nil {
		t.Fatalf("failed to create first SSTable: %v", err)
	}

	err = sstable.Write(
		filepath.Join(dataDir, "sstable-000002.sst"),
		[]sstable.Entry{
			{
				Key:       "user",
				Tombstone: true,
			},
		},
	)

	if err != nil {
		t.Fatalf("failed to create second SSTable: %v", err)
	}

	output, err := Compact(dataDir, 3)

	if err != nil {
		t.Fatalf("compaction failed: %v", err)
	}

	_, exists, tombstone, err :=
		sstable.Get(output, "user")

	if err != nil {
		t.Fatalf("failed to read compacted SSTable: %v", err)
	}

	if !exists {
		t.Fatal("expected tombstone entry to exist")
	}

	if !tombstone {
		t.Fatal("expected user to remain a tombstone")
	}
}

func TestCompactRemovesOldSSTables(t *testing.T) {
	dataDir := t.TempDir()

	for i := 1; i <= 3; i++ {
		path := filepath.Join(
			dataDir,
			"sstable-"+formatID(i)+".sst",
		)

		err := sstable.Write(
			path,
			[]sstable.Entry{
				{
					Key:   "key",
					Value: "value",
				},
			},
		)

		if err != nil {
			t.Fatalf(
				"failed to create SSTable %d: %v",
				i,
				err,
			)
		}
	}

	output, err := Compact(dataDir, 4)

	if err != nil {
		t.Fatalf("compaction failed: %v", err)
	}

	if _, err := os.Stat(output); err != nil {
		t.Fatalf("compacted SSTable does not exist: %v", err)
	}

	for i := 1; i <= 3; i++ {
		path := filepath.Join(
			dataDir,
			"sstable-"+formatID(i)+".sst",
		)

		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf(
				"expected old SSTable to be removed: %s",
				path,
			)
		}
	}
}

func formatID(id int) string {
	return formatSixDigits(id)
}

func formatSixDigits(id int) string {
	if id < 10 {
		return "00000" + string(rune('0'+id))
	}

	if id < 100 {
		return "0000" + string(rune('0'+id/10)) +
			string(rune('0'+id%10))
	}

	// The test only uses IDs 1–3.
	return "000001"
}
