package stats

import (
	"sync"
	"testing"
)

func TestStore_MostEmpty(t *testing.T) {
	store := New()
	_, ok := store.Most()
	if ok {
		t.Fatal("expected ok == false but got true")
	}
}

func TestStore_MostSingle(t *testing.T) {
	s := New()
	s.Increment("Hel")
	s.Increment("Hel")
	s.Increment("My")

	got, ok := s.Most()
	if !ok {
		t.Fatal("expected ok=true")
	}
	if got.Key != "Hel" || got.Count != 2 {
		t.Errorf("Most() = %+v, want {Hel 2}", got)
	}
}
func TestStore_MostTie(t *testing.T) {
	store := New()
	store.Increment("Helmy")
	store.Increment("Helmy")
	store.Increment("Bc")
	store.Increment("Bc")
	store.Increment("HelmyBc")

	// In case of a tie, the winner is arbitrary.
	// "Helmy" and "Bc" are tied at 2 hits.
	// So we test both keys and that the winning key has the correct count.
	got, ok := store.Most()
	if !ok {
		t.Fatal("expected ok == true")
	}
	if got.Key != "Helmy" && got.Key != "Bc" {
		t.Fatalf("expected 'Helmy' or 'Bc' as the winning key but got %s", got.Key)
	}
	if got.Count != 2 {
		t.Fatalf("expected 2 hits for the winning key but got %d", got.Count)
	}

}

func TestStore_ConcurrentIncrement(t *testing.T) {
	store := New()
	const goroutines = 100
	const perGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			for range perGoroutine {
				store.Increment("test")
			}
		}()
	}
	wg.Wait()

	got, ok := store.Most()
	if !ok {
		t.Fatal("expected ok==true")
	}
	want := int64(goroutines * perGoroutine)
	if got.Count != want {
		t.Errorf("Count = %d, want %d (possible race in Increment)", got.Count, want)
	}
}
