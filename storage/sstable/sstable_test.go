package sstable

import (
	"os"
	"testing"
)

func TestSSTableWriteAndRead(t *testing.T) {
	file := "test.sst"

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

	result, err := ReadAll(file)

	if err != nil {
		t.Fatalf("failed to read SSTable: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(result))
	}

	if result[0].Key != "city" {
		t.Fatalf("expected first key to be city, got %s", result[0].Key)
	}

	if result[1].Key != "course" {
		t.Fatalf("expected second key to be course, got %s", result[1].Key)
	}

	if result[2].Key != "name" {
		t.Fatalf("expected third key to be name, got %s", result[2].Key)
	}
}

func TestSSTableGet(t *testing.T) {
	file := "test_get.sst"

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

	value, exists, tombstone, err := Get(file, "city")

	if err != nil {
		t.Fatalf("failed to get key: %v", err)
	}

	if tombstone {
		t.Fatal("expected key to not be a tombstone")
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

func TestSSTableGetMissingKey(t *testing.T) {
	file := "test_missing.sst"

	defer os.Remove(file)

	entries := []Entry{
		{Key: "name", Value: "Yaswanthi"},
	}

	err := Write(file, entries)

	if err != nil {
		t.Fatalf("failed to write SSTable: %v", err)
	}

	_, exists, _, err := Get(file, "age")

	if err != nil {
		t.Fatalf("failed to search SSTable: %v", err)
	}

	if exists {
		t.Fatal("expected age not to exist")
	}
}
