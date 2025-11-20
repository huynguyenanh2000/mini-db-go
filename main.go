package main

import "fmt"

const INTERNAL_MAX_KEYS = 4

type Node interface {
	// FindLaststLE(findKey int) int
	// InsertKV(insertKey int, insertChild Node)
	// Split() BTreeInternalNode
}

type BTreeInternalNode struct {
	nkey     int
	keys     [INTERNAL_MAX_KEYS]int
	children [INTERNAL_MAX_KEYS]*Node
}

func NewINode() BTreeInternalNode {
	var newKeys [INTERNAL_MAX_KEYS]int
	var newChildren [INTERNAL_MAX_KEYS]*Node
	return BTreeInternalNode{
		nkey:     0,
		keys:     newKeys,
		children: newChildren,
	}
}

// FindLaststLE finds the last key less than or equal to the given key.
func (node *BTreeInternalNode) FindLaststLE(findKey int) int {
	pos := -1
	for i := 0; i < node.nkey; i++ {
		if node.keys[i] <= findKey {
			pos = i
		}
	}
	return pos
}

func (node *BTreeInternalNode) InsertKV(insertKey int, insertChild Node) {
	// Find last less or equal as position to insert
	pos := node.FindLaststLE(insertKey)

	for i := node.nkey - 1; i > pos; i-- {
		node.keys[i+1] = node.keys[i]
		node.children[i+1] = node.children[i]
	}

	node.keys[pos+1] = insertKey
	node.children[pos+1] = &insertChild
	node.nkey++
}

// Split a node into 2 nodes
func (node *BTreeInternalNode) Split() BTreeInternalNode {
	var newKeys [INTERNAL_MAX_KEYS]int
	var newChildren [INTERNAL_MAX_KEYS]*Node

	// Split in the middle
	pos := node.nkey / 2
	for i := pos; i < node.nkey; i++ {
		newKeys[i-pos] = node.keys[i]
		newChildren[i-pos] = node.children[i]
		node.keys[i] = 0
		node.children[i] = nil
	}

	newNode := BTreeInternalNode{
		nkey:     node.nkey - pos,
		keys:     newKeys,
		children: newChildren,
	}
	node.nkey = pos
	return newNode
}

func main() {
	fmt.Println("B-Tree Internal Node Example")
}
