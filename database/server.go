package database

import (
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	quorum "go_code/基于区块链的非关系型数据库/quorum"
)

var localBlockChain *BC.BlockChain
var localNode *quorum.BlockChainNode

func Run(lbc *BC.BlockChain, node *quorum.BlockChainNode) {
	localBlockChain = lbc
	localNode = node
	// go func() {
	// 	for {
	// 		time.Sleep(time.Second * 2)
	// 		fmt.Println(NUM, "笔交易耗时", Total/1000000, "(ms)")
	// 		fmt.Println("队列数量", BC.BlockQueue.Len())
	// 	}
	// }()
}
