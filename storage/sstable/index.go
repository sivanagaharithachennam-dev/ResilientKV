package sstable

// IndexEntry stores the key and its position in the SSTable file.
type IndexEntry struct {
	Key    string
	Offset int64
}

// Index stores SSTable key positions.
type Index struct {
	Entries []IndexEntry
}

// NewIndex creates an empty SSTable index.
func NewIndex() *Index {
	return &Index{
		Entries: make([]IndexEntry, 0),
	}
}

// Add adds a key and its file offset to the index.
func (i *Index) Add(key string, offset int64) {
	i.Entries = append(i.Entries, IndexEntry{
		Key:    key,
		Offset: offset,
	})
}
