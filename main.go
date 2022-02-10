package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	BC "go_code/基于区块链的非关系型数据库/blockchain"
	quorum "go_code/基于区块链的非关系型数据库/quorum"
)

var localBlockChain *BC.BlockChain
var localNode *quorum.NodeInfo

func LoadGenesisFile(filename string) (*quorum.NodeInfo, error) {
	var info quorum.NodeInfo

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil, err
	}
	// 读取钱包
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// 解码
	gob.Register(elliptic.P256())

	decoder := gob.NewDecoder(bytes.NewReader(quorum.AesDecrypt(content, []byte("1234567812345678"))))
	err = decoder.Decode(&info)
	if err != nil {
		panic(err)
	}
	return &info, nil
}
func main() {
	var genesis_file_name string
	var bind_port int
	var err error
	flag.StringVar(&genesis_file_name, "f", "./genesis", "genesis文件")
	flag.IntVar(&bind_port, "port", 3300, "节点访问端口")
	localNode, err = LoadGenesisFile(genesis_file_name)
	if err != nil {
		panic(err)
	}
	BC.LoadLocalWallets()
	localBlockChain = BC.NewBlockChain(
		localNode.BCInfo.PassWorld,
		localNode.BCInfo.BlockTailHashKey,
		localNode.BCInfo.BlockChainDB)
	if localNode.BCInfo.TailBlockId == 0 {
		// 创建创世块
		genesis_block := BC.NewGenesisBlock()
		localBlockChain.SignBlock(localNode.BCInfo.PriKey, "", genesis_block)
	}

	fmt.Println("hello world")
}
