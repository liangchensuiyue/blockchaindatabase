package rbtree

type Int int

func (x Int) Less(then Item) bool {
	return x < then.(Int)
}

type UInt32 uint32

func (x UInt32) Less(then Item) bool {
	return x < then.(UInt32)
}

type String string

func (x String) Less(then Item) bool {
	return x < then.(String)
}
