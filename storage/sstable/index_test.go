package sstable

import (
	"os"
	"testing"
)

func TestIndexAdd(t *testing.T) {
	index := NewIndex()

	index.Add("city", 0)
	index.Add("course", 12)
	index.Add("name", 25)

	if len(index.Entries) != 3 {
		t.Fatalf("expected 3 index entries, got %d", len(index.Entries))
	}

	if index.Entries[0].Key != "city" {
		t.Fatalf("expected city, got %s", index.Entries[0].Key)
	}

	if index.Entries[1].Offset != 12 {
		t.Fatalf("expected offset 12, got %d", index.Entries[1].Offset)
	}
}

func TestBuildIndex(t *testing.T) {
	file := "test_index.sst"

	defer os.Remove(file)

	entries := []Entry{
		{Key: "name", Value: "Yaswanthi"},
		{Key: "city", Value: "Guntur"},
		{Key: "course", Value: "MTech"},
	}

	err := Write(file, entries)
	if err != nil {
		t.Fatalf("failed to write SSTable: %v", err)
	}

	index, err := BuildIndex(file)
	if err != nil {
		t.Fatalf("failed to build index: %v", err)
	}

	if len(index.Entries) != 3 {
		t.Fatalf("expected 3 index entries, got %d", len(index.Entries))
	}
}

func TestGetWithIndex(t *testing.T) {
	file := "test_index_get.sst"

	defer os.Remove(file)

	entries := []Entry{
		{Key: "name", Value: "Yaswanthi"},
		{Key: "city", Value: "Guntur"},
		{Key: "course", Value: "MTech"},
	}

	err := Write(file, entries)
	if err != nil {
		t.Fatalf("failed to write SSTable: %v", err)
	}

	index, err := BuildIndex(file)
	if err != nil {
		t.Fatalf("failed to build index: %v", err)
	}

	value, exists, tombstone, err := GetWithIndex(file, index, "city")

	if err != nil {
		t.Fatalf("failed to get value: %v", err)
	}

	if tombstone {
		t.Fatal("expected key to not be a tombstone")
	}

	if !exists {
		t.Fatal("expected city to exist")
	}

	if value != "Guntur" {
		t.Fatalf("expected Guntur, got %s", value)
	}
}
