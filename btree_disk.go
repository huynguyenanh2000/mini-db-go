package main

// B++ tree structure
type BPTreeDisk struct {
	head MetaPage
}

func NewBPTreeDisk() BPTreeDisk {
	return BPTreeDisk{
		head: MetaPage{
			header: PageHeader{
				pageType:        0,
				nextPagePointer: 0,
			},
		},
	}
}

type InsertResult struct {
	nodePtr      uint64
	nodePromoKey KeyEntry
	newNodePtr   uint64 // Need to split else 0
	newPromoKey  KeyEntry
}

// Insert a key value pair
// After inserting, check if need split
// If need split, insert back to parent
// func (tree *BPTreeDisk) insertRecursive(pagePtr uint64, insertKey KeyEntry, insertValue int) InsertResult {
// 	if convertedNode, ok := node.(*BTreeInternalNode); ok {
// 		pos := convertedNode.FindLaststLE(insertKey)
// 		if convertedNode.nkey == 0 {
// 			firstLeaf := NewLNode()
// 			firstLeaf.InsertKV(insertKey, insertValue)
// 			convertedNode.InsertKV(insertKey, &firstLeaf)
// 		} else {
// 			if pos == -1 {
// 				pos = 0
// 			}
// 			child := convertedNode.childrens[pos]
// 			insertResult := tree.insertRecursive(*child, insertKey, insertValue)

// 			if convertedChild, ok := (*child).(*BTreeLeafNode); ok {
// 				convertedNode.keys[0] = convertedChild.keys[0]
// 			} else {
// 				convertedChild := (*child).(*BTreeInternalNode)
// 				convertedNode.keys[0] = convertedChild.keys[0]
// 			}

// 			if insertResult != nil {
// 				if convertedChild, ok := insertResult.(*BTreeLeafNode); ok {
// 					convertedNode.InsertKV(convertedChild.keys[0], convertedChild)
// 				} else {
// 					convertedChild := insertResult.(*BTreeInternalNode)
// 					convertedNode.InsertKV(convertedChild.keys[0], convertedChild)
// 				}
// 			}

// 			// After insert, check if need split
// 			if convertedNode.nkey == INTERNAL_MAX_KEYS {
// 				newInternal := convertedNode.Split()
// 				return &newInternal
// 			}
// 		}
// 	} else {
// 		convertedNode := node.(*BTreeLeafNode)
// 		convertedNode.InsertKV(insertKey, insertValue)

// 		// Check need split
// 		if convertedNode.nkey == INTERNAL_MAX_KEYS {
// 			// Split
// 			newLeaf := convertedNode.Split()
// 			return &newLeaf
// 		}
// 	}
// 	return nil
// }

// func (tree *BPTreeDisk) Insert(insertKey int, insertValue int) {
// 	insertResult := tree.insertRecursive(tree.head, insertKey, insertValue)
// 	if insertResult != nil {
// 		convertedChild := insertResult.(*BTreeInternalNode)
// 		newHead := NewINode()
// 		newHead.nkey = 2
// 		newHead.keys[0] = tree.head.(*BTreeInternalNode).keys[0]
// 		newHead.keys[1] = convertedChild.keys[0]
// 		newHead.childrens[0] = tree.head.(*BTreeInternalNode).childrens[0]
// 		newHead.childrens[1] = convertedChild.childrens[0]
// 		tree.head = &newHead
// 	}
// }
