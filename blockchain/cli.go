package blockchain

import (
	"fmt"
	"os"
	"strconv"
)

type CLI struct {
	Bc *BlockChain
}

const Usage = `
	printChain			打印区块链
	getBalance	--address ADDRESS 获取地址的余额
	send FROM TO AMOUNT MINER DATA "tx info"
	newWallet		"创建一个新的钱包"
	listAddress		列举所有的钱包地址
`

func (cli *CLI) Run() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println(Usage)
		return
	}
	cmd := args[1]
	switch cmd {
	case "printChain":
		fmt.Println("区块链信息:")
		cli.Bc.Print()
		//打印区块
	case "getBalance":
		// fmt.Println("获取余额")
		if len(args) == 4 && args[2] == "--address" {
			// fmt.Println("address")
			cli.GetBalance(args[3])
		} else {
			cli.PrintBlock()
		}
	case "send":
		if len(args) != 7 {
			fmt.Println(Usage)
			return
		}
		fmt.Println("转账开始......")
		from := args[2]
		to := args[3]
		amount, err := strconv.ParseFloat(args[4], 64)
		if err != nil {
			fmt.Println("参数类型错误s")
			fmt.Println(Usage)
			return
		}
		miner := args[5]
		data := args[6]
		cli.Send(from, to, amount, miner, data)
	case "newWallet":
		fmt.Println("创建新的钱包")
		cli.NewWallet()
	case "listAddress":
		fmt.Println("地址列表")
		cli.listAddress()
	default:
		fmt.Println("无效的参数")
		fmt.Println(Usage)
	}
}
