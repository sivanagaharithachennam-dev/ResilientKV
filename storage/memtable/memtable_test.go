package memtable

import "testing"

func TestMemTablePutGet(t *testing.T) {
	m := New()

	m.Put("name", "Yaswanthi")

	value, exists := m.Get("name")

	if !exists {
		t.Fatal("expected key to exist")
	}

	if value != "Yaswanthi" {
		t.Fatalf("expected Yaswanthi, got %s", value)
	}
}

func TestMemTableUpdate(t *testing.T) {
	m := New()

	m.Put("name", "Old")
	m.Put("name", "New")

	value, _ := m.Get("name")

	if value != "New" {
		t.Fatalf("expected New, got %s", value)
	}
}

func TestMemTableDelete(t *testing.T) {
	m := New()

	m.Put("name", "Yaswanthi")
	m.Delete("name")

	_, exists := m.Get("name")

	if exists {
		t.Fatal("expected key to be deleted")
	}

	if !m.IsDeleted("name") {
		t.Fatal("expected tombstone to exist")
	}
}

func TestMemTableSize(t *testing.T) {
	m := New()

	m.Put("a", "1")
	m.Put("b", "2")
	m.Put("c", "3")

	if m.Size() != 3 {
		t.Fatalf("expected size 3, got %d", m.Size())
	}
}
