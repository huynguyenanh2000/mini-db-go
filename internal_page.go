package main

import (
	"bytes"
	"encoding/binary"
)

// 0: Meta Page
// 1: Internal Page
// 2: Leaf Page
// ...: not support
type PageHeader struct {
	pageType        uint8
	nextPagePointer uint64
}

// Manual read and write byte
func (header *PageHeader) writeToBuffer(buffer *bytes.Buffer) {
	err := binary.Write(buffer, binary.BigEndian, header.pageType)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buffer, binary.BigEndian, header.nextPagePointer)
	if err != nil {
		panic(err)
	}
}

func (header *PageHeader) readFromBuffer(buffer *bytes.Buffer) {
	err := binary.Read(buffer, binary.BigEndian, &header.pageType)
	if err != nil {
		panic(err)
	}
	err = binary.Read(buffer, binary.BigEndian, &header.nextPagePointer)
	if err != nil {
		panic(err)
	}
}

// ==============================================================================================================================

type MetaPage struct {
	header PageHeader
	// checksum,...
}

// Manual read and write byte
func (page *MetaPage) writeToBuffer(buffer *bytes.Buffer) {
	page.header.writeToBuffer(buffer)
}

func (page *MetaPage) readFromBuffer(buffer *bytes.Buffer) {
	page.header.readFromBuffer(buffer)
}

// ================================================================================================================================

const MAX_KEY_SIZE = 8

type KeyEntry struct {
	len  uint16
	data [MAX_KEY_SIZE]uint8
}

func NewKeyEntryFromInt(input int64) KeyEntry {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, input)
	dataSlice := buf.Bytes()
	dataLen := len(dataSlice)
	var data [MAX_KEY_SIZE]uint8
	for i := MAX_KEY_SIZE - dataLen; i < MAX_KEY_SIZE; i++ {
		data[i] = dataSlice[i-(MAX_KEY_SIZE-dataLen)]
	}
	return KeyEntry{
		len:  uint16(dataLen),
		data: data,
	}
}

func (key *KeyEntry) writeToBuffer(buffer *bytes.Buffer) {
	err := binary.Write(buffer, binary.BigEndian, key.len)
	if err != nil {
		panic(err)
	}
	for i := MAX_KEY_SIZE - key.len; i < MAX_KEY_SIZE; i++ {
		err := binary.Write(buffer, binary.BigEndian, key.data[i])
		if err != nil {
			panic(err)
		}
	}
}

func (key *KeyEntry) readFromBuffer(buffer *bytes.Buffer) {
	err := binary.Read(buffer, binary.BigEndian, &key.len)
	if err != nil {
		panic(err)
	}
	for i := MAX_KEY_SIZE - key.len; i < MAX_KEY_SIZE; i++ {
		err := binary.Read(buffer, binary.BigEndian, &key.data[i])
		if err != nil {
			panic(err)
		}
	}
}

func (key *KeyEntry) compare(rhs *KeyEntry) int {
	res := 0
	for i := range MAX_KEY_SIZE {
		if key.data[i] < rhs.data[i] {
			return -1
		}

		if key.data[i] > rhs.data[i] {
			return 1
		}
	}

	return res
}

// ================================================================================================================================

// [header | u8 u8 | k0 k1 k2 ... | 0 0 0 0 0 0]
type BTreeInternalPage struct {
	header    PageHeader
	nkey      uint16
	keys      [INTERNAL_MAX_KEYS]KeyEntry
	childrens [INTERNAL_MAX_KEYS]uint64
}

func (page *BTreeInternalPage) writeToBuffer(buffer *bytes.Buffer) {
	page.header.writeToBuffer(buffer)
	err := binary.Write(buffer, binary.BigEndian, page.nkey)
	if err != nil {
		panic(err)
	}
	for i := 0; i < int(page.nkey); i++ {
		page.keys[i].writeToBuffer(buffer)
	}
	for i := 0; i < int(page.nkey); i++ {
		err := binary.Write(buffer, binary.BigEndian, page.childrens[i])
		if err != nil {
			panic(err)
		}
	}
}

func (page *BTreeInternalPage) readFromBuffer(buffer *bytes.Buffer, isReadHeader bool) {
	if isReadHeader {
		page.header.readFromBuffer(buffer)
	}
	err := binary.Read(buffer, binary.BigEndian, &page.nkey)
	if err != nil {
		panic(err)
	}
	for i := 0; i < int(page.nkey); i++ {
		page.keys[i].readFromBuffer(buffer)
	}
	for i := 0; i < int(page.nkey); i++ {
		err := binary.Read(buffer, binary.BigEndian, &page.childrens[i])
		if err != nil {
			panic(err)
		}
	}
}

func NewIPage() BTreeInternalPage {
	var newKeys [INTERNAL_MAX_KEYS]KeyEntry
	var newChildren [INTERNAL_MAX_KEYS]uint64
	return BTreeInternalPage{
		nkey:      0,
		keys:      newKeys,
		childrens: newChildren,
		header: PageHeader{
			pageType:        1,
			nextPagePointer: 0,
		},
	}
}

// FindLaststLE finds the last key less than or equal to the given key.
func (node *BTreeInternalPage) FindLaststLE(findKey *KeyEntry) int {
	pos := -1
	for i := 0; i < int(node.nkey); i++ {
		if node.keys[i].compare(findKey) <= 0 {
			pos = i
		}
	}
	return pos
}

func (node *BTreeInternalPage) InsertKV(insertKey *KeyEntry, insertChildPtr uint64) {
	// Find last less or equal as position to insert
	pos := node.FindLaststLE(insertKey)

	for i := int(node.nkey) - 1; i > pos; i-- {
		node.keys[i+1] = node.keys[i]
		node.childrens[i+1] = node.childrens[i]
	}

	node.keys[pos+1] = *insertKey
	node.childrens[pos+1] = insertChildPtr
	node.nkey++
}

// Split a node into 2 nodes
func (node *BTreeInternalPage) Split() BTreeInternalPage {
	var newKeys [INTERNAL_MAX_KEYS]KeyEntry
	var newChildren [INTERNAL_MAX_KEYS]uint64
	// Split in the middle
	pos := node.nkey / 2
	for i := pos; i < node.nkey; i++ {
		newKeys[i-pos] = node.keys[i]
		newChildren[i-pos] = node.childrens[i]
		node.keys[i] = KeyEntry{}
		node.childrens[i] = 0
	}

	newNode := BTreeInternalPage{
		nkey:      node.nkey - pos,
		keys:      newKeys,
		childrens: newChildren,
	}
	node.nkey = pos
	return newNode
}
