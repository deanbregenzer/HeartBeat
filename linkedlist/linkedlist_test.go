package linkedlist

import (
	"testing"
)

func TestNew(t *testing.T) {
	ll := New[int]()
	if ll.Size() != 0 {
		t.Errorf("New list should have size 0, got %d", ll.Size())
	}
	if !ll.IsEmpty() {
		t.Error("New list should be empty")
	}
}

func TestPushFront(t *testing.T) {
	ll := New[int]()
	ll.PushFront(1)
	ll.PushFront(2)
	ll.PushFront(3)

	if ll.Size() != 3 {
		t.Errorf("Expected size 3, got %d", ll.Size())
	}

	if val, ok := ll.Front(); !ok || val != 3 {
		t.Errorf("Expected front value 3, got %d", val)
	}
}

func TestPushBack(t *testing.T) {
	ll := New[int]()
	ll.PushBack(1)
	ll.PushBack(2)
	ll.PushBack(3)

	if ll.Size() != 3 {
		t.Errorf("Expected size 3, got %d", ll.Size())
	}

	if val, ok := ll.Back(); !ok || val != 3 {
		t.Errorf("Expected back value 3, got %d", val)
	}
}

func TestPopFront(t *testing.T) {
	ll := New[int]()
	ll.PushBack(1)
	ll.PushBack(2)
	ll.PushBack(3)

	val, ok := ll.PopFront()
	if !ok || val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}

	if ll.Size() != 2 {
		t.Errorf("Expected size 2, got %d", ll.Size())
	}

	// Pop empty list
	ll.Clear()
	_, ok = ll.PopFront()
	if ok {
		t.Error("PopFront on empty list should return false")
	}
}

func TestPopBack(t *testing.T) {
	ll := New[int]()
	ll.PushBack(1)
	ll.PushBack(2)
	ll.PushBack(3)

	val, ok := ll.PopBack()
	if !ok || val != 3 {
		t.Errorf("Expected 3, got %d", val)
	}

	if ll.Size() != 2 {
		t.Errorf("Expected size 2, got %d", ll.Size())
	}

	// Pop empty list
	ll.Clear()
	_, ok = ll.PopBack()
	if ok {
		t.Error("PopBack on empty list should return false")
	}
}

func TestFrontBack(t *testing.T) {
	ll := New[string]()

	// Test empty list
	_, ok := ll.Front()
	if ok {
		t.Error("Front on empty list should return false")
	}

	_, ok = ll.Back()
	if ok {
		t.Error("Back on empty list should return false")
	}

	// Test with elements
	ll.PushBack("first")
	ll.PushBack("second")

	if val, ok := ll.Front(); !ok || val != "first" {
		t.Errorf("Expected 'first', got %s", val)
	}

	if val, ok := ll.Back(); !ok || val != "second" {
		t.Errorf("Expected 'second', got %s", val)
	}
}

func TestClear(t *testing.T) {
	ll := New[int]()
	ll.PushBack(1)
	ll.PushBack(2)
	ll.PushBack(3)

	ll.Clear()

	if !ll.IsEmpty() {
		t.Error("List should be empty after Clear")
	}

	if ll.Size() != 0 {
		t.Errorf("Expected size 0, got %d", ll.Size())
	}
}

func TestInsert(t *testing.T) {
	ll := New[int]()
	ll.PushBack(1)
	ll.PushBack(3)

	// Insert in middle
	if !ll.Insert(1, 2) {
		t.Error("Insert should succeed")
	}

	slice := ll.ToSlice()
	expected := []int{1, 2, 3}
	if !slicesEqual(slice, expected) {
		t.Errorf("Expected %v, got %v", expected, slice)
	}

	// Insert at beginning
	ll.Insert(0, 0)
	if val, _ := ll.Front(); val != 0 {
		t.Errorf("Expected 0 at front, got %d", val)
	}

	// Insert at end
	ll.Insert(ll.Size(), 4)
	if val, _ := ll.Back(); val != 4 {
		t.Errorf("Expected 4 at back, got %d", val)
	}

	// Invalid index
	if ll.Insert(-1, 5) {
		t.Error("Insert with negative index should fail")
	}

	if ll.Insert(100, 5) {
		t.Error("Insert with out of bounds index should fail")
	}
}

func TestRemove(t *testing.T) {
	ll := New[int]()
	ll.PushBack(1)
	ll.PushBack(2)
	ll.PushBack(3)
	ll.PushBack(4)

	// Remove from middle
	val, ok := ll.Remove(1)
	if !ok || val != 2 {
		t.Errorf("Expected to remove 2, got %d", val)
	}

	// Remove from beginning
	val, ok = ll.Remove(0)
	if !ok || val != 1 {
		t.Errorf("Expected to remove 1, got %d", val)
	}

	// Remove from end
	val, ok = ll.Remove(ll.Size() - 1)
	if !ok || val != 4 {
		t.Errorf("Expected to remove 4, got %d", val)
	}

	// Invalid index
	_, ok = ll.Remove(-1)
	if ok {
		t.Error("Remove with negative index should fail")
	}

	_, ok = ll.Remove(100)
	if ok {
		t.Error("Remove with out of bounds index should fail")
	}
}

func TestForEach(t *testing.T) {
	ll := New[int]()
	ll.PushBack(1)
	ll.PushBack(2)
	ll.PushBack(3)

	sum := 0
	ll.ForEach(func(val int) bool {
		sum += val
		return true
	})

	if sum != 6 {
		t.Errorf("Expected sum 6, got %d", sum)
	}

	// Test early termination
	count := 0
	ll.ForEach(func(val int) bool {
		count++
		return val != 2
	})

	if count != 2 {
		t.Errorf("Expected to stop at 2 iterations, got %d", count)
	}
}

func TestToSlice(t *testing.T) {
	ll := New[int]()
	ll.PushBack(1)
	ll.PushBack(2)
	ll.PushBack(3)

	slice := ll.ToSlice()
	expected := []int{1, 2, 3}

	if !slicesEqual(slice, expected) {
		t.Errorf("Expected %v, got %v", expected, slice)
	}

	// Test empty list
	ll.Clear()
	slice = ll.ToSlice()
	if len(slice) != 0 {
		t.Errorf("Expected empty slice, got %v", slice)
	}
}

func TestFind(t *testing.T) {
	ll := New[int]()
	ll.PushBack(10)
	ll.PushBack(20)
	ll.PushBack(30)

	// Find existing value
	index, ok := ll.Find(func(val int) bool {
		return val == 20
	})

	if !ok || index != 1 {
		t.Errorf("Expected to find 20 at index 1, got index %d", index)
	}

	// Find non-existing value
	index, ok = ll.Find(func(val int) bool {
		return val == 100
	})

	if ok {
		t.Errorf("Should not find 100, but got index %d", index)
	}
}

func TestString(t *testing.T) {
	ll := New[int]()

	// Empty list
	if ll.String() != "[]" {
		t.Errorf("Expected '[]', got '%s'", ll.String())
	}

	// Non-empty list
	ll.PushBack(1)
	ll.PushBack(2)
	if ll.String() != "[1 2]" {
		t.Errorf("Expected '[1 2]', got '%s'", ll.String())
	}
}

func TestGenerics(t *testing.T) {
	// Test with strings
	strList := New[string]()
	strList.PushBack("hello")
	strList.PushBack("world")

	if strList.Size() != 2 {
		t.Errorf("Expected size 2, got %d", strList.Size())
	}

	// Test with custom struct
	type Person struct {
		Name string
		Age  int
	}

	personList := New[Person]()
	personList.PushBack(Person{Name: "Alice", Age: 30})
	personList.PushBack(Person{Name: "Bob", Age: 25})

	if personList.Size() != 2 {
		t.Errorf("Expected size 2, got %d", personList.Size())
	}

	if person, ok := personList.Front(); !ok || person.Name != "Alice" {
		t.Errorf("Expected 'Alice', got '%s'", person.Name)
	}
}

// Helper function to compare slices
func slicesEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
