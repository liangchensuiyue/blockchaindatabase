package main

import (
	"fmt"

	bc "go_code/基于区块链的非关系型数据库/blockchain"
)

func main() {
	bc.LoadLocalWallets()
	fmt.Println("hello world")
}
