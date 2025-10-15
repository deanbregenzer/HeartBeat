// Package linkedlist provides a generic doubly linked list implementation.
package linkedlist

import "fmt"

// Node represents a single element in the linked list.
// It holds a value and pointers to the next and previous nodes.
type Node[T any] struct {
	Value T
	Next  *Node[T]
	Prev  *Node[T]
}

// LinkedList represents a doubly linked list data structure.
// It maintains pointers to the head and tail for efficient operations.
type LinkedList[T any] struct {
	head *Node[T]
	tail *Node[T]
	size int
}

// New creates and returns an empty linked list.
func New[T any]() *LinkedList[T] {
	return &LinkedList[T]{}
}

// PushFront adds a new element to the front of the list.
func (ll *LinkedList[T]) PushFront(value T) {
	node := &Node[T]{Value: value}

	if ll.head == nil {
		ll.head = node
		ll.tail = node
	} else {
		node.Next = ll.head
		ll.head.Prev = node
		ll.head = node
	}

	ll.size++
}

// PushBack adds a new element to the back of the list.
func (ll *LinkedList[T]) PushBack(value T) {
	node := &Node[T]{Value: value}

	if ll.tail == nil {
		ll.head = node
		ll.tail = node
	} else {
		node.Prev = ll.tail
		ll.tail.Next = node
		ll.tail = node
	}

	ll.size++
}

// PopFront removes and returns the first element.
// Returns zero value and false if the list is empty.
func (ll *LinkedList[T]) PopFront() (T, bool) {
	if ll.head == nil {
		var zero T
		return zero, false
	}

	value := ll.head.Value
	ll.head = ll.head.Next

	if ll.head == nil {
		ll.tail = nil
	} else {
		ll.head.Prev = nil
	}

	ll.size--
	return value, true
}

// PopBack removes and returns the last element.
// Returns zero value and false if the list is empty.
func (ll *LinkedList[T]) PopBack() (T, bool) {
	if ll.tail == nil {
		var zero T
		return zero, false
	}

	value := ll.tail.Value
	ll.tail = ll.tail.Prev

	if ll.tail == nil {
		ll.head = nil
	} else {
		ll.tail.Next = nil
	}

	ll.size--
	return value, true
}

// Front returns the first element without removing it.
// Returns zero value and false if the list is empty.
func (ll *LinkedList[T]) Front() (T, bool) {
	if ll.head == nil {
		var zero T
		return zero, false
	}
	return ll.head.Value, true
}

// Back returns the last element without removing it.
// Returns zero value and false if the list is empty.
func (ll *LinkedList[T]) Back() (T, bool) {
	if ll.tail == nil {
		var zero T
		return zero, false
	}
	return ll.tail.Value, true
}

// Size returns the number of elements in the list.
func (ll *LinkedList[T]) Size() int {
	return ll.size
}

// IsEmpty returns true if the list contains no elements.
func (ll *LinkedList[T]) IsEmpty() bool {
	return ll.size == 0
}

// Clear removes all elements from the list.
func (ll *LinkedList[T]) Clear() {
	ll.head = nil
	ll.tail = nil
	ll.size = 0
}

// ForEach applies a function to each element in the list.
// Iteration stops if the function returns false.
func (ll *LinkedList[T]) ForEach(fn func(T) bool) {
	current := ll.head
	for current != nil {
		if !fn(current.Value) {
			return
		}
		current = current.Next
	}
}

// ToSlice returns a slice containing all elements in order.
func (ll *LinkedList[T]) ToSlice() []T {
	result := make([]T, 0, ll.size)
	ll.ForEach(func(value T) bool {
		result = append(result, value)
		return true
	})
	return result
}

// String returns a string representation of the list.
func (ll *LinkedList[T]) String() string {
	if ll.IsEmpty() {
		return "[]"
	}
	return fmt.Sprintf("%v", ll.ToSlice())
}

// Insert adds a value at the specified index.
// Returns false if the index is out of bounds.
func (ll *LinkedList[T]) Insert(index int, value T) bool {
	if index < 0 || index > ll.size {
		return false
	}

	if index == 0 {
		ll.PushFront(value)
		return true
	}

	if index == ll.size {
		ll.PushBack(value)
		return true
	}

	current := ll.head
	for i := 0; i < index; i++ {
		current = current.Next
	}

	node := &Node[T]{Value: value}
	node.Prev = current.Prev
	node.Next = current
	current.Prev.Next = node
	current.Prev = node

	ll.size++
	return true
}

// Remove deletes the element at the specified index.
// Returns the removed value and true, or zero value and false if index is invalid.
func (ll *LinkedList[T]) Remove(index int) (T, bool) {
	if index < 0 || index >= ll.size {
		var zero T
		return zero, false
	}

	if index == 0 {
		return ll.PopFront()
	}

	if index == ll.size-1 {
		return ll.PopBack()
	}

	current := ll.head
	for i := 0; i < index; i++ {
		current = current.Next
	}

	value := current.Value
	current.Prev.Next = current.Next
	current.Next.Prev = current.Prev

	ll.size--
	return value, true
}

// Find searches for the first occurrence of a value using a comparison function.
// Returns the index and true if found, or -1 and false otherwise.
func (ll *LinkedList[T]) Find(cmp func(T) bool) (int, bool) {
	current := ll.head
	index := 0

	for current != nil {
		if cmp(current.Value) {
			return index, true
		}
		current = current.Next
		index++
	}

	return -1, false
}
