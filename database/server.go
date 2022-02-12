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
}
