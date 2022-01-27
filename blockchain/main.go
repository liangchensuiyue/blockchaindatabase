package blockchain

// "fmt"

// "math/big"

func main1() {

	bc := NewBlockChain("179pGpZNG5GJGRUUrEBmJegy5VrzjwWcMC fff")
	cli := CLI{bc}
	cli.Run()
	// // fmt.Println(a)
	// bc.AddBlock("良辰")
	// bc.AddBlock("岁月")
	// bc.AddBlock("哈哈哈哈")
	// fmt.Println()

	// it := bc.NewIterator()
	// block := &block.Block{}
	// for ;block!=nil;block = it.Next(){
	// 	fmt.Println("============================================")
	// 	fmt.Printf("区块版本: %d\n", block.Version)
	// 	fmt.Printf("前区块hash: %x\n", block.PrevHash)
	// 	fmt.Printf("MerkelRoot: %x\n", block.MerkelRoot)
	// 	fmt.Printf("时间戳: %d\n", block.TimeStamp)
	// 	fmt.Printf("难度值: %x\n", block.Difficulty)
	// 	fmt.Printf("Nonce: %d\n", block.Nonce)
	// 	fmt.Printf("Hash: %x\n", block.Hash)
	// 	fmt.Printf("数据: %s\n", block.Data)
	// 	fmt.Println("============================================")
	// }
	// bc.Print()
	// for bc.HasNext(){
	// 	block,err := bc.Next()
	// 	if err != nil{
	// 		panic(err)
	// 	}
	// 	fmt.Println("============================================")
	// 	fmt.Printf("区块版本: %d\n", block.Version)
	// 	fmt.Printf("前区块hash: %x\n", block.PrevHash)
	// 	fmt.Printf("MerkelRoot: %x\n", block.MerkelRoot)
	// 	fmt.Printf("时间戳: %d\n", block.TimeStamp)
	// 	fmt.Printf("难度值: %x\n", block.Difficulty)
	// 	fmt.Printf("Nonce: %d\n", block.Nonce)
	// 	fmt.Printf("Hash: %x\n", block.Hash)
	// 	fmt.Printf("数据: %s\n", block.Data)
	// 	fmt.Println("============================================")

	// }

}
