package lru

import (
	"container/list"
)

type Cache struct {
	//缓存的数量
	MaxEntries int
	//保存缓存
	Cachelist *list.List
	//映射
	cache map[interface{}]*list.Element
	//函数指针
	Onout func(key interface{}, value interface{})
}

//映射结构
type Entry struct {
	key   interface{}
	value interface{}
}

//初始化
func NewCache(maxEntries int) *Cache {
	return &Cache{MaxEntries: maxEntries, Cachelist: list.New(), cache: make(map[interface{}]*list.Element)}
}

//获取缓存的长度
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.Cachelist.Len()
}

//删除一个元素
func (c *Cache) removeElement(e *list.Element) {
	c.Cachelist.Remove(e)
	kv := e.Value.(*Entry)
	delete(c.cache, kv.key) //map映射中删除
	if c.Onout != nil {
		c.Onout(kv.key, kv.value) //删除
	}
}

//删除最后一个访问
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	ele := c.Cachelist.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

//删除key
func (c *Cache) Remove(key interface{}) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

//创建数据插入cache
func (c *Cache) Add(key interface{}, value interface{}) {
	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.Cachelist = list.New() //不存在开辟内存
	}
	if ee, ok := c.cache[key]; ok {
		//存在，迁移到头部
		c.Cachelist.MoveToFront(ee)     //提取到头部
		ee.Value.(*Entry).value = value //更新数据
		return
	}
	ele := c.Cachelist.PushFront(&Entry{key, value})
	c.cache[key] = ele
	if c.MaxEntries != 0 && c.Cachelist.Len() > c.MaxEntries {
		c.RemoveOldest() //删除最后一个
	}

}

//提取数据cache +1
func (c *Cache) Get(key interface{}) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.Cachelist.MoveToFront(ele) //提取到头部
		return ele.Value.(*Entry).value, true
	}
	return nil, false
}

// func main() {
// 	t1 := &Entry{"yincheng", "yincheng.txt"}
// 	t2 := &Entry{"liu", "liu.txt"}
// 	t3 := &Entry{"xin", "xin.txt"}
// 	t4 := &Entry{"cheng", "cheng.txt"}

// 	myc := NewCache(3)
// 	myc.Add(t1.key, t1.value)
// 	myc.Add(t2.key, t2.value)
// 	myc.Add(t3.key, t3.value)
// 	myc.Add(t4.key, t4.value)

// 	fmt.Println(myc, myc.Cachelist)
// 	fmt.Println(myc.Get(t2.key))
// 	fmt.Println(myc, myc.Cachelist)

// }
