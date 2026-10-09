package concurrent

import (
	"sync"
	"testing"
)

func TestConcurrentMapPutGet(t *testing.T) {
	m := NewConcurrentMap(16)

	m.Put("name", "Yaswanthi")

	value, exists := m.Get("name")

	if !exists {
		t.Fatal("expected key to exist")
	}

	if value != "Yaswanthi" {
		t.Fatalf("expected Yaswanthi, got %s", value)
	}
}

func TestConcurrentMapUpdate(t *testing.T) {
	m := NewConcurrentMap(16)

	m.Put("name", "Old")
	m.Put("name", "New")

	value, _ := m.Get("name")

	if value != "New" {
		t.Fatalf("expected New, got %s", value)
	}
}

func TestConcurrentMapDelete(t *testing.T) {
	m := NewConcurrentMap(16)

	m.Put("name", "Yaswanthi")
	m.Delete("name")

	_, exists := m.Get("name")

	if exists {
		t.Fatal("expected key to be deleted")
	}
}

func TestConcurrentMapSize(t *testing.T) {
	m := NewConcurrentMap(16)

	m.Put("a", "1")
	m.Put("b", "2")
	m.Put("c", "3")

	if m.Size() != 3 {
		t.Fatalf("expected size 3, got %d", m.Size())
	}
}

func TestConcurrentMapConcurrentAccess(t *testing.T) {
	m := NewConcurrentMap(32)

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			key := string(rune('a' + (id % 26)))

			m.Put(key, "value")

			_, _ = m.Get(key)

			if id%10 == 0 {
				m.Delete(key)
			}
		}(i)
	}

	wg.Wait()
}
