package main

import (
	"bytes"
	"os"
)

// ====================================File allocator=====================================
type FileAllocator struct {
	lastPointer uint64
}

// Always return a pointer on disk to write data to
// <= 4096 bytes -> increased by 4-96
func (a *FileAllocator) alloc() uint64 {
	oldPointer := a.lastPointer
	a.lastPointer += 4096
	return oldPointer
}

// TODO: Free to reuse memory

// ======================================== B++ tree structure ================================
type BPTreeDisk struct {
	fileName      string
	fileAllocator FileAllocator
}

func NewBPTreeDisk(fileName string) BPTreeDisk {
	return BPTreeDisk{
		fileName: fileName,
		fileAllocator: FileAllocator{
			lastPointer: 4096,
		},
	}
}

// Reuse buffer style: buffer always of size 4096
func (tree *BPTreeDisk) readBlockAtPointer(ptr uint64, buffer *bytes.Buffer, file *os.File) {
	inbuf := make([]byte, 4096)
	_, err := file.ReadAt(inbuf, int64(ptr))
	if err != nil {
		panic(err)
	}
	buffer.Reset()
	buffer.Write(inbuf)
}

// Return a disk pointer to this data
func (tree *BPTreeDisk) writeBufferToFile(buffer *bytes.Buffer, file *os.File) uint64 {
	lastPtr := tree.fileAllocator.alloc()
	_, err := file.WriteAt(buffer.Bytes(), int64(lastPtr))
	if err != nil {
		panic(err)
	}
	return lastPtr
}

func (tree *BPTreeDisk) writeBufferToFileFirst(buffer *bytes.Buffer, file *os.File) {
	_, err := file.WriteAt(buffer.Bytes(), 0)
	if err != nil {
		panic(err)
	}
}

type InsertResult struct {
	nodePtr      uint64
	nodePromoKey KeyEntry
	newNodePtr   uint64 // Need to split, else 0
	newPromoKey  KeyEntry
}

func getKeyEntryFromKeyVal(kv *KeyVal) KeyEntry {
	return KeyEntry{
		len:  kv.keylen,
		data: kv.key,
	}
}

// Insert a key value pair
// After inserting, check if need split
// If need split, insert back to parent
func (tree *BPTreeDisk) insertRecursive(node any, insertKey *KeyEntry, insertKV *KeyVal, buffer *bytes.Buffer, file *os.File) InsertResult {
	if convertedNode, ok := node.(*BTreeInternalPage); ok {
		pos := convertedNode.FindLaststLE(insertKey)
		if convertedNode.nkey == 0 {
			firstLeaf := NewLPage()
			firstLeaf.InsertKV(insertKV)
			buffer.Reset()
			firstLeaf.writeToBuffer(buffer)
			leafPtr := tree.writeBufferToFile(buffer, file)
			convertedNode.InsertKV(insertKey, leafPtr)
		} else {
			if pos == -1 {
				pos = 0
			}
			// Try to convert back to either leaf or internal
			header := PageHeader{}
			header.readFromBuffer(buffer)
			var childNode any
			if header.pageType == 1 {
				// Internal page
				ipage := BTreeInternalPage{header: header}
				ipage.readFromBuffer(buffer, false)
				childNode = &ipage
			} else {
				// Leaf page
				lpage := BTreeLeafPage{header: header}
				lpage.readFromBuffer(buffer, false)
				childNode = &lpage
			}

			insertResult := tree.insertRecursive(childNode, insertKey, insertKV, buffer, file)
			convertedNode.keys[pos] = insertResult.nodePromoKey
			convertedNode.childrens[pos] = insertResult.nodePtr

			if insertResult.newNodePtr != 0 {
				convertedNode.InsertKV(&insertResult.newPromoKey, insertResult.newNodePtr)
			}

			// After insert, check if need split
			if convertedNode.nkey == INTERNAL_MAX_KEYS {
				newInternal := convertedNode.Split()
				// Save current page
				buffer.Reset()
				convertedNode.writeToBuffer(buffer)
				oldPtr := tree.writeBufferToFile(buffer, file)
				// Save new page
				buffer.Reset()
				newInternal.writeToBuffer(buffer)
				newPtr := tree.writeBufferToFile(buffer, file)
				return InsertResult{
					nodePtr:      oldPtr,
					nodePromoKey: convertedNode.keys[0],
					newNodePtr:   newPtr,
					newPromoKey:  newInternal.keys[0],
				}
			} else {
				// Save current page
				buffer.Reset()
				convertedNode.writeToBuffer(buffer)
				oldPtr := tree.writeBufferToFile(buffer, file)
				return InsertResult{
					nodePtr:      oldPtr,
					nodePromoKey: convertedNode.keys[0],
					newNodePtr:   0,
					newPromoKey:  KeyEntry{},
				}
			}
		}
	} else {
		convertedNode := node.(*BTreeLeafPage)
		convertedNode.InsertKV(insertKV)

		// Check need split
		if convertedNode.nkv == LEAF_MAX_KV {
			// Split
			newLeaf := convertedNode.Split()
			// Save current page
			buffer.Reset()
			convertedNode.writeToBuffer(buffer)
			oldPtr := tree.writeBufferToFile(buffer, file)
			// Save new page
			buffer.Reset()
			newLeaf.writeToBuffer(buffer)
			newPtr := tree.writeBufferToFile(buffer, file)
			return InsertResult{
				nodePtr:      oldPtr,
				nodePromoKey: getKeyEntryFromKeyVal(&convertedNode.kv[0]),
				newNodePtr:   newPtr,
				newPromoKey:  getKeyEntryFromKeyVal(&newLeaf.kv[0]),
			}
		} else {
			// Save current page
			buffer.Reset()
			convertedNode.writeToBuffer(buffer)
			oldPtr := tree.writeBufferToFile(buffer, file)
			return InsertResult{
				nodePtr:      oldPtr,
				nodePromoKey: getKeyEntryFromKeyVal(&convertedNode.kv[0]),
				newNodePtr:   0,
				newPromoKey:  KeyEntry{},
			}
		}
	}
	return InsertResult{}
}

func (tree *BPTreeDisk) Insert(insertKeyInt int, insertValueInt int) {
	buffer := new(bytes.Buffer) // Buffer size = 0
	insertKey := NewKeyEntryFromInt(int64(insertKeyInt))
	insertVal := NewKeyValFromInt(int64(insertKeyInt), int64(insertValueInt))

	// Step 1: Open file
	file, err := os.OpenFile(tree.fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer file.Close() // Persist

	// Step 2: Read MetaPage
	tree.readBlockAtPointer(0, buffer, file) // Buffer size = 4096
	metaPage := MetaPage{}
	metaPage.readFromBuffer(buffer) // Buffer size decrease

	internalPage := BTreeInternalPage{}
	// Step 3: Read first internal page
	if metaPage.header.nextPagePointer != 0 {
		tree.readBlockAtPointer(metaPage.header.nextPagePointer, buffer, file) // Buffer size = 4096
		internalPage.readFromBuffer(buffer, true)                              // Buffer size decrease
	}

	// Step 4: Insert sub structure
	insertResult := tree.insertRecursive(&internalPage, &insertKey, &insertVal, buffer, file)

	// Step 5: Modify MetaPage and save to disk
	var firstInternalPagePtr uint64
	if insertResult.newNodePtr != 0 {
		// Insert a new page
		newFirstIPage := NewIPage()
		newFirstIPage.nkey = 2
		newFirstIPage.keys[0] = insertResult.nodePromoKey
		newFirstIPage.childrens[0] = insertResult.nodePtr
		newFirstIPage.keys[1] = insertResult.newPromoKey
		newFirstIPage.childrens[1] = insertResult.newNodePtr
		buffer.Reset()
		newFirstIPage.writeToBuffer(buffer)
		firstInternalPagePtr = tree.writeBufferToFile(buffer, file)
	} else {
		firstInternalPagePtr = insertResult.nodePtr
	}

	// Assume last step has the last internal page ptr
	metaPage.header.nextPagePointer = firstInternalPagePtr
	buffer.Reset()
	metaPage.writeToBuffer(buffer)
	tree.writeBufferToFileFirst(buffer, file)
}
