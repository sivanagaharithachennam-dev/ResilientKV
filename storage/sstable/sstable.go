package sstable

import (
	"bufio"
	"os"
	"sort"
	"strings"
)

// Entry represents one key-value pair or tombstone.
type Entry struct {
	Key       string
	Value     string
	Tombstone bool
}

// Write creates an SSTable from entries.
func Write(path string, entries []Entry) error {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Key < entries[j].Key
	})

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, entry := range entries {
		var line string

		if entry.Tombstone {
			line = "DELETE:" + entry.Key + "\n"
		} else {
			line = entry.Key + "=" + entry.Value + "\n"
		}

		if _, err := writer.WriteString(line); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// ReadAll reads all entries from an SSTable.
func ReadAll(path string) ([]Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var entries []Entry

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "DELETE:") {
			entries = append(entries, Entry{
				Key:       strings.TrimPrefix(line, "DELETE:"),
				Tombstone: true,
			})
			continue
		}

		parts := strings.SplitN(line, "=", 2)

		if len(parts) != 2 {
			continue
		}

		entries = append(entries, Entry{
			Key:   parts[0],
			Value: parts[1],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// Get searches for a key in the SSTable.
//
// Returns:
// value     - stored value
// exists    - whether the key was found
// tombstone - whether the key was deleted
func Get(path string, key string) (string, bool, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", false, false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "DELETE:") {
			deletedKey := strings.TrimPrefix(line, "DELETE:")

			if deletedKey == key {
				return "", true, true, nil
			}

			continue
		}

		parts := strings.SplitN(line, "=", 2)

		if len(parts) != 2 {
			continue
		}

		if parts[0] == key {
			return parts[1], true, false, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", false, false, err
	}

	return "", false, false, nil
}

// BuildIndex creates an index containing the file offset of every key.
func BuildIndex(path string) (*Index, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	index := NewIndex()

	var offset int64

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "DELETE:") {
			key := strings.TrimPrefix(line, "DELETE:")
			index.Add(key, offset)
		} else {
			parts := strings.SplitN(line, "=", 2)

			if len(parts) == 2 {
				index.Add(parts[0], offset)
			}
		}

		offset += int64(len(line)) + 1
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return index, nil
}

// GetWithIndex searches an SSTable using its index.
func GetWithIndex(path string, index *Index, key string) (string, bool, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", false, false, err
	}
	defer file.Close()

	var offset int64 = -1

	for _, entry := range index.Entries {
		if entry.Key == key {
			offset = entry.Offset
			break
		}
	}

	if offset == -1 {
		return "", false, false, nil
	}

	_, err = file.Seek(offset, 0)
	if err != nil {
		return "", false, false, err
	}

	reader := bufio.NewReader(file)

	line, err := reader.ReadString('\n')

	if err != nil && len(line) == 0 {
		return "", false, false, err
	}

	line = strings.TrimSpace(line)

	if strings.HasPrefix(line, "DELETE:") {
		deletedKey := strings.TrimPrefix(line, "DELETE:")

		if deletedKey == key {
			return "", true, true, nil
		}

		return "", false, false, nil
	}

	parts := strings.SplitN(line, "=", 2)

	if len(parts) != 2 {
		return "", false, false, nil
	}

	if parts[0] != key {
		return "", false, false, nil
	}

	return parts[1], true, false, nil
}
