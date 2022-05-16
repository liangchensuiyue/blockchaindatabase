package blockchain

import (
	"bytes"
	"encoding/gob"
	"errors"

	"io/ioutil"
	"log"
	"sync"
)

var _lock1 *sync.Mutex = &sync.Mutex{}

type QueueObject struct {
	TargetBlock *Block
	Handle      func(int, int)
}
type node struct {
	Value QueueObject
	Next  *node
	Pre   *node
}
type Queue struct {
	front *node
	rear  *node
	num   int
	lock  *sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{nil, nil, 0, &sync.Mutex{}}
}
func (Q *Queue) InsertFront(v QueueObject) {
	Q.lock.Lock()
	defer Q.lock.Unlock()
	node := &node{v, nil, nil}
	if Q.num == 0 {
		Q.front = node
		Q.rear = node
		Q.num = 1
		return
	}
	Q.front.Pre = node
	node.Next = Q.front
	Q.front = node
	Q.num++
}
func (Q *Queue) Insert(v QueueObject) {
	Q.lock.Lock()
	defer Q.lock.Unlock()
	node := &node{v, nil, nil}
	if Q.num == 0 {
		Q.front = node
		Q.rear = node
		Q.num = 1
		return
	}
	Q.rear.Next = node
	node.Pre = Q.rear
	Q.rear = node
	Q.num++
	// Q.SaveToDisk()

}
func (Q *Queue) Delete() {
	Q.lock.Lock()
	if Q.num == 0 {
		return
	}
	if Q.num == 1 {
		Q.front = nil
		Q.rear = nil
		Q.num = 0
		Q.lock.Unlock()
		return
	}
	Q.front = Q.front.Next
	Q.front.Pre = nil
	Q.num--
	Q.lock.Unlock()
	Q.SaveToDisk()
}
func (Q *Queue) Load() {
	blocks := []QueueObject{}
	content, err := ioutil.ReadFile("./cachequeue")
	if err != nil {
		Q.SaveToDisk()
		return
	}

	decoder := gob.NewDecoder(bytes.NewReader(content))

	err = decoder.Decode(&blocks)
	if err != nil {
		log.Panic(err)
		return
	}
	Q.front = nil
	Q.rear = nil
	Q.num = 0
	// Q.lock.Lock()
	// defer Q.lock.Unlock()
	for _, v := range blocks {
		Q.Insert(v)
	}
}
func (Q *Queue) SaveToDisk() {
	blocks := []QueueObject{}
	if Q.num == 0 {
		return
	}
	r := Q.front
	for r != nil {
		blocks = append(blocks, r.Value)
		r = r.Next
	}

	var content bytes.Buffer

	// Curve 是一个接口类型

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(blocks)
	if err != nil {
		log.Panic(err)
	}
	_lock1.Lock()
	defer _lock1.Unlock()
	err = ioutil.WriteFile("./cachequeue", content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
func (Q *Queue) FindBlockByHash(hash []byte) (*Block, error) {
	if Q.num == 0 {
		return nil, errors.New("not found")
	}
	r := Q.front
	for r != nil {
		if bytes.Equal(r.Value.TargetBlock.Hash, hash) {
			return r.Value.TargetBlock, nil
		}
		r = r.Next
	}
	return nil, errors.New("not found")
}
func (Q *Queue) Front() (QueueObject, error) {
	if Q.num == 0 {
		return QueueObject{}, errors.New("nout found")
	}
	return Q.front.Value, nil
}
func (Q *Queue) Len() int {
	return Q.num
}
func (Q *Queue) Find(handle func(*Block) bool) (*Block, error) {
	if Q.num == 0 {
		return nil, errors.New("not found")
	}
	r := Q.rear
	for r != nil {
		if !handle(r.Value.TargetBlock) {
			return r.Value.TargetBlock, nil
		}
		r = r.Pre
	}
	return nil, errors.New("not found")
}
