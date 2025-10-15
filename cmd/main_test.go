package main

import "testing"

func TestHelloWorld(t *testing.T) {
	expected := "hello World"
	result := HelloWorld()

	if result != expected {
		t.Errorf("HelloWorld() = %q, want %q", result, expected)
	}
}
