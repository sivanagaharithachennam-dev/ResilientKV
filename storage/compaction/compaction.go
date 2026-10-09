package compaction

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"resilientkv/storage/sstable"
)

// Compact merges all SSTables in a directory into one SSTable.
//
// Newer SSTables have higher numeric IDs.
// If the same key appears multiple times, the newest entry wins.
func Compact(dataDir string, outputID int) (string, error) {
	files, err := os.ReadDir(dataDir)
	if err != nil {
		return "", err
	}

	type tableFile struct {
		path string
		id   int
	}

	var tables []tableFile

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(file.Name()) != ".sst" {
			continue
		}

		id, err := extractSSTableID(file.Name())
		if err != nil {
			continue
		}

		tables = append(tables, tableFile{
			path: filepath.Join(dataDir, file.Name()),
			id:   id,
		})
	}

	if len(tables) == 0 {
		return "", fmt.Errorf("no SSTables found")
	}

	// Oldest first.
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].id < tables[j].id
	})

	// Later entries overwrite earlier entries.
	latest := make(map[string]sstable.Entry)

	for _, table := range tables {
		entries, err := sstable.ReadAll(table.path)
		if err != nil {
			return "", err
		}

		for _, entry := range entries {
			latest[entry.Key] = entry
		}
	}

	// Convert map to deterministic slice.
	keys := make([]string, 0, len(latest))

	for key := range latest {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	merged := make([]sstable.Entry, 0, len(keys))

	for _, key := range keys {
		merged = append(merged, latest[key])
	}

	outputPath := filepath.Join(
		dataDir,
		fmt.Sprintf("sstable-%06d.sst", outputID),
	)

	if err := sstable.Write(outputPath, merged); err != nil {
		return "", err
	}

	// Remove old SSTables, but keep the newly created output file.
	for _, table := range tables {
		if table.path == outputPath {
			continue
		}

		if err := os.Remove(table.path); err != nil {
			return "", err
		}
	}

	return outputPath, nil
}

// extractSSTableID extracts the numeric ID from a filename such as:
//
// sstable-000123.sst
func extractSSTableID(filename string) (int, error) {
	if !strings.HasPrefix(filename, "sstable-") {
		return 0, fmt.Errorf("invalid SSTable filename")
	}

	if filepath.Ext(filename) != ".sst" {
		return 0, fmt.Errorf("invalid SSTable extension")
	}

	number := strings.TrimSuffix(
		strings.TrimPrefix(filename, "sstable-"),
		".sst",
	)

	id, err := strconv.Atoi(number)
	if err != nil {
		return 0, err
	}

	return id, nil
}
