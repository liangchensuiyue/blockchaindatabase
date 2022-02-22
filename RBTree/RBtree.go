package rbtree

import (
	"fmt"
)

const (
	RED   = true
	BLACK = false
)

func test() {
	fmt.Println("sdf")
}

type RBNode struct {
	Left   *RBNode
	Right  *RBNode
	Parent *RBNode // 父亲节点
	Color  bool    // 颜色
	Item
}
type Item interface {
	Less(than Item) bool
}

type RBtree struct {
	NIL   *RBNode
	Root  *RBNode
	count uint
}

func less(x, y Item) bool {
	return x.Less(y)
}
func NewRBTree() *RBtree {
	return new(RBtree).Init()
}
func (rbt *RBtree) Init() *RBtree {
	node := &RBNode{nil, nil, nil, BLACK, nil}
	return &RBtree{node, node, 0}
}

// 获取红黑树的长度
func (rbt *RBtree) Len() uint {
	return rbt.count
}

func (rbt *RBtree) max(x *RBNode) *RBNode {
	if x == rbt.NIL {
		return rbt.NIL
	}
	for x.Right != rbt.NIL {
		x = x.Right
	}
	return x
}

func (rbt *RBtree) min(x *RBNode) *RBNode {
	if x == rbt.NIL {
		return rbt.NIL
	}
	for x.Left != rbt.NIL {
		x = x.Left
	}
	return x
}

func (rbt *RBtree) Search(item Item) *RBNode {
	return rbt.search(&RBNode{rbt.NIL, rbt.NIL, rbt.NIL, RED, item})
}

// 搜索红黑树
func (rbt *RBtree) search(x *RBNode) *RBNode {
	pnode := rbt.Root
	for pnode != rbt.NIL {
		if less(pnode.Item, x.Item) {
			pnode = pnode.Right
		} else if less(x.Item, pnode.Item) {
			pnode = pnode.Left
		} else {
			break // 找到
		}
	}
	if pnode == rbt.NIL {
		return rbt.NIL
	}
	if pnode == rbt.Root && !less(pnode, rbt.Root) && !less(rbt.Root, pnode) {
		return pnode
	}
	// if pnode == rbt.Root && !less(pnode.Item, rbt.Root.Item) && !less(rbt.Root.Item, pnode.Item) {
	// 	return pnode
	// }
	return pnode
}

func (rbt *RBtree) leftRotate(x *RBNode) {
	if x.Right == rbt.NIL || x.Right == nil {
		return // 右孩子不能为 0
	}
	y := x.Right
	// if x == nil || y == nil{
	// 	fmt.Println(x, y," empty")
	// }
	x.Right = y.Left
	if y.Left != rbt.NIL {
		y.Left.Parent = x
	}
	y.Parent = x.Parent // 交换父节点
	if x.Parent == rbt.NIL {
		//根节点
		rbt.Root = y
	} else if x == x.Parent.Left { // x 在根节点左边
		x.Parent.Left = y
	} else { // x 在根节点右边
		x.Parent.Right = y
	}
	y.Left = x
	x.Parent = y
}
func (rbt *RBtree) rightRotate(x *RBNode) {
	if x.Left != nil {
		return // 左孩子不能为 0
	}
	y := x.Left
	x.Left = y.Right
	if y.Right != rbt.NIL {
		y.Right.Parent = x
	}
	y.Parent = x.Parent
	if x.Parent == rbt.NIL {
		rbt.Root = y
	} else if x == x.Parent.Left { // x 小于根节点
		x.Parent.Left = y
	} else { // x 大于根节点
		x.Parent.Right = y
	}
	y.Right = x
	x.Parent = y
}

func (rbt *RBtree) Size() uint {
	return rbt.count
}
func (rbt *RBtree) Insert(item Item) *RBNode {
	if item == nil {
		return nil
	}
	return rbt.insert(&RBNode{rbt.NIL, rbt.NIL, rbt.NIL, RED, item})
}
func (rbt *RBtree) insert(z *RBNode) *RBNode {
	x := rbt.Root
	y := rbt.NIL
	for x != rbt.NIL {
		y = x
		if less(z.Item, x.Item) {
			x = x.Left
		} else if less(x.Item, z.Item) {
			x = x.Right
		} else {
			return x // 数据已经存在
		}
	}
	z.Parent = y
	if y == rbt.NIL {
		rbt.Root = z
	} else if less(z.Item, y.Item) {
		y.Left = z
	} else {
		y.Right = z
	}
	rbt.count++
	rbt.insertFixup(z) // 调整平衡
	return z
}

// 插入之后，调整平衡
func (rbt *RBtree) insertFixup(z *RBNode) {
	for z.Parent.Color == RED { // 一直循环下去，直到根节点
		if z.Parent == z.Parent.Parent.Left { // 父亲节点在爷爷左边

			y := z.Parent.Parent.Right
			if y.Color == RED { // 判断大伯节点红色，黑色
				z.Parent.Color = BLACK
				y.Color = BLACK

				z.Parent.Parent.Color = RED
				z = z.Parent.Parent // 循环前进
			} else {
				if z == z.Parent.Right {
					z = z.Parent
					rbt.leftRotate(z)

				}
				z.Parent.Color = BLACK
				z.Parent.Parent.Color = RED
				rbt.rightRotate(z.Parent.Parent)
			}
		} else { // 父亲节点在爷爷右边
			y := z.Parent.Parent.Left // 叔叔节点
			if y.Color == RED {
				z.Parent.Color = BLACK
				y.Color = BLACK
				z.Parent.Parent.Color = RED
				z = z.Parent.Parent // 循环前进
			} else {
				if z == z.Parent.Left {
					z = z.Parent
					rbt.rightRotate(z)
				}
				// else{
				z.Parent.Color = BLACK
				z.Parent.Parent.Color = RED
				rbt.leftRotate(z.Parent.Parent)
				// }
			}
		}
	}
	rbt.Root.Color = BLACK
}

func (rbt *RBtree) GetDepth() int {
	var getDeepth func(node *RBNode) int
	getDeepth = func(node *RBNode) int {
		if node == nil {
			return 0
		}
		if node.Left == nil && node.Right == nil {
			return 1
		}
		var leftdeep int = getDeepth(node.Left)
		var rightdepp int = getDeepth(node.Right)
		if leftdeep > rightdepp {
			return leftdeep + 1
		} else {
			return rightdepp + 1
		}
	}
	return getDeepth(rbt.Root)
}

func (rbt *RBtree) searchle(x *RBNode) *RBNode {
	p := rbt.Root
	n := p // 备份根节点
	for n != rbt.NIL {
		if less(n.Item, x.Item) {
			p = n
			n = n.Right // 大于
		} else if less(x.Item, n.Item) {
			p = n
			n = n.Left //小于
		} else {
			return n
		}
	}
	if less(p.Item, x.Item) {
		return p
	}
	p = rbt.deuccessor(p) // 近似处理
	return p
}

func (rbt *RBtree) successor(x *RBNode) *RBNode {
	if x == rbt.NIL {
		return rbt.NIL
	}
	if x.Right != rbt.NIL {
		return rbt.min(x.Right) // 取得右边最小
	}
	y := x.Parent
	for y != rbt.NIL && x == y.Right {
		x = y
		y = y.Parent
	}
	return y
}
func (rbt *RBtree) deuccessor(x *RBNode) *RBNode {
	if x == rbt.NIL {
		return rbt.NIL
	}
	if x.Left != rbt.NIL {
		return rbt.max(x.Left) // 取得左边最大
	}
	y := x.Parent
	for y != rbt.NIL && x == y.Left {
		x = y
		y = y.Parent
	}
	return y
}
func (rbt *RBtree) Delete(item Item) Item {
	if item == nil {
		return nil
	}
	return rbt.delete(&RBNode{rbt.NIL, rbt.NIL, rbt.NIL, RED, item})
}
func (rbt *RBtree) delete(key *RBNode) *RBNode {
	z := rbt.search(key)
	if z == rbt.NIL {
		return rbt.NIL // 无需删除
	}
	var x *RBNode
	var y *RBNode
	ret := &RBNode{rbt.NIL, rbt.NIL, rbt.NIL, z.Color, z.Item}
	if z.Left == rbt.NIL || z.Right == rbt.NIL {
		y = z // 直接替换删除
	} else {
		y = rbt.successor(z) // 找到最接近的
	}
	if y.Left != rbt.NIL {
		x = y.Left
	} else {
		x = y.Right
	}
	x.Parent = y.Parent

	if y.Parent == rbt.NIL {
		rbt.Root = x

	} else if y == y.Parent.Left {
		y.Parent.Left = x
	} else {
		y.Parent.Right = x
	}
	if y != z {
		z.Item = y.Item
	}
	if y.Color == BLACK {
		rbt.deleteFixup(x)
	}
	rbt.count--
	return ret
}
func (rbt *RBtree) deleteFixup(x *RBNode) {
	for x != rbt.Root && x.Color == BLACK {
		if x == x.Parent.Left { // x 在左边
			w := x.Parent.Right // 哥哥节点
			if w.Color == RED { // 左边旋转
				w.Color = BLACK
				x.Parent.Color = RED
				rbt.leftRotate(x.Parent)
				w = x.Parent.Right
			}
			if w.Left.Color == BLACK && w.Right.Color == BLACK {
				w.Color = RED
				x = x.Parent
			} else {
				if w.Right.Color == BLACK {
					w.Left.Color = BLACK
					w.Color = RED
					rbt.rightRotate(w) // 右旋转
					w = x.Parent.Right
				}
				w.Color = x.Parent.Color
				x.Parent.Color = BLACK
				w.Right.Color = BLACK
				rbt.leftRotate(x.Parent)
				x = rbt.Root
			}
		} else { // x 在右边
			w := x.Parent.Left
			if w.Color == RED {
				w.Color = BLACK
				x.Parent.Color = RED
				rbt.rightRotate(x.Parent)
				w = x.Parent.Left
			} else if w.Left.Color == BLACK && w.Right.Color == BLACK {
				w.Color = RED
				x = x.Parent
			} else {
				if w.Left.Color == BLACK {
					w.Right.Color = BLACK
					w.Color = RED
					rbt.leftRotate(w) // 右旋转
					w = x.Parent.Left
				}
				w.Color = x.Parent.Color
				x.Parent.Color = BLACK
				w.Left.Color = BLACK
				rbt.rightRotate(x.Parent)
				x = rbt.Root
			}
		}
	}
	x.Color = BLACK // 循环到最后就是根节点
}
