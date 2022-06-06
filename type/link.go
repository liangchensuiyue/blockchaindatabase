package Type

import (
	"go_code/基于区块链的非关系型数据库/util"
)

type LinkNode struct {
	Prev  *LinkNode
	Next  *LinkNode
	Value interface{}
}
type LinkHead struct {
	Head *LinkNode
	Tail *LinkNode
}

func (l *LinkHead) Exsist(handler func(interface{}) bool) bool {
	p := l.Head
	for p != nil {
		if handler(p.Value) {
			return true
		}
		p = p.Next
	}
	return false
}
func (l *LinkHead) Insert(v interface{}) {
	node := &LinkNode{nil, nil, v}
	if l.Head == nil {
		l.Head = node
		l.Tail = node
		return
	}
	node.Next = l.Head
	l.Head.Prev = node
	l.Head = node
}
func (l *LinkHead) Delete(handler func(interface{}) bool) {
	pre := l.Head
	cur := pre
	for cur != nil && handler(cur.Value) {
		cur = cur.Next
	}
	l.Head = cur
	if cur == nil {
		return
	}
	pre = cur
	cur = pre.Next
	for cur != nil {
		if handler(cur.Value) {
			pre.Next = cur.Next
			cur.Next.Prev = pre
			continue
		}
		pre = cur
		cur = cur.Next
	}
	l.Tail = pre

}
func ByteToLink(data []byte, handler func([]byte) interface{}) *LinkHead {
	var length int64 = int64(len(data))
	if length == 0 {
		return nil
	}
	var head *LinkHead = &LinkHead{nil, nil}
	var cur int64 = 0
	size := util.BytesToInt64(data[cur : cur+8])
	h := &LinkNode{}
	h.Prev = nil
	h.Next = nil
	h.Value = handler(data[cur+8 : cur+8+size])
	head.Head = h
	cur = cur + 8 + size
	for cur < length {
		size := util.BytesToInt64(data[cur : cur+8])
		temp := &LinkNode{}
		temp.Prev = h
		temp.Next = nil
		temp.Value = handler(data[cur+8 : cur+8+size])
		h.Next = temp
		cur = cur + 8 + size
		h = temp
	}
	head.Tail = h
	return head
}
func LinkToByte(head *LinkHead, handler func(interface{}) []byte) []byte {
	if head.Head == nil {
		return []byte{}
	}
	data := []byte{}
	cur := head.Head
	for cur != nil {
		v := handler(cur.Value)
		size := int64(len(v))
		data = append(data, util.Int64Tobyte(size)...)
		data = append(data, v...)
		cur = cur.Next
	}
	return data
}

// 修改密码可以看作: 先删除用户，然后添加用户
