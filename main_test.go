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

func TestBTree(t *testing.T) {
	tree := NewBPTree()
	// head = inode[]
	tree.Insert(3, 3)
	// head = inode[3], 3 -> lnode[(3, 3)]
	if tree.head.(*BTreeInternalNode).keys[0] != 3 {
		t.Errorf("got key[0] = %v, expect %v", tree.head.(*BTreeInternalNode).keys[0], 3)
	}
	child := tree.head.(*BTreeInternalNode).childrens[0]
	if (*child).(*BTreeLeafNode).keys[0] != 3 {
		t.Errorf("got key[0] = %v, expect %v", (*child).(*BTreeLeafNode).keys[0], 3)
	}
	if (*child).(*BTreeLeafNode).values[0] != 3 {
		t.Errorf("got values[0] = %v, expect %v", (*child).(*BTreeLeafNode).values[0], 3)
	}

	// head = inode[3], 3 -> lnode([3, 3])
	tree.Insert(5, 5)
	// head = inode[3], 3 -> lnode([3, 3], [5, 5])
	if tree.head.(*BTreeInternalNode).keys[0] != 3 {
		t.Errorf("got key[0] = %v, expect %v", tree.head.(*BTreeInternalNode).keys[0], 3)
	}
	child = tree.head.(*BTreeInternalNode).childrens[0]
	if (*child).(*BTreeLeafNode).nkey != 2 {
		t.Errorf("got nkey = %v, expect %v", (*child).(*BTreeLeafNode).nkey, 2)
	}
	if (*child).(*BTreeLeafNode).keys[0] != 3 {
		t.Errorf("got key[0] = %v, expect %v", (*child).(*BTreeLeafNode).keys[0], 3)
	}
	if (*child).(*BTreeLeafNode).values[0] != 3 {
		t.Errorf("got values[0] = %v, expect %v", (*child).(*BTreeLeafNode).values[0], 3)
	}
	if (*child).(*BTreeLeafNode).keys[1] != 5 {
		t.Errorf("got key[0] = %v, expect %v", (*child).(*BTreeLeafNode).keys[0], 3)
	}
	if (*child).(*BTreeLeafNode).values[1] != 5 {
		t.Errorf("got values[0] = %v, expect %v", (*child).(*BTreeLeafNode).values[0], 3)
	}

	tree.Insert(2, 2)
	// head = iNode[2], 2 -> lnode[(2, 2), (3, 3), (5, 5)]
	if (*child).(*BTreeLeafNode).nkey != 3 {
		t.Errorf("got nkey = %v, expect %v", (*child).(*BTreeLeafNode).nkey, 3)
	}
	if (*child).(*BTreeLeafNode).keys[0] != 2 {
		t.Errorf("got key[0] = %v, expect %v", (*child).(*BTreeLeafNode).keys[0], 2)
	}
	if (*child).(*BTreeLeafNode).values[0] != 2 {
		t.Errorf("got values[0] = %v, expect %v", (*child).(*BTreeLeafNode).values[0], 2)
	}
	if (*child).(*BTreeLeafNode).keys[1] != 3 {
		t.Errorf("got key[0] = %v, expect %v", (*child).(*BTreeLeafNode).keys[0], 3)
	}
	if (*child).(*BTreeLeafNode).values[1] != 3 {
		t.Errorf("got values[0] = %v, expect %v", (*child).(*BTreeLeafNode).values[0], 3)
	}
	if (*child).(*BTreeLeafNode).keys[2] != 5 {
		t.Errorf("got key[0] = %v, expect %v", (*child).(*BTreeLeafNode).keys[0], 3)
	}
	if (*child).(*BTreeLeafNode).values[2] != 5 {
		t.Errorf("got values[0] = %v, expect %v", (*child).(*BTreeLeafNode).values[0], 3)
	}

	// head = inode[2], 2 -> lnode[(2, 2), (3, 3), (5, 5)]
	tree.Insert(8, 8)
	// head = inode[2], 2 -> lnode[(2, 2), (3, 3), (5, 5), (8, 8)]
	// head = inode[2, 5], 2 -> lnode[(2, 2), (3, 3)], 5 -> lnode[(5, 5), (8, 8)]
	if tree.head.(*BTreeInternalNode).nkey != 2 {
		t.Errorf("got values[0] = %v, expect %v", tree.head.(*BTreeInternalNode).nkey, 2)
	}
	if tree.head.(*BTreeInternalNode).keys[0] != 2 {
		t.Errorf("got key[0] = %v, expect %v", tree.head.(*BTreeInternalNode).keys[0], 2)
	}
	if tree.head.(*BTreeInternalNode).keys[1] != 5 {
		t.Errorf("got key[0] = %v, expect %v", tree.head.(*BTreeInternalNode).keys[1], 5)
	}
	child = tree.head.(*BTreeInternalNode).childrens[1]
	if (*child).(*BTreeLeafNode).nkey != 2 {
		t.Errorf("got nkey = %v, expect %v", (*child).(*BTreeLeafNode).nkey, 2)
	}
	if (*child).(*BTreeLeafNode).keys[0] != 5 {
		t.Errorf("got key[0] = %v, expect %v", (*child).(*BTreeLeafNode).keys[0], 3)
	}
	if (*child).(*BTreeLeafNode).values[0] != 5 {
		t.Errorf("got values[0] = %v, expect %v", (*child).(*BTreeLeafNode).values[0], 3)
	}
	if (*child).(*BTreeLeafNode).keys[1] != 8 {
		t.Errorf("got key[0] = %v, expect %v", (*child).(*BTreeLeafNode).keys[0], 3)
	}
	if (*child).(*BTreeLeafNode).values[1] != 8 {
		t.Errorf("got values[0] = %v, expect %v", (*child).(*BTreeLeafNode).values[0], 3)
	}

	// Renew test
	tree = NewBPTree()
	for i := range 100 {
		tree.Insert(i, i)
	}
}
