package wal

import (
	"bufio"
	"os"
	"sync"
)

// WAL represents a Write-Ahead Log.
type WAL struct {
	mu   sync.Mutex
	file *os.File
}

// New creates or opens a WAL file.
func New(path string) (*WAL, error) {
	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_APPEND|os.O_RDWR,
		0644,
	)

	if err != nil {
		return nil, err
	}

	return &WAL{
		file: file,
	}, nil
}

// Append writes an operation to the WAL.
func (w *WAL) Append(operation string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	_, err := w.file.WriteString(operation + "\n")

	if err != nil {
		return err
	}

	return w.file.Sync()
}

// ReadAll reads all operations from the WAL.
func (w *WAL) ReadAll() ([]string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, err := w.file.Seek(0, 0); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(w.file)

	var operations []string

	for scanner.Scan() {
		operations = append(operations, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return operations, nil
}

// Close closes the WAL file.
func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.file.Close()
}
