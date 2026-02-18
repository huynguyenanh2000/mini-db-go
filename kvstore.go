package main

type KV struct {
	fileName string
	tree     BPTreeDisk
}

func (db *KV) Open() {
	db.tree = NewBPTreeDisk(db.fileName)
}

func (db *KV) Get(key []byte) ([]byte, bool) {
	result := db.tree.Find(key)
	if result == nil {
		var valueBytes []byte = make([]byte, 0)
		return valueBytes, false
	}

	var valueBytes []byte = make([]byte, result.vallen)
	for i := 0; i < int(result.vallen); i++ {
		valueBytes[i] = result.val[i+(MAX_VAL_SIZE-int(result.vallen))]
	}
	return valueBytes, true
}

func (db *KV) Set(key, val []byte) {
	db.tree.Set(key, val)
}

func (db *KV) Del(key []byte) bool {
	return db.tree.Del(key)
}
