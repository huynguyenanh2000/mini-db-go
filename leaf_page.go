package main

import (
	"bytes"
	"encoding/binary"
)

const MAX_VAL_SIZE = 8

type KeyVal struct {
	keylen uint16
	vallen uint16
	key    [MAX_KEY_SIZE]uint8
	val    [MAX_VAL_SIZE]uint8
}

func NewKeyValFromInt(inputKey int64, inputVal int64) KeyVal {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, inputKey)
	dataSlice := buf.Bytes()
	keyLen := len(dataSlice)
	var key [MAX_KEY_SIZE]uint8
	for i := MAX_KEY_SIZE - keyLen; i < MAX_KEY_SIZE; i++ {
		key[i] = dataSlice[i-(MAX_KEY_SIZE-keyLen)]
	}
	buf = new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, inputKey)
	dataSlice = buf.Bytes()
	valLen := len(dataSlice)
	var val [MAX_VAL_SIZE]uint8
	for i := MAX_VAL_SIZE - valLen; i < MAX_VAL_SIZE; i++ {
		key[i] = dataSlice[i-(MAX_VAL_SIZE-valLen)]
	}
	return KeyVal{
		keylen: uint16(keyLen),
		vallen: uint16(valLen),
		key:    key,
		val:    val,
	}
}

func (key *KeyVal) writeToBuffer(buffer *bytes.Buffer) {
	err := binary.Write(buffer, binary.BigEndian, key.keylen)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buffer, binary.BigEndian, key.vallen)
	if err != nil {
		panic(err)
	}
	for i := MAX_KEY_SIZE - key.keylen; i < MAX_KEY_SIZE; i++ {
		err := binary.Write(buffer, binary.BigEndian, key.key[i])
		if err != nil {
			panic(err)
		}
	}
	for i := MAX_VAL_SIZE - key.vallen; i < MAX_VAL_SIZE; i++ {
		err := binary.Write(buffer, binary.BigEndian, key.val[i])
		if err != nil {
			panic(err)
		}
	}
}

func (key *KeyVal) readFromBuffer(buffer *bytes.Buffer) {
	err := binary.Read(buffer, binary.BigEndian, &key.keylen)
	if err != nil {
		panic(err)
	}
	err = binary.Read(buffer, binary.BigEndian, &key.vallen)
	if err != nil {
		panic(err)
	}
	for i := MAX_KEY_SIZE - key.keylen; i < MAX_KEY_SIZE; i++ {
		err = binary.Read(buffer, binary.BigEndian, &key.key[i])
		if err != nil {
			panic(err)
		}
	}
	for i := MAX_VAL_SIZE - key.vallen; i < MAX_VAL_SIZE; i++ {
		err = binary.Read(buffer, binary.BigEndian, &key.val[i])
		if err != nil {
			panic(err)
		}
	}
}

func (key *KeyVal) compare(rhs *KeyVal) int {
	res := 0
	for i := range MAX_KEY_SIZE {
		if key.key[i] < rhs.key[i] {
			return -1
		}
		if key.key[i] > rhs.key[i] {
			return 1
		}
	}

	return res
}

// =======================================================

// Define leaf node
type BTreeLeafPage struct {
	header PageHeader
	nkv    int
	kv     [LEAF_MAX_KV]KeyVal
}

func NewLPage() BTreeLeafPage {
	var newKV [LEAF_MAX_KV]KeyVal
	return BTreeLeafPage{
		header: PageHeader{
			pageType:        2,
			nextPagePointer: 0,
		},
		nkv: 0,
		kv:  newKV,
	}
}

func (page *BTreeLeafPage) writeToBuffer(buffer *bytes.Buffer) {
	page.header.writeToBuffer(buffer)
	err := binary.Write(buffer, binary.BigEndian, page.nkv)
	if err != nil {
		panic(err)
	}
	for i := 0; i < int(page.nkv); i++ {
		page.kv[i].writeToBuffer(buffer)
	}
}

func (page *BTreeLeafPage) readFromBuffer(buffer *bytes.Buffer, isReadHeader bool) {
	if isReadHeader {
		page.header.readFromBuffer(buffer)
	}
	err := binary.Read(buffer, binary.BigEndian, page.nkv)
	if err != nil {
		panic(err)
	}
	for i := 0; i < int(page.nkv); i++ {
		page.kv[i].readFromBuffer(buffer)
	}
}

func (node *BTreeLeafPage) FindLastLE(findKV *KeyVal) int {
	pos := -1
	for i := 0; i < int(node.nkv); i++ {
		if node.kv[i].compare(findKV) <= 0 {
			pos = i
		}
	}
	return pos
}

func (node *BTreeLeafPage) InsertKV(insertKV *KeyVal) {
	pos := node.FindLastLE(insertKV)
	for i := node.nkv - 1; i > pos; i-- {
		node.kv[i+1] = node.kv[i]
	}
	node.kv[pos+1] = *insertKV
	node.nkv += 1
}

func (node *BTreeLeafPage) Split() BTreeLeafPage {
	var newKV [LEAF_MAX_KV]KeyVal
	pos := node.nkv / 2
	for i := pos; i < node.nkv; i++ {
		newKV[i-pos] = node.kv[i]
		node.kv[i] = KeyVal{}
	}
	newNode := BTreeLeafPage{
		nkv: node.nkv - pos,
		kv:  newKV,
	}
	node.nkv = pos
	return newNode
}
