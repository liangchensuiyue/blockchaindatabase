package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"

	BC "go_code/基于区块链的非关系型数据库/blockchain"
	db "go_code/基于区块链的非关系型数据库/database"
	quorum "go_code/基于区块链的非关系型数据库/quorum"
)

var localBlockChain *BC.BlockChain
var localNode *quorum.BlockChainNode

func StartDraftWork() {
	draft := BC.GetLocalDraftFromDisk()
	rw := BC.LocalWallets.GetBlockChainRootWallet()
	go draft.Work(func(newblock *BC.Block, e error) {
		if len(newblock.TxInfos) == 0 && !newblock.IsGenesisBlock() {
			// 如果不是创世块，并且交易数目为0 ，则不能打包
			return
		}
		if newblock.IsGenesisBlock() {
			localBlockChain.SignBlock(rw.Private, true, newblock)

		} else {
			localBlockChain.SignBlock(rw.Private, false, newblock)
		}
		localNode.DistribuBlock(newblock, func(total, fail int) {
			flag := localBlockChain.VerifyBlock(rw.PubKey, newblock)
			if flag {
				e := localBlockChain.AddBlock(newblock)
				if e != nil {
					fmt.Println(e)
					return
				}
				for _, tx := range newblock.TxInfos {
					BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
					for _, addr := range tx.ShareAddress {
						BC.LocalWallets.TailBlockHashMap[addr] = newblock.Hash
					}

				}
				BC.LocalWallets.SaveToFile()
				fmt.Println("校验成功")
				return
			}
			fmt.Println("区块校验失败")
		})
	})
}
func addblocks(blocks []*BC.Block) {
	rw := BC.LocalWallets.GetBlockChainRootWallet()

	for _, newblock := range blocks {
		flag := localBlockChain.VerifyBlock(rw.PubKey, newblock)
		if flag {
			localBlockChain.AddBlock(newblock)
			for _, tx := range newblock.TxInfos {
				BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
				for _, addr := range tx.ShareAddress {
					BC.LocalWallets.TailBlockHashMap[addr] = newblock.Hash
				}

			}
		} else {
			return
		}
	}
}
func runLocalTestCli() {
	reader := bufio.NewReader(os.Stdin)
	for {

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
			// put age 15 int
			err := db.Put(cmds[1], []byte(cmds[2]), cmds[3], cmds[4], false, []string{}, true)
			fmt.Println("put", err)
		case "del":
			// del age
			db.Del(cmds[1], cmds[2], false, []string{}, true)
		case "get":
			// get age
			block, index := db.Get(cmds[1], cmds[2], false, []string{})
			fmt.Println("get:")
			if block != nil {
				fmt.Println("blockid:", block.BlockId)
				// fmt.Println("block_hash", block.Hash)
				fmt.Println("pre_block_hash:", base64.RawStdEncoding.EncodeToString(block.PreBlockHash))
				fmt.Println("block_hash:", base64.RawStdEncoding.EncodeToString(block.Hash))
				fmt.Println("key-value:", block.TxInfos[index].Key, string(block.TxInfos[index].Value))
			} else {
				fmt.Println("未查询到")
			}
		case "newuser":
			db.CreateUser(cmds[1], cmds[2])
		case "print_quorum":
			for _, node := range localNode.Quorum {
				if node.LocalIp == localNode.LocalIp {
					fmt.Println(node.LocalIp, "(本机)")
				}
			}
		case "print":
			localBlockChain.Traverse(func(block *BC.Block, err error) {
				if block != nil {
					fmt.Println("---------------------------")
					fmt.Println("blockid:", block.BlockId)
					// fmt.Println("block_hash", block.Hash)
					fmt.Println("pre_block_hash:", base64.RawStdEncoding.EncodeToString(block.PreBlockHash))
					fmt.Println("block_hash:", base64.RawStdEncoding.EncodeToString(block.Hash))

					for i, tx := range block.TxInfos {
						w, e := BC.LocalWallets.GetUserWallet(BC.GenerateAddressFromPubkey(tx.PublicKey))
						fmt.Println("交易索引:", i)
						fmt.Println("user:", w.Username)
						fmt.Println("key-value:", tx.Key, string(tx.Value))
						fmt.Println("sharemode:", tx.Share)
						fmt.Println("delmark:", tx.DelMark)
						fmt.Println("shareuser:")
						for _, uaddr := range tx.ShareAddress {
							w, e = BC.LocalWallets.GetUserWallet(uaddr)
							if e == nil {
								fmt.Println(w.NewAddress())
							}
						}
					}

				}

			})
			fmt.Printf("---------------------------\n\n")

		case "print_tail_block":
			block, _ := localBlockChain.GetTailBlock()
			fmt.Println("tail_blockId", block.BlockId)
			fmt.Println("tail_hash", block.Hash)
			for i, tx := range block.TxInfos {
				fmt.Println("交易", i)
				fmt.Println(tx.Key)
				fmt.Println(tx.Value)
			}
		case "print_global_wallet":
			rw := BC.LocalWallets.GetBlockChainRootWallet()
			user_address := rw.NewAddress()
			fmt.Println(rw.Username, user_address)

			// 判断用户是否创建
			_hash, _ := BC.LocalWallets.GetUserTailBlockHash(user_address)

			b, e := localBlockChain.GetBlockByHash(_hash)
			if e != nil {
				break
			}
			for {
				if b.IsGenesisBlock() {
					break
				}
				for _, tx := range b.TxInfos {
					_hash = tx.PreBlockHash
					fmt.Println(tx.Key, strings.Split(string(tx.Value), " ")[1])

				}
				b, _ = localBlockChain.GetBlockByHash(_hash)
			}
		case "print_local_wallet":
			for addr, w := range BC.LocalWallets.WalletsMap {
				fmt.Println(w.Username, addr)
			}
		default:
			fmt.Println(cmds)
		}
	}
}
func main() {
	var genesis_file_name string
	var err error
	flag.StringVar(&genesis_file_name, "f", "./genesis", "genesis文件")
	localNode, err = quorum.LoadGenesisFile(genesis_file_name)
	if err != nil {
		panic(err)
	}
	localBlockChain = BC.NewBlockChain(
		localNode.BCInfo.BlockTailHashKey,

		localNode.BCInfo.BlockChainDB)
	BC.LoadLocalWallets()
	_, err = localBlockChain.GetAddressFromUsername("liangchen")
	if err != nil {
		wa := BC.NewWallet("liangchen", localNode.BCInfo.PassWorld)
		wa.Private = localNode.BCInfo.PriKey
		wa.PubKey = localNode.BCInfo.PubKey
		BC.LocalWallets.WalletsMap[wa.NewAddress()] = wa
		BC.LocalWallets.SaveToFile()
	}
	quorum.Broadcast(localBlockChain)

	newbllocks, e := quorum.BlockSynchronization()
	if e == nil {
		addblocks(newbllocks)
	}

	_tailblock, _ := localBlockChain.GetTailBlock()
	if _tailblock == nil {
		// 创建创世块
		genesis_block := BC.NewGenesisBlock()
		genesis_block.BlockId = 1

		rw := BC.LocalWallets.GetBlockChainRootWallet()
		localBlockChain.SignBlock(rw.Private, true, genesis_block)
		err = localBlockChain.AddBlock(genesis_block)
		if err != nil {
			panic(err)
		}
		BC.LocalWallets.TailBlockHashMap[rw.NewAddress()] = genesis_block.Hash
	}

	quorum.StartGrpcWork()
	StartDraftWork()

	db.Run(localBlockChain, localNode)
	fmt.Println("hello world")

	runLocalTestCli()
}
