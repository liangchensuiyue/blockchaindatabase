package blockchain

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

type Draft struct {

	// 草稿中的交易
	TxInfos []*Transaction

	// 交易打包数目上限
	PackNum int

	// 交易打包时间上限 单位 秒
	Time int

	// 草稿运行状态
	WorkStatus bool

	DraftBlock *sync.Mutex
}

var draft_datat_file_name string = "draft"
var _pre_time time.Time

func (draft *Draft) loadFile() {
	_, err := os.Stat(walletFile)
	if os.IsNotExist(err) {
		draft.TxInfos = make([]*Transaction, 10)
		return
	}
	// 读取钱包
	content, err := ioutil.ReadFile(draft_datat_file_name)
	if err != nil {
		log.Panic(err)
	}

	// 解码
	gob.Register(elliptic.P256())

	decoder := gob.NewDecoder(bytes.NewReader(content))

	var d Draft
	err = decoder.Decode(&d)
	if err != nil {
		log.Panic(err)
	}
	// ws = &wsLocal
	draft.TxInfos = d.TxInfos
}
func (draft *Draft) saveToFile() {
	defer draft.DraftBlock.Unlock()
	/*
		如果 Encode/Decode 类型是interface或者struct中某些字段是interface{}的时候
		需要在gob中注册interface可能的所有实现或者可能类型
	*/
	var content bytes.Buffer

	// Curve 是一个接口类型
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(draft)
	if err != nil {
		log.Panic(err)
	}
	draft.DraftBlock.Lock()
	err = ioutil.WriteFile(draft_datat_file_name, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
func (draft *Draft) PackBlock() (*Block, error) {
	defer draft.DraftBlock.Unlock()
	draft.DraftBlock.Lock()
	var newblock *Block
	if len(draft.TxInfos) > draft.PackNum {
		newblock = NewBlock(draft.TxInfos[:draft.PackNum])
		draft.TxInfos = draft.TxInfos[draft.PackNum:]

	} else {
		newblock = NewBlock(draft.TxInfos)
		draft.TxInfos = []*Transaction{}
	}
	return newblock, nil
}
func (draft *Draft) GetTxInfosNum() int {
	return len(draft.TxInfos)
}
func (draft *Draft) Work(handler func(*Block, error)) {
	_pre_time = time.Now()
	for {
		if !draft.WorkStatus {
			continue
		}
		cur := time.Now()
		_flag := cur.After(_pre_time.Add(time.Second * time.Duration(draft.Time)))
		if _flag {
			// 到了所设置的草稿打包的时间
			b, e := draft.PackBlock()
			handler(b, e)
			_pre_time = cur
		}
	}
}
