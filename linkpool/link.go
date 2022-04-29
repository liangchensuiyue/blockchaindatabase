package linkpool

import (
	"bytes"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Node struct {
	IP           string
	ID           []byte
	CloseHandler func()
	DeadLine     time.Time
	Next         *Node
}
type LinkPool struct {
	Head           *Node
	ValidTime      int
	PeriodicalTime int
}

func (n *Node) GenerateId() {
	u1, _ := uuid.NewV4()
	n.ID = u1.Bytes()
}
func (n *Node) AddTimeToDeadline(s int) {
	n.DeadLine = n.DeadLine.Add(time.Duration(s) * time.Second)
}
func (n *Node) IsValid() bool {
	return n.DeadLine.After(time.Now())
}

func (pool *LinkPool) DisposeNode(ID []byte) {
	if pool.Head == nil && pool.Head.Next == nil {
		return
	}
	pre := pool.Head
	cur := pool.Head.Next
	for cur != nil {
		if bytes.Equal(ID, cur.ID) {
			pre.Next = cur.Next
			break
		}

		pre = cur
		cur = cur.Next
	}
}
func (pool *LinkPool) addNode(node *Node) {
	if pool.Head == nil {
		return
	}
	p := pool.Head
	for p.Next != nil {

		p = p.Next
	}
	p.Next = node
	node.Next = nil

}
func (pool *LinkPool) QueryNodeByID(id []byte) bool {
	p := pool.Head
	for p != nil {
		if bytes.Equal(p.ID, id) {
			return true
		}
	}
	return false

}
func (linkpool *LinkPool) AddNode(IP string, handler func()) []byte {
	newnode := &Node{
		IP:           IP,
		CloseHandler: handler,
		DeadLine:     time.Now().Add(time.Duration(linkpool.ValidTime) * time.Second),
		Next:         nil,
	}
	newnode.GenerateId()
	linkpool.addNode(newnode)
	return newnode.ID
}
func (linkpool *LinkPool) AddactiveTime(id []byte) {
	p := linkpool.Head
	for p != nil {
		if bytes.Equal(p.ID, id) {
			p.AddTimeToDeadline(linkpool.ValidTime)
		}
		p = p.Next
	}
}

var Global_Link_pool *LinkPool

func (pool *LinkPool) PeriodicalDipose() {
	if pool.Head == nil && pool.Head.Next == nil {
		return
	}
	pre := pool.Head
	cur := pool.Head.Next
	for cur != nil {
		if time.Now().After(cur.DeadLine) {
			pre.Next = cur.Next
			cur.CloseHandler()
			cur = cur.Next
			continue
		}

		pre = cur
		cur = cur.Next
	}
}
func init() {
	Global_Link_pool = &LinkPool{
		ValidTime:      60,
		PeriodicalTime: 30,
	}
	Global_Link_pool.Head = &Node{Next: nil}
	go func() {
		time.Sleep(time.Duration(Global_Link_pool.PeriodicalTime) * time.Second)
		Global_Link_pool.PeriodicalDipose()
	}()
}
