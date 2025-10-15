# LinkedList

A generic doubly linked list implementation in Go with type safety and comprehensive operations.

## Features

- **Type-safe**: Uses Go generics for compile-time type safety
- **Doubly linked**: Efficient bidirectional traversal
- **Zero dependencies**: Pure Go implementation
- **Fully tested**: 100% test coverage
- **Idiomatic**: Follows Go best practices and conventions

## Installation

```bash
go get github.com/deanbregenzer/cysl/linkedlist
```

## Usage

### Creating a List

```go
import "github.com/deanbregenzer/cysl/linkedlist"

// Integer list
intList := linkedlist.New[int]()

// String list
strList := linkedlist.New[string]()

// Custom type
type Person struct {
    Name string
    Age  int
}
personList := linkedlist.New[Person]()
```

### Basic Operations

```go
ll := linkedlist.New[int]()

// Add elements
ll.PushBack(1)    // [1]
ll.PushBack(2)    // [1, 2]
ll.PushFront(0)   // [0, 1, 2]

// Access elements
front, ok := ll.Front()  // 0, true
back, ok := ll.Back()    // 2, true

// Remove elements
val, ok := ll.PopFront() // 0, true -> [1, 2]
val, ok = ll.PopBack()   // 2, true -> [1]

// Check size
size := ll.Size()        // 1
empty := ll.IsEmpty()    // false

// Clear all
ll.Clear()               // []
```

### Advanced Operations

```go
ll := linkedlist.New[int]()
ll.PushBack(1)
ll.PushBack(2)
ll.PushBack(3)

// Insert at specific index
ll.Insert(1, 99)  // [1, 99, 2, 3]

// Remove at specific index
val, ok := ll.Remove(1)  // 99, true -> [1, 2, 3]

// Find element
index, ok := ll.Find(func(v int) bool {
    return v == 2
})  // 1, true

// Iterate
ll.ForEach(func(v int) bool {
    fmt.Println(v)
    return true  // continue iteration
})

// Convert to slice
slice := ll.ToSlice()  // []int{1, 2, 3}
```

## API Reference

### Creation

- `New[T any]() *LinkedList[T]` - Creates an empty list

### Adding Elements

- `PushFront(value T)` - Adds element to front
- `PushBack(value T)` - Adds element to back
- `Insert(index int, value T) bool` - Inserts at index

### Removing Elements

- `PopFront() (T, bool)` - Removes and returns front element
- `PopBack() (T, bool)` - Removes and returns back element
- `Remove(index int) (T, bool)` - Removes element at index
- `Clear()` - Removes all elements

### Accessing Elements

- `Front() (T, bool)` - Returns front element without removing
- `Back() (T, bool)` - Returns back element without removing

### Querying

- `Size() int` - Returns number of elements
- `IsEmpty() bool` - Checks if list is empty
- `Find(cmp func(T) bool) (int, bool)` - Finds element by predicate

### Iteration

- `ForEach(fn func(T) bool)` - Applies function to each element
- `ToSlice() []T` - Converts to slice

### Utilities

- `String() string` - Returns string representation

## Performance

| Operation | Time Complexity | Space Complexity |
|-----------|----------------|------------------|
| PushFront | O(1) | O(1) |
| PushBack | O(1) | O(1) |
| PopFront | O(1) | O(1) |
| PopBack | O(1) | O(1) |
| Front | O(1) | O(1) |
| Back | O(1) | O(1) |
| Insert | O(n) | O(1) |
| Remove | O(n) | O(1) |
| Find | O(n) | O(1) |
| Size | O(1) | O(1) |
| ToSlice | O(n) | O(n) |

## Testing

Run tests with:

```bash
go test ./linkedlist/...
```

Run tests with coverage:

```bash
go test ./linkedlist/... -cover
```

## License

See project root for license information.
