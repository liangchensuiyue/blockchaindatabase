package blockchain

type Dict struct {
	Dictht []*HT
	curidx int
	rehash bool
}

func (dt *Dict) Isbusy() bool {
	return dt.rehash
}
func (dt *Dict) DumpToTable(size int64) {
	dt.rehash = true
	newidx := (dt.curidx + 1) % 2
	dt.Dictht[newidx] = &HT{
		Size: size,
	}
	dt.Dictht[newidx].Init()
	dt.Dictht[dt.curidx].Traverse(func(entry *DictEntry) {
		dt.Dictht[newidx].Put(entry.Key, entry.Value)
		entry.Value = nil
	})
	dt.Dictht[dt.curidx] = &HT{}
	dt.curidx = newidx
}
func (dt *Dict) StartClean(s int) {
	dt.Dictht[dt.curidx].Shrink()
}
func (dt *Dict) Init() {
	dt.curidx = 0
	dt.Dictht = []*HT{&HT{}, &HT{}}
	dt.Dictht[dt.curidx].Size = 10000
	dt.Dictht[dt.curidx].Init()
	dt.rehash = false
}
func (dt *Dict) Put(k string, v *Transaction) {
	if dt.Dictht[dt.curidx].Used > (dt.Dictht[dt.curidx].Size*400)/10 {
		dt.Dictht[dt.curidx].Shrink()
	}
	dt.Dictht[dt.curidx].Put(k, v)
}
func (dt *Dict) Get(k string) []*Transaction {
	return dt.Dictht[dt.curidx].Get(k)
}

var Global_DICT *Dict

func init() {
	Global_DICT = &Dict{}
	Global_DICT.Init()
}
