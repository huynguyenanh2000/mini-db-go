package main

import (
	"bytes"
	"os"
)

// ====================================File allocator=====================================
type FileAllocator struct {
	lastFree  uint64
	freeBlock []uint64
}

// Always return a pointer on disk to write data to
// <= 4096 bytes -> increased by 4-96
func (a *FileAllocator) alloc() uint64 {
	if len(a.freeBlock) == 0 {
		ptr := a.lastFree * 4096
		a.lastFree += 1
		return ptr
	}
	ptr := a.freeBlock[0] * 4096
	a.freeBlock = a.freeBlock[1:]
	return ptr
}

func (a *FileAllocator) free(ptr uint64) {
	a.freeBlock = append(a.freeBlock, ptr/4096)
}

func (a *FileAllocator) writeAllToFile(file *os.File) {}

func LoadFileAllocator(fileName string) FileAllocator {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	return FileAllocator{}
}

// ======================================== B++ tree structure ================================
type BPTreeDisk struct {
	fileName      string
	fileAllocator FileAllocator
}

func NewBPTreeDisk(fileName string) BPTreeDisk {
	return BPTreeDisk{
		fileName: fileName,
		fileAllocator: FileAllocator{
			lastFree:  1,
			freeBlock: []uint64{},
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

type DelResult struct {
	nodePtr      uint64
	nodePromoKey KeyEntry
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
func (tree *BPTreeDisk) insertRecursive(node any, insertKey *KeyEntry, insertKV *KeyVal, buffer *bytes.Buffer, file *os.File, deletedPtr []uint64) InsertResult {
	if convertedNode, ok := node.(*BTreeInternalPage); ok {
		pos := convertedNode.FindLastLE(insertKey)
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
			child := convertedNode.childrens[pos]
			tree.readBlockAtPointer(child, buffer, file)
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

			insertResult := tree.insertRecursive(childNode, insertKey, insertKV, buffer, file, deletedPtr)
			convertedNode.keys[pos] = insertResult.nodePromoKey
			deletedPtr = append(deletedPtr, convertedNode.childrens[pos])
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
				// Can free old block

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

func (tree *BPTreeDisk) Insert(insertKeyBytes, insertValueBytes []byte) {
	buffer := new(bytes.Buffer) // Buffer size = 0
	insertKey := NewKeyEntryFromBytes(insertKeyBytes)
	insertVal := NewKeyValFromBytes(insertKeyBytes, insertValueBytes)

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

	deletedPtr := make([]uint64, 0)

	// Step 4: Insert sub structure
	insertResult := tree.insertRecursive(&internalPage, &insertKey, &insertVal, buffer, file, deletedPtr)

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
	if metaPage.header.nextPagePointer != 0 {
		deletedPtr = append(deletedPtr, metaPage.header.nextPagePointer)
	}
	metaPage.header.nextPagePointer = firstInternalPagePtr
	buffer.Reset()
	metaPage.writeToBuffer(buffer)
	tree.writeBufferToFileFirst(buffer, file)
	// Defragment
	for _, x := range deletedPtr {
		tree.fileAllocator.free(x)
	}
}

func (tree *BPTreeDisk) Find(key []byte) *KeyVal {
	buffer := new(bytes.Buffer)
	findKeyEntry := NewKeyEntryFromBytes(key)
	var emptyVal []byte = make([]byte, 0)
	findKeyVal := NewKeyValFromBytes(key, emptyVal)
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

	var node any
	node = &internalPage

	for {
		if convertedNode, ok := node.(*BTreeInternalPage); ok {
			pos := convertedNode.FindLastLE(&findKeyEntry)
			if pos == -1 {
				return nil
			}
			child := convertedNode.childrens[pos]
			tree.readBlockAtPointer(child, buffer, file)
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
			node = childNode
		} else {
			convertedNode := node.(*BTreeLeafPage)
			pos := convertedNode.FindLastLE(&findKeyVal)
			if pos == -1 {
				return nil
			}
			foundKeyVal := convertedNode.kv[pos]
			if foundKeyVal.compare(&findKeyVal) == 0 {
				return &foundKeyVal
			}
			return nil
		}
	}
}

// assume key can be found always
func (tree *BPTreeDisk) setRecursive(node any, setKey *KeyEntry, setKV *KeyVal, buffer *bytes.Buffer, file *os.File) InsertResult {
	if convertedNode, ok := node.(*BTreeInternalPage); ok {
		pos := convertedNode.FindLastLE(setKey) // always have
		child := convertedNode.childrens[pos]
		tree.readBlockAtPointer(child, buffer, file)
		header := PageHeader{}
		header.readFromBuffer(buffer)
		var childNode any
		if header.pageType == 1 {
			ipage := BTreeInternalPage{header: header}
			ipage.readFromBuffer(buffer, false)
			childNode = &ipage
		} else {
			lpage := BTreeLeafPage{header: header}
			lpage.readFromBuffer(buffer, false)
			childNode = &lpage
		}

		setResult := tree.setRecursive(childNode, setKey, setKV, buffer, file)
		convertedNode.childrens[pos] = setResult.nodePtr

		buffer.Reset()
		convertedNode.writeToBuffer(buffer)
		oldPtr := tree.writeBufferToFile(buffer, file)
		return InsertResult{
			nodePtr:      oldPtr,
			nodePromoKey: convertedNode.keys[0],
			newNodePtr:   0,
			newPromoKey:  KeyEntry{},
		}
	} else {
		convertedNode := node.(*BTreeLeafPage)
		pos := convertedNode.FindLastLE(setKV)
		convertedNode.kv[pos] = *setKV

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

func (tree *BPTreeDisk) Set(setKeyBytes []byte, setValueBytes []byte) {
	findRes := tree.Find(setKeyBytes)
	if findRes == nil {
		tree.Insert(setKeyBytes, setValueBytes)
		return
	}

	buffer := new(bytes.Buffer)
	setKey := NewKeyEntryFromBytes(setKeyBytes)
	setKV := NewKeyValFromBytes(setKeyBytes, setValueBytes)

	// Step 1: Open file
	file, err := os.OpenFile(tree.fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Step 2: Read MetaPage
	tree.readBlockAtPointer(0, buffer, file)
	metaPage := MetaPage{}
	metaPage.readFromBuffer(buffer)

	internalPage := BTreeInternalPage{}
	// Step 3: Read first internal page
	if metaPage.header.nextPagePointer != 0 {
		tree.readBlockAtPointer(metaPage.header.nextPagePointer, buffer, file)
		internalPage.readFromBuffer(buffer, true)
	}

	// Step 4: Set sub structure
	setResult := tree.setRecursive(&internalPage, &setKey, &setKV, buffer, file)

	// Step 5: Modify MetaPage and save to disk
	firstInternalPagePtr := setResult.nodePtr
	// Assume last step has the first internal page ptr
	metaPage.header.nextPagePointer = firstInternalPagePtr
	buffer.Reset()
	metaPage.writeToBuffer(buffer)
}

func (tree *BPTreeDisk) delRecursive(node any, insertKey *KeyEntry, insertKV *KeyVal, buffer *bytes.Buffer, file *os.File) DelResult {
	if convertedNode, ok := node.(*BTreeInternalPage); ok {
		pos := convertedNode.FindLastLE(insertKey)
		child := convertedNode.childrens[pos]
		tree.readBlockAtPointer(child, buffer, file)

		header := PageHeader{}
		header.readFromBuffer(buffer)
		var childNode any
		if header.pageType == 1 {
			iPage := BTreeInternalPage{header: header}
			iPage.readFromBuffer(buffer, false)
			childNode = &iPage
		} else {
			lPage := BTreeLeafPage{header: header}
			lPage.readFromBuffer(buffer, false)
			childNode = &lPage
		}

		delResult := tree.delRecursive(childNode, insertKey, insertKV, buffer, file)
		if delResult.nodePtr == 0 {
			convertedNode.DelKVAtPos(pos)
			if convertedNode.nkey == 0 {
				return DelResult{
					nodePtr:      0,
					nodePromoKey: KeyEntry{},
				}
			}
		} else {
			convertedNode.keys[pos] = delResult.nodePromoKey
			convertedNode.childrens[pos] = delResult.nodePtr
		}

		buffer.Reset()
		convertedNode.writeToBuffer(buffer)
		oldPtr := tree.writeBufferToFile(buffer, file)
		return DelResult{
			nodePtr:      oldPtr,
			nodePromoKey: convertedNode.keys[0],
		}
	} else {
		convertedNode := node.(*BTreeLeafPage)
		convertedNode.DelKV(insertKV)
		if convertedNode.nkv == 0 {
			return DelResult{
				nodePtr:      0,
				nodePromoKey: KeyEntry{},
			}
		}

		buffer.Reset()
		convertedNode.writeToBuffer(buffer)
		oldPtr := tree.writeBufferToFile(buffer, file)
		return DelResult{
			nodePtr:      oldPtr,
			nodePromoKey: getKeyEntryFromKeyVal(&convertedNode.kv[0]),
		}
	}
}

func (tree *BPTreeDisk) Del(key []byte) bool {
	findRes := tree.Find(key)
	if findRes == nil {
		return false
	}

	buffer := new(bytes.Buffer)
	delKeyE := NewKeyEntryFromBytes(key)
	var emptyVal []byte = make([]byte, 0)
	delKeyV := NewKeyValFromBytes(key, emptyVal)

	// Step 1: Open file
	file, err := os.OpenFile(tree.fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Step 2: Read metaPage
	tree.readBlockAtPointer(0, buffer, file)
	metaPage := MetaPage{}
	metaPage.readFromBuffer(buffer)

	internalPage := BTreeInternalPage{}
	// Step 3: Read first internal page
	if metaPage.header.nextPagePointer != 0 {
		tree.readBlockAtPointer(metaPage.header.nextPagePointer, buffer, file)
		internalPage.readFromBuffer(buffer, true)
	}

	// Step 4: Insert sub structure
	delResult := tree.delRecursive(&internalPage, &delKeyE, &delKeyV, buffer, file)
	// Step 5: Modify metapage and save to disk
	firstInternalPagePtr := delResult.nodePtr
	metaPage.header.nextPagePointer = firstInternalPagePtr
	buffer.Reset()
	metaPage.writeToBuffer(buffer)
	tree.writeBufferToFileFirst(buffer, file)
	return true
}
