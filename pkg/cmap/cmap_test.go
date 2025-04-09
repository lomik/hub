package cmap

import (
	"sync"
	"testing"
	"time"
)

func TestCMap_BasicOperations(t *testing.T) {
	t.Parallel()

	t.Run("Set/Get existing key", func(t *testing.T) {
		c := New()
		key := "testKey"
		value := 42
		c.Set(key, value)

		if v, ok := c.Get(key); !ok || v != value {
			t.Errorf("Get() = (%v, %v), want (%v, true)", v, ok, value)
		}
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		c := New()
		if v, ok := c.Get("non-existent"); ok || v != 0 {
			t.Errorf("Get() = (%v, %v), want (0, false)", v, ok)
		}
	})

	t.Run("Delete existing key", func(t *testing.T) {
		c := New()
		key := "toDelete"
		c.Set(key, 100)
		c.Delete(key)

		if _, ok := c.Get(key); ok {
			t.Error("Key still exists after Delete()")
		}
	})

	t.Run("Delete non-existent key", func(t *testing.T) {
		c := New()
		// Should not panic
		c.Delete("non-existent")
	})

	t.Run("Len after operations", func(t *testing.T) {
		c := New()
		c.Set("a", 1)
		c.Set("b", 2)
		c.Delete("a")

		if l := c.Len(); l != 1 {
			t.Errorf("Len() = %d, want 1", l)
		}
	})
}

func TestCMap_Iterate(t *testing.T) {
	t.Parallel()
	c := New()

	keys := []string{"a", "b", "c"}
	for i, k := range keys {
		c.Set(k, i+1)
	}

	t.Run("iterate all elements", func(t *testing.T) {
		seen := make(map[string]int)
		c.Iterate(func(k string, v int) {
			seen[k] = v
		})

		if len(seen) != len(keys) {
			t.Errorf("Iterate() visited %d elements, want %d", len(seen), len(keys))
		}

		for _, k := range keys {
			if _, ok := seen[k]; !ok {
				t.Errorf("Key %q not visited", k)
			}
		}
	})

}

func TestCMap_Concurrency(t *testing.T) {
	t.Parallel()
	c := New()
	const workers = 100
	const iterations = 1000

	var wg sync.WaitGroup
	wg.Add(workers * 2)

	// Concurrent writers
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := string(rune(id)) + string(rune(j))
				c.Set(key, j)
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := string(rune(id)) + string(rune(j))
				c.Get(key)
				c.Len()
			}
		}(i)
	}

	// Allow time for all operations to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait with timeout to prevent hanging
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for concurrent operations")
	}

	// Verify data integrity after concurrent access
	// Allow for some key overwrites due to concurrent writes
	expectedMinimum := workers * iterations / 2
	if c.Len() < expectedMinimum {
		t.Errorf("Unexpected map size after concurrency test: %d (min expected %d)",
			c.Len(), expectedMinimum)
	}
}

func TestCMap_Clear(t *testing.T) {
	t.Parallel()
	c := New()

	t.Run("clear empty map", func(t *testing.T) {
		c.Clear()
		if l := c.Len(); l != 0 {
			t.Errorf("Len() after clear = %d, want 0", l)
		}
	})

	t.Run("clear populated map", func(t *testing.T) {
		c.Set("a", 1)
		c.Set("b", 2)
		c.Clear()

		if l := c.Len(); l != 0 {
			t.Errorf("Len() after clear = %d, want 0", l)
		}
		if _, ok := c.Get("a"); ok {
			t.Error("Key 'a' still exists after clear")
		}
	})

	t.Run("concurrent clear", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(2)

		// Populate map
		for i := 0; i < 1000; i++ {
			c.Set(string(rune(i)), i)
		}

		go func() {
			defer wg.Done()
			c.Clear()
		}()

		go func() {
			defer wg.Done()
			c.Clear()
		}()

		wg.Wait()

		if l := c.Len(); l != 0 {
			t.Errorf("Len() after concurrent clear = %d, want 0", l)
		}
	})
}

func TestCMap_Add(t *testing.T) {
	t.Parallel()
	c := New()

	t.Run("add to new key", func(t *testing.T) {
		key := "new"
		c.Add(key, 5)
		if v, _ := c.Get(key); v != 5 {
			t.Errorf("Add() = %d, want 5", v)
		}
	})

	t.Run("add to existing key", func(t *testing.T) {
		key := "exists"
		c.Set(key, 10)
		c.Add(key, 3)
		if v, _ := c.Get(key); v != 13 {
			t.Errorf("Add() = %d, want 13", v)
		}
	})

	t.Run("negative delta", func(t *testing.T) {
		key := "negative"
		c.Set(key, 8)
		c.Add(key, -5)
		if v, _ := c.Get(key); v != 3 {
			t.Errorf("Add() = %d, want 3", v)
		}
	})

	t.Run("concurrent adds", func(t *testing.T) {
		key := "concurrent"
		const routines = 100
		const addsPerRoutine = 100

		var wg sync.WaitGroup
		wg.Add(routines)

		for i := 0; i < routines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < addsPerRoutine; j++ {
					c.Add(key, 1)
				}
			}()
		}

		wg.Wait()

		if v, _ := c.Get(key); v != routines*addsPerRoutine {
			t.Errorf("Add() = %d, want %d", v, routines*addsPerRoutine)
		}
	})
}

func TestCMap_Eq(t *testing.T) {
	t.Parallel()

	t.Run("equal maps", func(t *testing.T) {
		c := New()
		c.Set("a", 1)
		c.Set("b", 2)

		other := map[string]int{"a": 1, "b": 2}
		if !c.Eq(other) {
			t.Error("Expected maps to be equal, but they are not")
		}
	})

	t.Run("different values", func(t *testing.T) {
		c := New()
		c.Set("a", 1)
		c.Set("b", 2)

		other := map[string]int{"a": 1, "b": 3}
		if c.Eq(other) {
			t.Error("Expected maps to be different, but they are equal")
		}
	})

	t.Run("missing key in CMap", func(t *testing.T) {
		c := New()
		c.Set("a", 1)

		other := map[string]int{"a": 1, "b": 2}
		if c.Eq(other) {
			t.Error("Expected maps to be different due to missing key in CMap, but they are equal")
		}
	})

	t.Run("extra key in CMap", func(t *testing.T) {
		c := New()
		c.Set("a", 1)
		c.Set("b", 2)

		other := map[string]int{"a": 1}
		if c.Eq(other) {
			t.Error("Expected maps to be different due to extra key in CMap, but they are equal")
		}
	})

	t.Run("empty maps", func(t *testing.T) {
		c := New()

		other := map[string]int{}
		if !c.Eq(other) {
			t.Error("Expected empty maps to be equal, but they are not")
		}
	})

	t.Run("CMap empty, other non-empty", func(t *testing.T) {
		c := New()

		other := map[string]int{"a": 1}
		if c.Eq(other) {
			t.Error("Expected maps to be different, but they are equal")
		}
	})

	t.Run("CMap non-empty, other empty", func(t *testing.T) {
		c := New()
		c.Set("a", 1)

		other := map[string]int{}
		if c.Eq(other) {
			t.Error("Expected maps to be different, but they are equal")
		}
	})
}
