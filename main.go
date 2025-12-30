package main

import "fmt"

const INTERNAL_MAX_KEYS = 4

type Node interface {
	// FindLaststLE(findKey int) int
	// InsertKV(insertKey int, insertChild Node)
	// Split() BTreeInternalNode
}

type BTreeInternalNode struct {
	nkey      int
	keys      [INTERNAL_MAX_KEYS]int
	childrens [INTERNAL_MAX_KEYS]*Node
}

func NewINode() BTreeInternalNode {
	var newKeys [INTERNAL_MAX_KEYS]int
	var newChildren [INTERNAL_MAX_KEYS]*Node
	return BTreeInternalNode{
		nkey:      0,
		keys:      newKeys,
		childrens: newChildren,
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
		node.childrens[i+1] = node.childrens[i]
	}

	node.keys[pos+1] = insertKey
	node.childrens[pos+1] = &insertChild
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
		newChildren[i-pos] = node.childrens[i]
		node.keys[i] = 0
		node.childrens[i] = nil
	}

	newNode := BTreeInternalNode{
		nkey:      node.nkey - pos,
		keys:      newKeys,
		childrens: newChildren,
	}
	node.nkey = pos
	return newNode
}

// Define leaf node
type BTreeLeafNode struct {
	nkey   int
	keys   [INTERNAL_MAX_KEYS]int
	values [INTERNAL_MAX_KEYS]int
}

func NewLNode() BTreeLeafNode {
	var newKeys [INTERNAL_MAX_KEYS]int
	var newVals [INTERNAL_MAX_KEYS]int
	return BTreeLeafNode{
		nkey:   0,
		keys:   newKeys,
		values: newVals,
	}
}

// FindLaststLE finds the last key less than or equal to the given key.
func (node *BTreeLeafNode) FindLaststLE(findKey int) int {
	pos := -1
	for i := 0; i < node.nkey; i++ {
		if node.keys[i] <= findKey {
			pos = i
		}
	}
	return pos
}

func (node *BTreeLeafNode) InsertKV(insertKey int, insertValue int) {
	// Find last less or equal as position to insert
	pos := node.FindLaststLE(insertKey)

	for i := node.nkey - 1; i > pos; i-- {
		node.keys[i+1] = node.keys[i]
		node.values[i+1] = node.values[i]
	}

	node.keys[pos+1] = insertKey
	node.values[pos+1] = insertValue
	node.nkey++
}

// Split a node into 2 nodes
func (node *BTreeLeafNode) Split() BTreeLeafNode {
	var newKeys [INTERNAL_MAX_KEYS]int
	var newValues [INTERNAL_MAX_KEYS]int
	// Split in the middle
	pos := node.nkey / 2
	for i := pos; i < node.nkey; i++ {
		newKeys[i-pos] = node.keys[i]
		newValues[i-pos] = node.values[i]
		node.keys[i] = 0
		node.values[i] = 0
	}

	newNode := BTreeLeafNode{
		nkey:   node.nkey - pos,
		keys:   newKeys,
		values: newValues,
	}
	node.nkey = pos
	return newNode
}

// B++ tree structure
type BPTree struct {
	head Node
}

func NewBPTree() BPTree {
	newINode := NewINode()
	return BPTree{
		head: &newINode,
	}
}

// Insert a key value pair
// After inserting, check if need split
// If need split, insert back to parent
func (tree *BPTree) insertRecursive(node Node, insertKey int, insertValue int) Node {
	if convertedNode, ok := node.(*BTreeInternalNode); ok {
		pos := convertedNode.FindLaststLE(insertKey)
		if convertedNode.nkey == 0 {
			firstLeaf := NewLNode()
			firstLeaf.InsertKV(insertKey, insertValue)
			convertedNode.InsertKV(insertKey, &firstLeaf)
		} else {
			if pos == -1 {
				pos = 0
			}
			child := convertedNode.childrens[pos]
			insertResult := tree.insertRecursive(*child, insertKey, insertValue)

			if convertedChild, ok := (*child).(*BTreeLeafNode); ok {
				convertedNode.keys[0] = convertedChild.keys[0]
			} else {
				convertedChild := (*child).(*BTreeInternalNode)
				convertedNode.keys[0] = convertedChild.keys[0]
			}

			if insertResult != nil {
				if convertedChild, ok := insertResult.(*BTreeLeafNode); ok {
					convertedNode.InsertKV(convertedChild.keys[0], convertedChild)
				} else {
					convertedChild := insertResult.(*BTreeInternalNode)
					convertedNode.InsertKV(convertedChild.keys[0], convertedChild)
				}
			}

			// After insert, check if need split
			if convertedNode.nkey == INTERNAL_MAX_KEYS {
				newInternal := convertedNode.Split()
				return &newInternal
			}
		}
	} else {
		convertedNode := node.(*BTreeLeafNode)
		convertedNode.InsertKV(insertKey, insertValue)

		// Check need split
		if convertedNode.nkey == INTERNAL_MAX_KEYS {
			// Split
			newLeaf := convertedNode.Split()
			return &newLeaf
		}
	}
	return nil
}

func (tree *BPTree) Insert(insertKey int, insertValue int) {
	insertResult := tree.insertRecursive(tree.head, insertKey, insertValue)
	if insertResult != nil {
		convertedChild := insertResult.(*BTreeInternalNode)
		newHead := NewINode()
		newHead.nkey = 2
		newHead.keys[0] = tree.head.(*BTreeInternalNode).keys[0]
		newHead.keys[1] = convertedChild.keys[0]
		newHead.childrens[0] = tree.head.(*BTreeInternalNode).childrens[0]
		newHead.childrens[1] = convertedChild.childrens[0]
		tree.head = &newHead
	}
}

func main() {
	fmt.Println("B-Tree Internal Node Example")
}
