package hashtable

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"algorithms/iterator"
)

func TestHashTableImplementsIterator(t *testing.T) {
	var _ iterator.Iterator[Entry[string, string]] = NewHashTable[string, string]()
}

func TestInsertElement(t *testing.T) {
	hashTable := NewHashTable[string, string]()

	hashTable.Insert("foo", "bar")

	if hashTable.actualBucketSize != 1 {
		t.Errorf("Expected size to be 1, got %d", hashTable.actualBucketSize)
	}

	_, index := hashTable.Hash("foo")

	if hashTable.buckets[index].entry.Key != "foo" {
		t.Errorf("Expected key to be 'foo', got %s", hashTable.buckets[0].entry.Key)
	}
}

func TestInsertDuplicatedElement(t *testing.T) {
	hashTable := NewHashTable[string, string]()

	hashTable.Insert("foo", "bar")
	hashTable.Insert("foo", "baz")

	if hashTable.Size() != 1 {
		t.Errorf("Expected size to be 1, got %d", hashTable.Size())
	}

	_, index := hashTable.Hash("foo")

	if hashTable.buckets[index].entry.Value != "baz" {
		t.Errorf("Expected value to be 'baz', got %s", hashTable.buckets[0].entry.Value)
	}
}

func TestGetElement(t *testing.T) {
	hashTable := NewHashTable[string, string]()

	hashTable.Insert("foo", "bar")

	value := hashTable.Get("foo")

	if value != "bar" {
		t.Errorf("Expected value to be 'bar', got %s", value)
	}
}

func TestDeleteKey(t *testing.T) {
	hashTable := NewHashTable[string, string]()

	hashTable.Insert("foo", "bar")
	hashTable.Delete("foo")

	if hashTable.actualBucketSize != 0 {
		t.Errorf("Expected size to be 0, got %d", hashTable.actualBucketSize)
	}
}

func TestGetIterKeyValueFromHashTable(t *testing.T) {
	hashTable := NewHashTable[string, string]()

	hashTable.Insert("foo", "bar")
	hashTable.Insert("baz", "qux")

	expectedEntries := map[string]string{
		"foo": "bar",
		"baz": "qux",
	}

	counter := 0

	for entry := range hashTable.Iter() {
		if expectedEntries[entry.Key] != entry.Value {
			t.Errorf("Expected value to be %s, got %s", expectedEntries[entry.Key], entry.Value)
		}

		counter++
	}

	if counter != 2 {
		t.Errorf("Expected counter to be 2, got %d", counter)
	}
}

func TestSize(t *testing.T) {
	hashTable := NewHashTable[string, string]()

	hashTable.Insert("foo", "bar")
	hashTable.Insert("baz", "qux")

	if hashTable.Size() != 2 {
		t.Errorf("Expected size to be 2, got %d", hashTable.Size())
	}
}

func TestIncreaseBucketLengthWhenMoreThan50ElementsAreInserted(t *testing.T) {
	hashTable := NewHashTable[string, string]()

	for i := 0; i < 20; i++ {
		hashTable.Insert(fmt.Sprint(i), "bar")
	}

	if hashTable.Size() != 20 {
		t.Errorf("Expected bucket length to be 100, got %d", hashTable.Size())
	}
}

func TestShouldPanicWhenAInexistentKeyIsProvided(t *testing.T) {
	hashTable := NewHashTable[string, string]()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	hashTable.Get("foo")
}

func TestMap(t *testing.T) {
	hashTable := NewHashTable[string, string]()

	hashTable.Insert("foo", "bar")
	hashTable.Insert("baz", "qux")

	collection := hashTable.Map(func(entry Entry[string, string]) interface{} {
		return entry.Key
	})

	counter := 0

	for key := range collection.Iter() {
		if key != "foo" && key != "baz" {
			t.Errorf("Expected value to be 'foo' or 'baz', got %s", key)
		}

		counter++
	}
}

func TestFilter(t *testing.T) {
	hashTable := NewHashTable[string, string]()

	hashTable.Insert("foo", "bar")
	hashTable.Insert("baz", "qux")

	collection := hashTable.Filter(func(entry Entry[string, string]) bool {
		return entry.Key == "foo"
	})

	counter := 0

	for entry := range collection.Iter() {
		if entry.Key != "foo" {
			t.Errorf("Expected value to be 'foo', got %s", entry.Key)
		}

		counter++
	}
}

func TestForEach(t *testing.T) {
	hashTable := NewHashTable[string, string]()

	hashTable.Insert("foo", "bar")
	hashTable.Insert("baz", "qux")

	counter := 0

	hashTable.ForEach(func(entry Entry[string, string]) {
		if entry.Key != "foo" && entry.Key != "baz" {
			t.Errorf("Expected value to be 'foo' or 'baz', got %s", entry.Key)
		}

		counter++
	})
}

func TestPerformanceWithTime(t *testing.T) {
	hashTable := NewHashTable[string, string]()

	go PrintMemUsage()

	start := time.Now()

	for i := 1; i < 1_000_000; i++ {
		hashTable.Insert(fmt.Sprint(i), fmt.Sprint(i))
	}

	for i := 1; i < 1_000_000; i++ {
		if hashTable.Get(fmt.Sprint(i)) != fmt.Sprint(i) {
			t.Errorf("Expected value to be %s, got %s", fmt.Sprint(i), hashTable.Get(fmt.Sprint(i)))
		}
	}

	go PrintMemUsage()

	elapsed := time.Since(start)

	fmt.Printf("Time elapsed: %s\n", elapsed)

	if elapsed > 5*time.Second {
		t.Errorf("Expected time to be less than 5 second, got %s", elapsed)
	}
}

func TestPerformanceNumbersWithTime(t *testing.T) {
	hashTable := NewHashTable[int, int]()

	start := time.Now()

	for i := 1; i < 1_000_000; i++ {
		hashTable.Insert(i, i)
	}

	// assert values
	for i := 1; i < 1_000_000; i++ {
		if hashTable.Get(i) != i {
			t.Errorf("Expected value to be %s, got %d", fmt.Sprint(i), hashTable.Get(i))
		}
	}

	elapsed := time.Since(start)

	fmt.Printf("Time elapsed: %s\n", elapsed)

	if elapsed > 5*time.Second {
		t.Errorf("Expected time to be less than 5 second, got %s", elapsed)
	}
}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func TestWithTraditionalMapOfGoLang(t *testing.T) {
	hashTable := make(map[string]string)

	go PrintMemUsage()

	start := time.Now()

	for i := 1; i < 1_000_000; i++ {
		hashTable[fmt.Sprint(i)] = fmt.Sprint(i)
	}

	for i := 1; i < 1_000_000; i++ {
		if hashTable[fmt.Sprint(i)] != fmt.Sprint(i) {
			t.Errorf("Expected value to be %s, got %s", fmt.Sprint(i), hashTable[fmt.Sprint(i)])
		}
	}

	go PrintMemUsage()

	elapsed := time.Since(start)

	fmt.Printf("Time elapsed: %s\n", elapsed)

	if elapsed > 5*time.Second {
		t.Errorf("Expected time to be less than 5 second, got %s", elapsed)
	}
}
