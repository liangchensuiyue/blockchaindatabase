package blockchain

import (
	"crypto/md5"
	"encoding/hex"
	"sync"
)

type DictEntry struct {
	Key   string
	Value *Transaction
	Next  *DictEntry
}
type HT struct {
	Table  map[string]*DictEntry
	Size   int64 // 哈希表大小
	Used   int64 //哈希表已有节点数量
	Rowlen map[string]int64
	lock   *sync.Mutex
}

func GetMd5(str string) string {
	m := md5.New()
	m.Write([]byte(str))
	c := m.Sum(nil)
	return hex.EncodeToString(c)
}
func (ht *HT) Init() {
	ht.Table = make(map[string]*DictEntry)
	ht.Rowlen = make(map[string]int64)
	// var i int64
	// for i = 0; i < ht.Size; i++ {
	// 	ht.Table[i] = nil
	// 	ht.Rowlen[i] = 0
	// }
	ht.lock = &sync.Mutex{}
}
func (ht *HT) Put(k string, v *Transaction) {
	ht.lock.Lock()
	head := ht.Table[k]
	ht.Used++
	if head == nil {
		ht.Rowlen[k] = 0
		ht.Table[k] = &DictEntry{
			Key:   k,
			Value: v,
			Next:  nil,
		}
	} else {
		ht.Rowlen[k] = 0
		ht.Table[k] = &DictEntry{
			Key:   k,
			Value: v,
			Next:  head,
		}
	}
	ht.lock.Unlock()
}

// func (ht *HT) Put(k string, v *Transaction) {
// 	md5Str := GetMd5(k) //取得md5
// 	tempsubstr := md5Str[:16]
// 	hexVal, _ := strconv.ParseInt(tempsubstr, 16, 64) //生成
// 	idx := hexVal % ht.Size

// 	head := ht.Table[idx]
// 	ht.lock.Lock()
// 	ht.Used++
// 	if head == nil {
// 		ht.Rowlen[idx]++
// 		ht.Table[idx] = &DictEntry{
// 			Key:   k,
// 			Value: v,
// 			Next:  nil,
// 		}
// 	} else {
// 		ht.Rowlen[idx]++
// 		ht.Table[idx] = &DictEntry{
// 			Key:   k,
// 			Value: v,
// 			Next:  head,
// 		}
// 	}
// 	ht.lock.Unlock()
// }
func (ht *HT) Shrink() {
	ht.lock.Lock()
	var head *DictEntry
	for k, _ := range ht.Table {
		head = ht.Table[k]
		if ht.Rowlen[k] <= 100 {
			continue
		}
		num := (ht.Rowlen[k] * 2) / 3
		for num <= 0 {
			head = head.Next
			num--
		}
		head.Next = nil
	}
	ht.lock.Unlock()
}

// func (ht *HT) Shrink() {
// 	ht.lock.Lock()
// 	var i int64
// 	var head *DictEntry
// 	for i = 0; i < ht.Size; i++ {
// 		head = ht.Table[i]
// 		if ht.Rowlen[i] <= 100 {
// 			continue
// 		}
// 		num := (ht.Rowlen[i] * 2) / 3
// 		for num <= 0 {
// 			head = head.Next
// 			num--
// 		}
// 		head.Next = nil
// 	}
// 	ht.lock.Unlock()
// }
func (ht *HT) Get(k string) []*Transaction {
	head, ok := ht.Table[k]
	if !ok || head == nil {
		return []*Transaction{}
	}
	rs := []*Transaction{}
	for head != nil {
		rs = append(rs, head.Value)
		head = head.Next
	}
	return rs
}

// func (ht *HT) Get(k string) []*Transaction {
// 	md5Str := GetMd5(k) //取得md5
// 	tempsubstr := md5Str[:16]
// 	hexVal, _ := strconv.ParseInt(tempsubstr, 16, 64) //生成
// 	idx := hexVal % ht.Size

// 	head := ht.Table[idx]
// 	if head == nil {
// 		return []*Transaction{}
// 	}
// 	rs := []*Transaction{}
// 	for head != nil {
// 		rs = append(rs, head.Value)
// 		head = head.Next
// 	}
// 	return rs
// }
func (ht *HT) Traverse(handler func(*DictEntry)) {
	var head *DictEntry
	for _, v := range ht.Table {
		head = v
		for head != nil {
			handler(head)
			head = head.Next
		}
	}
}

// func (ht *HT) Traverse(handler func(*DictEntry)) {
// 	var i int64
// 	var head *DictEntry
// 	for i = 0; i < ht.Size; i++ {
// 		head = ht.Table[i]
// 		for head != nil {
// 			handler(head)
// 			head = head.Next
// 		}
// 	}
// }
func init() {

}
