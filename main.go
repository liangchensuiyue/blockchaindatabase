package main

import (
	"bufio"
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	BC "go_code/基于区块链的非关系型数据库/blockchain"
	db "go_code/基于区块链的非关系型数据库/database"
	quorum "go_code/基于区块链的非关系型数据库/quorum"
)

var localBlockChain *BC.BlockChain
var localNode *quorum.BlockChainNode

func LoadGenesisFile(filename string) (*quorum.BlockChainNode, error) {
	var info quorum.BlockChainNode

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
func StartDraftWork() {
	draft := BC.GetLocalDraftFromDisk()
	draft.Work(func(newblock *BC.Block, e error) {
		if len(newblock.TxInfos) == 0 && !newblock.IsGenesisBlock() {
			// 如果不是创世块，并且交易数目为0 ，则不能打包
			return
		}
		if newblock.IsGenesisBlock() {
			localBlockChain.SignBlock(localNode.BCInfo.PriKey, true, newblock)

		} else {
			localBlockChain.SignBlock(localNode.BCInfo.PriKey, false, newblock)
		}
		localNode.DistribuBlock(newblock)
	})
}
func addblocks(blocks []*BC.Block) {
	for _, newblock := range blocks {
		flag := localBlockChain.VerifyBlock(localNode.BCInfo.PubKey, newblock)
		if flag {
			localBlockChain.AddBlock(newblock)
			for _, tx := range newblock.TxInfos {
				for _, addr := range tx.ShareAddress {
					wa, e := BC.LocalWallets.GetUserWallet(addr)
					if e != nil {
						wa.TailBlockHash = newblock.Hash
					}
				}

			}
		} else {
			return
		}
	}
}
func runLocalTestCli() {
	reader := bufio.NewReader(os.Stdin)
	_clistr, _, _ := reader.ReadLine()
	clistr := string(_clistr)
	_cmds := strings.Split(clistr, " ")
	cmds := []string{}
	for _, v := range _cmds {
		if v != "" {
			cmds = append(cmds, v)
		}
	}
	switch cmds[0] {
	case "put":
		err := db.Put(cmds[1], []byte(cmds[2]), cmds[3], cmds[4], false, []string{}, true)
		fmt.Println("del", err)
	case "del":
		db.Del(cmds[1], cmds[2], false, []string{}, true)
	case "get":
		v := db.Get(cmds[1], cmds[2], false, []string{})
		fmt.Println("get", string(v), len(v))
	case "newuser":
		db.CreateUser(cmds[1], cmds[2])
	default:
		fmt.Println(cmds)
	}
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
	quorum.Broadcast()
	newbllocks, e := quorum.BlockSynchronization()
	if e != nil {
		addblocks(newbllocks)
	}
	localBlockChain = BC.NewBlockChain(
		localNode.BCInfo.PassWorld,
		localNode.BCInfo.BlockTailHashKey,
		localNode.BCInfo.BlockChainDB)
	if localNode.BCInfo.TailBlockId == 0 {
		// 创建创世块
		genesis_block := BC.NewGenesisBlock()
		genesis_block.BlockId = 1
		localBlockChain.SignBlock(localNode.BCInfo.PriKey, true, genesis_block)
		err = localBlockChain.AddBlock(genesis_block)
		if err != nil {
			panic(err)
		}
	}

	quorum.StartGrpcWork(localNode, localBlockChain)
	StartDraftWork()

	db.Run(localBlockChain, localNode)

	runLocalTestCli()
	fmt.Println("hello world")
}
