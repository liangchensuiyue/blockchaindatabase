package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	BC "go_code/基于区块链的非关系型数据库/blockchain"
	uc "go_code/基于区块链的非关系型数据库/client"
	"go_code/基于区块链的非关系型数据库/database"
	db "go_code/基于区块链的非关系型数据库/database"
	quorum "go_code/基于区块链的非关系型数据库/quorum"
	Test "go_code/基于区块链的非关系型数据库/test"
	"go_code/基于区块链的非关系型数据库/util"
	view "go_code/基于区块链的非关系型数据库/view"
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
		// if newblock.IsGenesisBlock() {
		// 	localBlockChain.SignBlock(rw.Private, true, newblock)

		// } else {
		// 	localBlockChain.SignBlock(rw.Private, false, newblock)
		// }
		BC.BlockQueue.Insert(BC.QueueObject{
			TargetBlock: newblock,
			Handle: func(total, fail int) {
				flag := localBlockChain.VerifyBlock(rw.PubKey, newblock)
				if flag {
					e := localBlockChain.AddBlock(newblock)
					if e != nil {
						fmt.Println(e)
						return
					}
					for _, tx := range newblock.TxInfos {
						// BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
						// for _, addr := range tx.ShareAddress {
						// 	BC.LocalWallets.TailBlockHashMap[addr] = newblock.Hash
						// }
						if tx.Share {
							schn := BC.LocalWallets.ShareChanMap[tx.ShareChan]
							schn.TailBlockHash = newblock.Hash
						} else {
							BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
						}

					}
					BC.LocalWallets.SaveToFile()
					fmt.Println("同步区块", newblock.BlockId, "校验成功")
					return
				}
				fmt.Println("同步区块", newblock.BlockId, "校验失败")
			},
		})
	})
}
func addblocks(blocks []*BC.Block) {
	rw := BC.LocalWallets.GetBlockChainRootWallet()

	for _, newblock := range blocks {
		flag := localBlockChain.VerifyBlock(rw.PubKey, newblock)
		if flag {
			localBlockChain.AddBlock(newblock)
			if newblock.IsGenesisBlock() {
				BC.LocalWallets.TailBlockHashMap[rw.NewAddress()] = newblock.Hash
			}
			for _, tx := range newblock.TxInfos {
				// fmt.Println(tx.Share, tx.ShareAddress, base64.RawStdEncoding.EncodeToString(newblock.Hash))
				if tx.Share {
					schn := BC.LocalWallets.ShareChanMap[tx.ShareChan]
					schn.TailBlockHash = newblock.Hash

				} else {
					BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
				}
			}
			BC.LocalWallets.SaveToFile()
		} else {
			fmt.Println("同步区块", newblock.BlockId, "校验失败")
			return
		}
	}
}
func runLocalTestCli() {
	reader := bufio.NewReader(os.Stdin)
	// fmt.Println("start cli:")
	for {
		login_username := ""
		login_useraddress := ""
		pass := ""
		for {
			fmt.Printf(">>> ")
			flag := false

			_clistr, _, _ := reader.ReadLine()
			clistr := string(_clistr)
			_cmds := strings.Split(clistr, " ")
			cmds := []string{}
			for _, v := range _cmds {
				if v != "" {
					cmds = append(cmds, v)
				}
			}
			if len(cmds) == 0 {
				continue
			}
			switch cmds[0] {
			case "help":
				fmt.Println("newuser [username] [passworld] -- 创建用户")
				fmt.Println("latestblock --打印最新的块信息")
				fmt.Println("isaccountant --是否有记账权力")
				fmt.Println("login [username] [passworld] --登录")
				fmt.Println("print_quorum --打印集群信息")
			case "newuser":
				if len(cmds) < 3 {
					fmt.Println("格式错误")
					break
				}
				err := db.CreateUser(cmds[1], cmds[2])
				if err != nil {
					fmt.Println("创建失败:", err.Error())
				} else {
					fmt.Println("创建成功")
				}
			case "latestblock":
				b, _ := localBlockChain.GetTailBlock()
				fmt.Println("BlockId:", b.BlockId)
				fmt.Println("PreBLockHash:", base64.RawStdEncoding.EncodeToString(b.PreBlockHash))
				fmt.Println("Hash:", base64.RawStdEncoding.EncodeToString(b.Hash))
				fmt.Println("Timestamp:", b.Timestamp)
				fmt.Println("TxNums:", len(b.TxInfos))
			case "isaccountant":
				fmt.Println(quorum.LocalNodeIsAccount())
			case "login":
				if len(cmds) < 3 {
					fmt.Println("格式错误")
					break
				}
				e := database.VeriftUser(cmds[1], cmds[2])
				if e != nil {
					fmt.Println("登录失败:", e.Error())
				} else {
					login_useraddress, e = db.GetAddressFromUsername(cmds[1])
					if e != nil {
						fmt.Println("登录失败:", e.Error())
						break
					}
					pass = cmds[2]
					flag = true
					login_username = cmds[1]
				}
			case "print_quorum":
				for _, node := range localNode.Quorum {
					if node.LocalIp == localNode.LocalIp {
						fmt.Println(node.LocalIp, "(本机)")
					}
				}
			default:
				fmt.Println("格式错误")

			}
			if flag {
				break
			}

		}
		for {
			fmt.Printf("(%s)>>> ", login_username)
			_clistr, _, _ := reader.ReadLine()
			clistr := string(_clistr)
			_cmds := strings.Split(clistr, " ")
			cmds := []string{}
			flag := false
			for _, v := range _cmds {
				if v != "" {
					cmds = append(cmds, v)
				}
			}
			if len(cmds) == 0 {
				continue
			}
			switch cmds[0] {
			case "help":
				fmt.Println("newchan  -- 创建分享管道")
				fmt.Println("listchan --列出用户相关的分享管道")
				fmt.Println("put -- 录入数据")
				fmt.Println("get -- 获取数据")
				fmt.Println("del -- 删除数据")
				fmt.Println("isaccountant --是否有记账权力")
				fmt.Println("print_quorum --打印集群信息")
				fmt.Println("detail -- 查看区块信息")
				fmt.Println("print_global_wallet -- 查看全部用户地址")
				fmt.Println("print_local_wallet -- 查看本地用户地址")
				fmt.Println("exit -- 退出当前登录")
			case "newchan":
				if len(cmds) < 2 {
					fmt.Println("格式错误 newchan [channanme] ")
					break
				}
				newchan := &BC.ShareChan{
					Channame: cmds[1],
				}
				newchan.YieldKey()
				newchan.Creator = login_username
				newchan.CreatorAddress = login_useraddress
				ok := BC.UserIsChanCreator(newchan.Channame, login_useraddress)
				if ok {
					fmt.Println("改chan已存在")
					break
				}
				newchan.JoinKey = util.AesEncrypt([]byte(base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%d%s", time.Now().UnixNano(), newchan.Channame)))), newchan.Key)
				err := db.NewChan(newchan, login_username, login_useraddress)
				if err != nil {
					fmt.Println(err)
				}
			case "joinchan":
				if len(cmds) < 3 {
					fmt.Println("格式错误 look_join_key [channanme] [key] ")
					break
				}
				arr := strings.Split(cmds[2], ".")
				if len(arr) < 3 {
					fmt.Println("格式错误 key: s.s.s ")
					break
				}
				creator := arr[0]

				craddress, err := db.GetAddressFromUsername(creator)
				if err != nil {
					fmt.Println(err)
				}
				if !db.IsExsistChan(cmds[1], craddress) {
					fmt.Printf("不存在的chan:%s.%s\n", creator, cmds[1])
				}
				if BC.UserIsInChan(login_useraddress, creator, cmds[1]) {
					break
				}
				err = db.JoinChan(cmds[1], login_username, login_useraddress, creator, arr[1], arr[2])
				if err != nil {
					fmt.Println(err)
				}
			case "listchan":
				for _, v := range BC.LocalWallets.ShareChanMap {
					// fmt.Println(v.Channame, v.Creator, v.CreatorAddress, login_useraddress)
					if BC.UserIsInChan(login_useraddress, v.Creator, v.Channame) {
						fmt.Printf("%s(%s): ", v.Channame, v.Creator)
						for _, u := range db.GetChanUsers(v.Channame, v.CreatorAddress) {
							fmt.Printf("%s ", u)
						}
						fmt.Println("")
					}
				}
			case "look_join_key":
				if len(cmds) < 3 {
					fmt.Println("格式错误 look_join_key [channanme] [username] ")
					break
				}
				_, err := db.GetAddressFromUsername(cmds[2])
				if err != nil {
					fmt.Println(err)
					break
				}
				v, ok := BC.LocalWallets.ShareChanMap[cmds[2]+"."+cmds[1]]
				if !ok {
					fmt.Printf("%s 没有chan: %s\n", cmds[2], cmds[1])
				}
				if v.Creator != login_username {
					fmt.Println("没有权限查看")
				}

				fmt.Printf("%s.%s.%s", v.Creator, base64.RawStdEncoding.EncodeToString(util.AesDecrypt(v.JoinKey, v.Key)), base64.RawStdEncoding.EncodeToString(v.Key))
			case "put":
				// put age 15 int
				if len(cmds) < 6 {
					fmt.Println("格式错误  put [key] [value] [datatype] [strict] [sharemode] [sharechan]")
					break
				}
				var err error
				if !util.GetBoolFromStr(cmds[5]) {
					key := util.Yield16ByteKey([]byte(pass))
					v := util.AesEncrypt([]byte(cmds[2]), key)
					err = db.Put(cmds[1], v, BC.STRING, login_useraddress, util.GetBoolFromStr(cmds[5]), "", util.GetBoolFromStr(cmds[4]))

				} else {
					err = db.Put(cmds[1], []byte(cmds[2]), BC.STRING, login_useraddress, util.GetBoolFromStr(cmds[5]), cmds[6], util.GetBoolFromStr(cmds[4]))

				}
				if err != nil {
					fmt.Println("put", err)

				}
			case "del":
				if len(cmds) < 4 {
					fmt.Println("格式错误  del [key] [strict] [sharemode] [sharechan]")
					break
				}
				// del age
				db.Del(cmds[1], login_useraddress, util.GetBoolFromStr(cmds[3]), cmds[4], util.GetBoolFromStr(cmds[2]))
			case "get":
				if len(cmds) < 3 {
					fmt.Println("格式错误  get [key] [sharemode] [sharechan]")
					break
				}
				// get age
				var block *BC.Block
				var index int
				pre := time.Now().UnixNano()
				if !util.GetBoolFromStr(cmds[2]) {
					block, index = db.Get(cmds[1], login_username, login_useraddress, util.GetBoolFromStr(cmds[2]), "")

				} else {
					if len(cmds) < 4 {
						fmt.Println("格式错误  get [key] [sharemode] [sharechan]")
						break
					}
					block, index = db.Get(cmds[1], login_username, login_useraddress, util.GetBoolFromStr(cmds[2]), cmds[3])

				}
				if block != nil {
					// fmt.Println("blockid:", block.BlockId)
					// fmt.Println("block_hash", block.Hash)
					// fmt.Println("pre_block_hash:", base64.RawStdEncoding.EncodeToString(block.PreBlockHash))
					// fmt.Println("block_hash:", base64.RawStdEncoding.EncodeToString(block.Hash))
					if !block.TxInfos[index].Share {
						key := util.Yield16ByteKey([]byte(pass))
						v := util.AesDecrypt(block.TxInfos[index].Value, key)
						fmt.Println("key-value:", block.TxInfos[index].Key, string(v))
					} else {
						v := util.AesDecrypt(block.TxInfos[index].Value, BC.LocalWallets.ShareChanMap[block.TxInfos[index].ShareChan].Key)
						fmt.Println("key-value:", block.TxInfos[index].Key, string(v))
					}

					cur := time.Now().UnixNano()
					fmt.Println("耗时:", (cur-pre)/1000000, "(ms)")
				} else {
					fmt.Println("未查询到")
				}
			case "print_quorum":
				for _, node := range localNode.Quorum {
					if node.LocalIp == localNode.LocalIp {
						fmt.Println(node.LocalIp, "(本机)")
					}
				}
			case "detail":
				if len(cmds) < 2 {
					fmt.Println("格式错误  detail BlockId")
					break
				}
				localBlockChain.Traverse(func(block *BC.Block, err error) bool {
					if fmt.Sprintf("%d", block.BlockId) == cmds[1] {
						for i, tx := range block.TxInfos {
							fmt.Println("交易索引:", i)
							fmt.Println("user_address:", BC.GenerateAddressFromPubkey(tx.PublicKey))
							fmt.Println("key-value:", tx.Key, string(tx.Value))
							fmt.Println("sharemode:", tx.Share)
							fmt.Println("datatype:", tx.DataType)
							fmt.Println("sharechan:", tx.ShareChan)
						}
					}
					return true
				})
			case "isaccountant":
				fmt.Println(quorum.LocalNodeIsAccount())
			case "print":
				localBlockChain.Traverse(func(block *BC.Block, err error) bool {
					if block != nil {
						fmt.Println("---------------------------")
						fmt.Println("blockid:", block.BlockId)
						// fmt.Println("block_hash", block.Hash)
						fmt.Println("pre_block_hash:", base64.RawStdEncoding.EncodeToString(block.PreBlockHash))
						fmt.Println("block_hash:", base64.RawStdEncoding.EncodeToString(block.Hash))

						for i, tx := range block.TxInfos {
							fmt.Println("交易索引:", i)
							fmt.Println("user_address:", BC.GenerateAddressFromPubkey(tx.PublicKey))
							// fmt.Println("key-value:", tx.Key, string(tx.Value))
							fmt.Println("sharemode:", tx.Share)
							fmt.Println("datatype:", tx.DataType)
							fmt.Println("sharechan:", tx.ShareChan)
						}

					}
					return true

				})
				fmt.Printf("---------------------------\n\n")

			case "latestblock":
				b, _ := localBlockChain.GetTailBlock()
				fmt.Println("BlockId:", b.BlockId)
				fmt.Println("PreBLockHash:", base64.RawStdEncoding.EncodeToString(b.PreBlockHash))
				fmt.Println("Hash:", base64.RawStdEncoding.EncodeToString(b.Hash))
				fmt.Println("Timestamp:", b.Timestamp)
				fmt.Println("TxNums:", len(b.TxInfos))
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
			case "exit":
				flag = true
				break
			default:
				fmt.Println("格式错误")
			}
			if flag {
				break
			}
		}
	}
}
func main() {
	var genesis_file_name string
	var err error
	flag.StringVar(&genesis_file_name, "f", "./genesis", "genesis文件")
	localNode, err = quorum.LoadGenesisFile(genesis_file_name)
	fmt.Println("监听地址:", localNode.LocalIp, localNode.LocalPort)
	fmt.Println("localNode.Quorum:")
	for _, v := range localNode.Quorum {
		fmt.Println(v.LocalIp)
	}
	if err != nil {
		panic(err)
	}

	localBlockChain = BC.NewBlockChain(
		localNode.BCInfo.BlockTailHashKey,

		localNode.BCInfo.BlockChainDB,
		func(name string) {
			quorum.GetShareChan(name)
		},
	)

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
		flag := localBlockChain.VerifyBlock(rw.PubKey, genesis_block)
		if !flag {
			fmt.Println("创世块校验失败")
		}
		err = localBlockChain.AddBlock(genesis_block)
		if err != nil {
			panic(err)
		}
		BC.LocalWallets.TailBlockHashMap[rw.NewAddress()] = genesis_block.Hash
		BC.LocalWallets.SaveToFile()
	}

	_nodes := []string{}
	for _, v := range localNode.Quorum {
		_nodes = append(_nodes, v.LocalIp)
	}
	go view.Run(localBlockChain, _nodes)
	quorum.StartGrpcWork()
	StartDraftWork()

	db.Run(localBlockChain, localNode)
	go uc.Run()
	// fmt.Println("hello world")
	Test.Test2()
	// Test.Test1()
	// Test.Test3()
	runLocalTestCli()
}
