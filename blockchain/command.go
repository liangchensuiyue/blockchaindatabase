package blockchain

import (
	"fmt"
)

func (cli *CLI) PrintBlock() {
	cli.Bc.Print()
}
func (cli *CLI) GetBalance(address string) {
	if !IsValidAddress(address) {
		fmt.Printf("无效的地址: %s\n", address)
		return
	}
	pubKeyHash := GetPubKeyFromAddress(address)
	//[67 123 139 48 116 203 53 94 57 41 140 180 120 190 102 12 234 201 133 212]

	utxios := cli.Bc.FindUTXOs(pubKeyHash)
	total := 0.0
	for _, utxo := range utxios {
		total += utxo.Value
	}
	fmt.Printf("%s 的余额为 %f\n", address, total)
}

// user  用户地址(由公钥导过来)
// strict 是否严格处理，数据不会进入草稿，直接验证打包，也就是一个区块只有一笔交易
// sharemode 是否和某个用户共享数据
// shareuser 一起共享数据的用户

func (cli *CLI) PutData(key string, value []byte, user string, strict bool, sharemode string, shareuser []string) {
	fmt.Println(user, "PutData", key, value)
	NewTransaction(key, value, user, sharemode, shareuser)
}
func (cli *CLI) Send(from, to string, amount float64, miner string, data string) {
	fmt.Printf("from:%s\n", from)
	fmt.Printf("to:%s\n", to)
	fmt.Printf("amount:%f\n", amount)
	fmt.Printf("miner:%s\n", miner)
	fmt.Printf("data:%s\n", data)

	// 挖矿交易
	coinbase := NewCoinbaseTX(miner, data)

	// 创建普通交易
	tx := NewTransaction(from, to, amount, cli.Bc)
	if tx == nil {
		fmt.Println("无效的交易")
		return
	}
	cli.Bc.AddBlock([]*Transaction{coinbase, tx})
	fmt.Println("转账成功")
}
func (cli *CLI) NewWallet() {
	ws := NewWallets()
	ws.CreateWallet()
	for address := range ws.WalletsMap {
		fmt.Printf("地址: %s\n", address)
	}
	// fmt.Println("创建成功!", wallet)
}
func (cli *CLI) listAddress() {
	ws := NewWallets()
	addresses := ws.GetAllAddresses()
	for i, address := range addresses {
		fmt.Printf("钱包%v: %s\n", i, address)
	}
}
