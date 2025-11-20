package main

import "testing"

func TestDatabase(t *testing.T) {
	node := NewINode()
	child := NewINode()
	node.InsertKV(3, child)
	// [3]
	if node.nkey != 1 {
		t.Errorf("Expected nkey to be 1, got %d", node.nkey)
	}
	if node.keys[0] != 3 {
		t.Errorf("Expected key at position 0 to be 3, got %d", node.keys[0])
	}
	node.InsertKV(10, child)
	// [3, 10]
	if node.nkey != 2 {
		t.Errorf("Expected nkey to be 2, got %d", node.nkey)
	}
	if node.keys[0] != 3 {
		t.Errorf("Expected key at position 0 to be 3, got %d", node.keys[0])
	}
	if node.keys[1] != 10 {
		t.Errorf("Expected key at position 1 to be 10, got %d", node.keys[1])
	}
	node.InsertKV(5, child)
	// [3, 5, 10]
	if node.nkey != 3 {
		t.Errorf("Expected nkey to be 3, got %d", node.nkey)
	}
	if node.keys[0] != 3 {
		t.Errorf("Expected key at position 0 to be 3, got %d", node.keys[0])
	}
	if node.keys[1] != 5 {
		t.Errorf("Expected key at position 1 to be 5, got %d", node.keys[1])
	}
	if node.keys[2] != 10 {
		t.Errorf("Expected key at position 2 to be 10, got %d", node.keys[2])
	}
	node.InsertKV(12, child)
	// [3, 5, 10, 12]
	if node.nkey != 4 {
		t.Errorf("Expected nkey to be 4, got %d", node.nkey)
	}
	newNode := node.Split()
	// original node: [3,5], new node: [10,12]
	if node.nkey != 2 {
		t.Errorf("Expected original node nkey to be 2 after split, got %d", node.nkey)
	}
	if newNode.nkey != 2 {
		t.Errorf("Expected new node nkey to be 2 after split, got %d", newNode.nkey)
	}
	if node.keys[0] != 3 {
		t.Errorf("Expected original node key at position 0 to be 3 after split, got %d", node.keys[0])
	}
	if node.keys[1] != 5 {
		t.Errorf("Expected original node key at position 1 to be 5 after split, got %d", node.keys[1])
	}
	if newNode.keys[0] != 10 {
		t.Errorf("Expected new node key at position 0 to be 10 after split, got %d", newNode.keys[0])
	}
	if newNode.keys[1] != 12 {
		t.Errorf("Expected new node key at position 1 to be 12 after split, got %d", newNode.keys[1])
	}
}
