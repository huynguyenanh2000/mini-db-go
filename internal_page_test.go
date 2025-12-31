package main

import (
	"bytes"
	"testing"
)

func TestINode(t *testing.T) {
	node := NewIPage()
	var c uint64 = 0
	key3 := NewKeyEntryFromInt(3)
	node.InsertKV(key3, c)
	// [3]
	if node.nkey != 1 {
		t.Errorf("Expected nkey to be 1, got %d", node.nkey)
	}
	if node.keys[0].compare(&key3) != 0 {
		t.Errorf("Expected key at position 0 to be 3, got %d", node.keys[0])
	}
	key10 := NewKeyEntryFromInt(10)
	node.InsertKV(key10, c)
	// [3, 10]
	if node.nkey != 2 {
		t.Errorf("Expected nkey to be 2, got %d", node.nkey)
	}
	if node.keys[0].compare(&key3) != 0 {
		t.Errorf("Expected key at position 0 to be 3, got %d", node.keys[0])
	}
	if node.keys[1].compare(&key10) != 0 {
		t.Errorf("Expected key at position 1 to be 10, got %d", node.keys[1])
	}
	key5 := NewKeyEntryFromInt(5)
	node.InsertKV(key5, c)
	// [3, 5, 10]
	if node.nkey != 3 {
		t.Errorf("Expected nkey to be 3, got %d", node.nkey)
	}
	if node.keys[0].compare(&key3) != 0 {
		t.Errorf("Expected key at position 0 to be 3, got %d", node.keys[0])
	}
	if node.keys[1].compare(&key5) != 0 {
		t.Errorf("Expected key at position 1 to be 5, got %d", node.keys[1])
	}
	if node.keys[2].compare(&key10) != 0 {
		t.Errorf("Expected key at position 2 to be 10, got %d", node.keys[2])
	}
	key12 := NewKeyEntryFromInt(12)
	node.InsertKV(key12, c)
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
	if node.keys[0].compare(&key3) != 0 {
		t.Errorf("Expected original node key at position 0 to be 3 after split, got %d", node.keys[0])
	}
	if node.keys[1].compare(&key5) != 0 {
		t.Errorf("Expected original node key at position 1 to be 5 after split, got %d", node.keys[1])
	}
	if newNode.keys[0].compare(&key10) != 0 {
		t.Errorf("Expected new node key at position 0 to be 10 after split, got %d", newNode.keys[0])
	}
	if newNode.keys[1].compare(&key12) != 0 {
		t.Errorf("Expected new node key at position 1 to be 12 after split, got %d", newNode.keys[1])
	}

	buf := new(bytes.Buffer)
	node.writeToBuffer(buf)

	clonedNode := NewIPage()
	clonedNode.readFromBuffer(buf)

	if clonedNode.nkey != 2 {
		t.Errorf("Expected original clonedNode nkey to be 2 after split, got %d", clonedNode.nkey)
	}

	if clonedNode.keys[0].compare(&key3) != 0 {
		t.Errorf("Expected original clonedNode key at position 0 to be 3 after split, got %d", clonedNode.keys[0])
	}
	if clonedNode.keys[1].compare(&key5) != 0 {
		t.Errorf("Expected original clonedNode key at position 1 to be 5 after split, got %d", clonedNode.keys[1])
	}
}
