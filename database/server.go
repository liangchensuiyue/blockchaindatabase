package database

import (
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	quorum "go_code/基于区块链的非关系型数据库/quorum"
	"go_code/基于区块链的非关系型数据库/test"
)

var localBlockChain *BC.BlockChain
var localNode *quorum.BlockChainNode

func Run(lbc *BC.BlockChain, node *quorum.BlockChainNode) {
	localBlockChain = lbc
	localNode = node
	test.SystemInfo.PrintTxInfo = func() {
		fmt.Println(NUM, "笔交易耗时", Total/1000000, "(ms)")
		fmt.Println("交易池中交易数量", BC.BlockQueue.Len())
	}
	test.SystemInfo.CleanTxInfo = func() {
		NUM = 0
		Total = 0
	}
}
