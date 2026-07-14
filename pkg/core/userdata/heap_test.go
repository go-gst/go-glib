package userdata

import (
	"testing"
)

func TestGetPointer(t *testing.T) {
	// Clear the free list for testing
	mu.Lock()
	unused = nil
	mu.Unlock()

	// Get a pointer and ensure it's not nil
	ptr := getPointer()
	if ptr == nil {
		t.Fatalf("Expected a valid pointer, got nil")
	}

	// Ensure the free list has been reduced by one
	mu.Lock()
	defer mu.Unlock()
	if len(unused) != chunkSize-1 {
		t.Fatalf("Expected freeList size to be %d, got %d", chunkSize-1, len(unused))
	}
}

func TestReturnPointer(t *testing.T) {
	// Clear the free list for testing
	mu.Lock()
	unused = nil
	mu.Unlock()

	// Get a pointer and return it
	ptr := getPointer()
	returnPointer(ptr)

	// Ensure the pointer is back in the free list
	mu.Lock()
	defer mu.Unlock()
	if len(unused) != chunkSize {
		t.Fatalf("Expected freeList size to be %d, got %d", chunkSize, len(unused))
	}

	// Ensure the returned pointer is the same as the one added
	if unused[len(unused)-1] != ptr {
		t.Fatalf("Expected returned pointer to match, but it did not")
	}
}
