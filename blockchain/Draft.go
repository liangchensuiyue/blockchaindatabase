package blockchain

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
)

type Draft struct {

	// 草稿中的交易
	TxInfos []*Transaction

	// 交易打包数目上限
	PackNum int

	// 交易打包时间上限 单位 秒
	time int
}

var draft_datat_file_name string = "draft"

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
	defer _lock.Unlock()
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
	_lock.Lock()
	err = ioutil.WriteFile(draft_datat_file_name, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
func (draft *Draft) PackBlock(pre_block *Block) (*Block, error) {

	block := NewBlock(pre_block.BlockId+1, pre_block.Hash, draft.TxInfos)
	if len(draft.TxInfos) > draft.PackNum {

	}
}
