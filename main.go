package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	BC "go_code/基于区块链的非关系型数据库/blockchain"
	uc "go_code/基于区块链的非关系型数据库/client"
	"go_code/基于区块链的非关系型数据库/database"
	db "go_code/基于区块链的非关系型数据库/database"
	quorum "go_code/基于区块链的非关系型数据库/quorum"
	test "go_code/基于区块链的非关系型数据库/test"
	Type "go_code/基于区块链的非关系型数据库/type"
	"go_code/基于区块链的非关系型数据库/util"
	"go_code/基于区块链的非关系型数据库/view"
)

var LocalBlockChain *BC.BlockChain
var LocalNode *quorum.BlockChainNode

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
				flag := LocalBlockChain.VerifyBlock(rw.PubKey, newblock)
				if flag {
					e := LocalBlockChain.AddBlock(newblock)
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
							// BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
							BC.LocalWallets.SetTailBlockHash(BC.GenerateAddressFromPubkey(tx.PublicKey), newblock.Hash)
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
		flag := LocalBlockChain.VerifyBlock(rw.PubKey, newblock)
		if flag {
			LocalBlockChain.AddBlock(newblock)
			if newblock.IsGenesisBlock() {
				// BC.LocalWallets.TailBlockHashMap[rw.NewAddress()] = newblock.Hash
				BC.LocalWallets.SetTailBlockHash(rw.NewAddress(), newblock.Hash)

			}
			for _, tx := range newblock.TxInfos {
				// fmt.Println(tx.Share, tx.ShareAddress, base64.RawStdEncoding.EncodeToString(newblock.Hash))
				if tx.Share {
					schn := BC.LocalWallets.ShareChanMap[tx.ShareChan]
					schn.TailBlockHash = newblock.Hash

				} else {
					// BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
					BC.LocalWallets.SetTailBlockHash(BC.GenerateAddressFromPubkey(tx.PublicKey), newblock.Hash)
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
	// db.CreateUser("user1", "123")
	// db.CreateUser("user2", "123")
	// db.CreateUser("user3", "123")
	// db.CreateUser("user4", "123")
	// db.CreateUser("user5", "123")
	reader := bufio.NewReader(os.Stdin)
	// fmt.Println("start cli:")
	for {
		login_username := ""
		login_useraddress := ""
		pass := ""
		strictmode := true
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
				fmt.Println("systxrate --系统交易处理速率")
				fmt.Println("sysblockrate --系统区块同步速率")
				fmt.Println("resetsystxi --重置系统交易处理速率")
				fmt.Println("resetsysbki --重置系统区块同步速率")
			case "newuser":
				if len(cmds) < 3 {
					fmt.Println("格式错误")
					break
				}
				fmt.Println("正在创建用户...")
				err := db.CreateUser(cmds[1], cmds[2])
				if err != nil {
					fmt.Println("创建失败:", err.Error())
				} else {
					fmt.Println("创建成功")
				}
			case "latestblock":
				b, _ := LocalBlockChain.GetTailBlock()
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
				for _, node := range LocalNode.Quorum {
					if node.LocalIp == LocalNode.LocalIp {
						fmt.Println(node.LocalIp, "(本机)")
					} else {
						fmt.Println(node.LocalIp)
					}
				}
			case "systxrate":
				// fmt.Println("----------------------------------------")
				test.SystemInfo.PrintTxInfo()
				// fmt.Println("----------------------------------------")
			case "resetsystxi":
				test.SystemInfo.CleanTxInfo()
			case "sysblockrate":
				// fmt.Println("----------------------------------------")
				test.SystemInfo.PrintBlockInfo()
				// fmt.Println("----------------------------------------")
			case "resetsysbki":
				test.SystemInfo.CleanBlockInfo()
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
				fmt.Println("newchan  channanme -- 创建分享管道")
				fmt.Println("delchan  channanme -- 删除管道")
				fmt.Println("joinchan  chankey -- 加入管道")
				fmt.Println("exitchan  channame -- 退出管道")
				fmt.Println("listchan --列出用户加入的分享管道")
				fmt.Println("putstr value [channame] -- 录入字符串")
				fmt.Println("putint32 value [channame] -- 录入 int32")
				fmt.Println("putint64 value [channame] -- 录入 int64")
				fmt.Println("putstrarr key [v1,...,vn] [sharechan] -- 录入 字符串 数组")
				fmt.Println("puti32arr key [v1,...,vn] [sharechan] -- 录入 int32 数组")
				fmt.Println("puti64arr key [v1,...,vn] [sharechan] -- 录入 int64 数组")
				fmt.Println("putstrset key [v1,...,vn] [sharechan] -- 录入 字符串 集合")
				fmt.Println("puti32set key [v1,...,vn] [sharechan] -- 录入 int32 集合")
				fmt.Println("puti64set key [v1,...,vn] [sharechan] -- 录入 int64 集合")
				fmt.Println("get key [channame] -- 获取数据")
				fmt.Println("del key [channame] -- 删除数据")
				fmt.Println("isaccountant --是否有记账权力")
				fmt.Println("print_quorum --打印集群信息")
				fmt.Println("detail blockid-- 查看区块信息")
				fmt.Println("togglemode -- 切换strict")
				fmt.Println("print_global_wallet -- 查看全部用户地址")
				fmt.Println("print_local_wallet -- 查看本地用户地址")
				fmt.Println("systxrate --系统交易处理速率")
				fmt.Println("sysblockrate --系统区块同步速率")
				fmt.Println("resetsystxi --重置系统交易处理速率")
				fmt.Println("resetsysbki --重置系统区块同步速率")
				fmt.Println("exit -- 退出当前登录")
			case "togglemode":
				strictmode = !strictmode
				fmt.Println("strictmode:", strictmode)
			case "newchan":
				if len(cmds) < 2 {
					fmt.Println("格式错误 newchan channanme ")
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
				if len(cmds) < 2 {
					fmt.Println("格式错误 joinchan key ")
					break
				}
				_str, e := base64.RawStdEncoding.DecodeString(cmds[1])
				if e != nil {
					fmt.Println("错误的密钥")
					break
				}
				arr := strings.Split(string(_str), ".")
				if len(arr) < 4 {
					fmt.Println("格式错误 key: s.s.s ")
					break
				}
				creator := arr[0]

				craddress, err := db.GetAddressFromUsername(creator)
				if err != nil {
					fmt.Println(err)
					break
				}
				if !db.IsExsistChan(arr[0]+"."+arr[1], craddress) {
					fmt.Printf("不存在的chan:%s.%s\n", creator, arr[1])
					break
				}
				if BC.UserIsInChan(login_useraddress, creator, arr[1]) {
					break
				}
				err = db.JoinChan(arr[1], login_username, login_useraddress, creator, arr[2], arr[3])
				if err != nil {
					fmt.Println(err)
				}
			case "exitchan":
				if len(cmds) < 3 {
					fmt.Println("格式错误 exitchan creator channame")
					break
				}
				db.ExitChan(cmds[1], cmds[2], login_username, login_useraddress)
			case "listchan":
				_map := make(map[string]bool)
				arrage := []string{}
				LocalBlockChain.Traverse(func(block *BC.Block, err error) bool {
					if block != nil {
						for i := len(block.TxInfos) - 1; i >= 0; i-- {
							tx := block.TxInfos[i]
							if tx.DataType == Type.NEW_CHAN {
								_, ok := _map[tx.Key]
								if !ok {
									arrage = append(arrage, tx.Key)
									_map[tx.Key] = true
								}

							}
							if tx.DataType == Type.DEL_CHAN {
								_map[tx.Key] = false
							}
						}

					}
					return true

				})
				for _, v := range arrage {
					// fmt.Println(v.Channame, v.Creator, v.CreatorAddress, login_useraddress)
					arr := strings.Split(v, ".")
					if BC.UserIsInChan(login_useraddress, arr[0], arr[1]) {
						fmt.Printf("%s.%s: ", arr[0], arr[1])
						addr, _ := database.GetAddressFromUsername(arr[0])
						for _, u := range db.GetChanUsers(arr[0]+"."+arr[1], arr[0], addr) {
							fmt.Printf("%s ", u)
						}
						fmt.Println("")
					}
				}
			case "look_join_key":
				if len(cmds) < 3 {
					fmt.Println("格式错误 look_join_key username channanme ")
					break
				}
				_, err := db.GetAddressFromUsername(cmds[1])
				if err != nil {
					fmt.Println(err)
					break
				}
				v, ok := BC.LocalWallets.ShareChanMap[cmds[1]+"."+cmds[2]]
				if !ok {
					fmt.Printf("%s 没有chan: %s\n", cmds[2], cmds[1])
					break
				}
				if v.Creator != login_username {
					fmt.Println("没有权限查看")
				}

				str := fmt.Sprintf("%s.%s.%s.%s\n", v.Creator, v.Channame, base64.RawStdEncoding.EncodeToString(util.AesDecrypt(v.JoinKey, v.Key)), base64.RawStdEncoding.EncodeToString(v.Key))
				fmt.Println(base64.RawStdEncoding.EncodeToString([]byte(str)))
			case "putstr":
				// put age 15 int
				if len(cmds) < 3 {
					fmt.Println("格式错误  putstr key value [sharechan]")
					break
				}
				var err error
				if len(cmds) >= 4 {
					err = db.Put(cmds[1], []byte(cmds[2]), Type.STRING, login_useraddress, true, cmds[3], strictmode)

				} else {
					key := util.Yield16ByteKey([]byte(pass))
					v := util.AesEncrypt([]byte(cmds[2]), key)
					err = db.Put(cmds[1], v, Type.STRING, login_useraddress, false, "", strictmode)

				}
				if err != nil {
					fmt.Println("put", err)

				}
			case "puti32":
				// put age 15 int
				if len(cmds) < 3 {
					fmt.Println("格式错误  put key value [sharechan]")
					break
				}
				intv, err := strconv.ParseInt(cmds[2], 0, 32)
				if err != nil {
					fmt.Println(err)
					break
				}

				if len(cmds) >= 4 {
					v := util.Int32ToBytes(int32(intv))
					err = db.Put(cmds[1], v, Type.INT32, login_useraddress, true, cmds[3], strictmode)

				} else {
					key := util.Yield16ByteKey([]byte(pass))
					v := util.AesEncrypt(util.Int32ToBytes(int32(intv)), key)
					err = db.Put(cmds[1], v, Type.INT32, login_useraddress, false, "", strictmode)

				}
				if err != nil {
					fmt.Println("put", err)

				}
			case "puti64":
				// put age 15 int
				if len(cmds) < 3 {
					fmt.Println("格式错误  puti64 key value [sharechan]")
					break
				}
				intv, err := strconv.ParseInt(cmds[2], 0, 64)
				if err != nil {
					fmt.Println(err)
					break
				}

				if len(cmds) >= 4 {
					v := util.Int64Tobyte(int64(intv))
					err = db.Put(cmds[1], v, Type.INT64, login_useraddress, true, cmds[3], strictmode)

				} else {
					key := util.Yield16ByteKey([]byte(pass))
					v := util.AesEncrypt(util.Int32ToBytes(int32(intv)), key)
					err = db.Put(cmds[1], v, Type.INT64, login_useraddress, false, "", strictmode)

				}
				if err != nil {
					fmt.Println("put", err)

				}
			case "putstrarr":
				// put age 15 int
				if len(cmds) < 3 {
					fmt.Println("格式错误  putstrarr key [v1,...,vn] [sharechan]")
					break
				}
				vstr := strings.TrimLeft(cmds[2], "[")
				vstr = strings.TrimRight(vstr, "]")
				varr := strings.Split(vstr, ",")
				vbyte := []byte{}
				var err error
				for i := 0; i < len(varr); i++ {
					if i == len(varr)-1 {
						vbyte = append(vbyte, []byte(strings.TrimSpace(varr[i]))...)
						break
					}
					vbyte = append(vbyte, []byte(strings.TrimSpace(varr[i]))...)
					vbyte = append(vbyte, byte(0))
				}
				if len(cmds) >= 4 {
					err = db.Put(cmds[1], vbyte, Type.STRING_ARRAY, login_useraddress, true, cmds[3], strictmode)

				} else {
					key := util.Yield16ByteKey([]byte(pass))
					v := util.AesEncrypt((vbyte), key)
					err = db.Put(cmds[1], v, Type.STRING_ARRAY, login_useraddress, false, "", strictmode)

				}
				if err != nil {
					fmt.Println("put", err)

				}
			case "puti32arr":
				// put age 15 int
				if len(cmds) < 3 {
					fmt.Println("格式错误  puti32arr key [v1,...,vn] [sharechan]")
					break
				}
				vstr := strings.TrimLeft(cmds[2], "[")
				vstr = strings.TrimRight(vstr, "]")
				varr := strings.Split(vstr, ",")
				vbyte := []byte{}
				var err error
				i := 0
				for i = 0; i < len(varr); i++ {
					intv, e := strconv.ParseInt(strings.TrimSpace(varr[i]), 0, 32)
					if e != nil {
						fmt.Println(e)
						break
					}
					vbyte = append(vbyte, util.Int32ToBytes(int32(intv))...)
				}
				if i < len(varr) {
					break
				}
				if len(cmds) >= 4 {
					err = db.Put(cmds[1], vbyte, Type.INT32_ARRAY, login_useraddress, true, cmds[3], strictmode)

				} else {
					key := util.Yield16ByteKey([]byte(pass))
					v := util.AesEncrypt((vbyte), key)
					err = db.Put(cmds[1], v, Type.INT32_ARRAY, login_useraddress, false, "", strictmode)

				}
				if err != nil {
					fmt.Println("put", err)

				}
			case "puti64arr":
				// put age 15 int
				if len(cmds) < 3 {
					fmt.Println("格式错误  puti64arr key [v1,...,vn] [sharechan]")
					break
				}
				vstr := strings.TrimLeft(cmds[2], "[")
				vstr = strings.TrimRight(vstr, "]")
				varr := strings.Split(vstr, ",")
				vbyte := []byte{}
				var err error
				i := 0
				for i = 0; i < len(varr); i++ {
					intv, e := strconv.ParseInt(strings.TrimSpace(varr[i]), 0, 32)
					if e != nil {
						fmt.Println(err)
						break
					}
					vbyte = append(vbyte, util.Int64Tobyte(int64(intv))...)
				}
				if i < len(varr) {
					break
				}
				if len(cmds) >= 4 {
					err = db.Put(cmds[1], vbyte, Type.INT64_ARRAY, login_useraddress, true, cmds[3], strictmode)

				} else {
					key := util.Yield16ByteKey([]byte(pass))
					v := util.AesEncrypt((vbyte), key)
					err = db.Put(cmds[1], v, Type.INT64_ARRAY, login_useraddress, false, "", strictmode)

				}
				if err != nil {
					fmt.Println("put", err)

				}
			case "putstrset":
				// put age 15 int
				if len(cmds) < 3 {
					fmt.Println("格式错误  putstrset key [v1,...,vn] [sharechan]")
					break
				}
				vstr := strings.TrimLeft(cmds[2], "[")
				vstr = strings.TrimRight(vstr, "]")
				varr := strings.Split(vstr, ",")

				for i := 1; i < len(varr); i++ {
					tempi := i
					tempv := varr[i]
					for ; strings.TrimSpace(varr[tempi]) < strings.TrimSpace(varr[tempi-1]) && tempi > 0; tempi-- {
						varr[tempi] = varr[tempi-1]
					}
					varr[tempi] = tempv
				}

				restr := []string{}
				for i := 0; i < len(varr); i++ {
					if i == len(varr)-1 {
						restr = append(restr, strings.TrimSpace(varr[i]))
						break
					}
					if strings.TrimSpace(varr[i]) == strings.TrimSpace(varr[i+1]) {
						continue
					}
					restr = append(restr, strings.TrimSpace(varr[i]))
				}

				vbyte := []byte{}
				var err error
				for i := 0; i < len(restr); i++ {
					if i == len(restr)-1 {
						vbyte = append(vbyte, []byte(restr[i])...)
						break
					}
					vbyte = append(vbyte, []byte(restr[i])...)
					vbyte = append(vbyte, byte(0))
				}
				if len(cmds) >= 4 {
					err = db.Put(cmds[1], vbyte, Type.STRING_SET, login_useraddress, true, cmds[3], strictmode)

				} else {
					key := util.Yield16ByteKey([]byte(pass))
					v := util.AesEncrypt((vbyte), key)
					err = db.Put(cmds[1], v, Type.STRING_ARRAY, login_useraddress, false, "", strictmode)

				}
				if err != nil {
					fmt.Println("put", err)

				}
			case "puti32set":
				// put age 15 int
				if len(cmds) < 3 {
					fmt.Println("格式错误  puti32set key [v1,...,vn] [sharechan]")
					break
				}
				vstr := strings.TrimLeft(cmds[2], "[")
				vstr = strings.TrimRight(vstr, "]")
				varr := strings.Split(vstr, ",")
				for i := 1; i < len(varr); i++ {
					tempi := i
					tempv := varr[i]
					for ; strings.TrimSpace(varr[tempi]) < strings.TrimSpace(varr[tempi-1]) && tempi > 0; tempi-- {
						varr[tempi] = varr[tempi-1]
					}
					varr[tempi] = tempv
				}
				restr := []string{}
				for i := 0; i < len(varr); i++ {
					if i == len(varr)-1 {
						restr = append(restr, strings.TrimSpace(varr[i]))
						break
					}
					if strings.TrimSpace(varr[i]) == strings.TrimSpace(varr[i+1]) {
						continue
					}
					restr = append(restr, strings.TrimSpace(varr[i]))
				}

				vbyte := []byte{}
				var err error
				i := 0
				for i = 0; i < len(restr); i++ {
					intv, e := strconv.ParseInt(restr[i], 0, 32)
					if e != nil {
						fmt.Println(err)
						break
					}
					vbyte = append(vbyte, util.Int32ToBytes(int32(intv))...)
				}
				if i < len(restr) {
					break
				}
				if len(cmds) >= 4 {
					err = db.Put(cmds[1], vbyte, Type.INT32_SET, login_useraddress, true, cmds[3], strictmode)

				} else {
					key := util.Yield16ByteKey([]byte(pass))
					v := util.AesEncrypt((vbyte), key)
					err = db.Put(cmds[1], v, Type.INT32_SET, login_useraddress, false, "", strictmode)

				}
				if err != nil {
					fmt.Println("put", err)

				}
			case "puti64set":
				// put age 15 int
				if len(cmds) < 3 {
					fmt.Println("格式错误  puti64set key [v1,...,vn] [sharechan]")
					break
				}
				vstr := strings.TrimLeft(cmds[2], "[")
				vstr = strings.TrimRight(vstr, "]")
				varr := strings.Split(vstr, ",")
				for i := 1; i < len(varr); i++ {
					tempi := i
					tempv := varr[i]
					for ; strings.TrimSpace(varr[tempi]) < strings.TrimSpace(varr[tempi-1]) && tempi > 0; tempi-- {
						varr[tempi] = varr[tempi-1]
					}
					varr[tempi] = tempv
				}
				restr := []string{}
				for i := 0; i < len(varr); i++ {
					if i == len(varr)-1 {
						restr = append(restr, strings.TrimSpace(varr[i]))
						break
					}
					if strings.TrimSpace(varr[i]) == strings.TrimSpace(varr[i+1]) {
						continue
					}
					restr = append(restr, strings.TrimSpace(varr[i]))
				}

				vbyte := []byte{}
				var err error
				i := 0
				for i = 0; i < len(restr); i++ {
					intv, e := strconv.ParseInt(restr[i], 0, 32)
					if e != nil {
						fmt.Println(err)
						break
					}
					vbyte = append(vbyte, util.Int64Tobyte(int64(intv))...)
				}
				if i < len(restr) {
					break
				}
				if len(cmds) >= 4 {
					err = db.Put(cmds[1], vbyte, Type.INT64_ARRAY, login_useraddress, true, cmds[3], strictmode)

				} else {
					key := util.Yield16ByteKey([]byte(pass))
					v := util.AesEncrypt((vbyte), key)
					err = db.Put(cmds[1], v, Type.INT64_ARRAY, login_useraddress, false, "", strictmode)

				}
				if err != nil {
					fmt.Println("put", err)

				}
			case "del":
				if len(cmds) < 2 {
					fmt.Println("格式错误  del [key] [sharechan]")
					break
				}
				// del age
				if len(cmds) >= 3 {
					db.Del(cmds[1], login_useraddress, true, cmds[2], strictmode)
				} else {
					db.Del(cmds[1], login_useraddress, false, "", strictmode)
				}
			case "get":
				if len(cmds) < 2 {
					fmt.Println("格式错误  get key  [sharechan]")
					break
				}
				// get age
				var block *BC.Block
				var index int
				pre := time.Now().UnixNano()
				if len(cmds) >= 3 {
					block, index = db.Get(cmds[1], login_username, login_useraddress, true, cmds[2])

				} else {
					block, index = db.Get(cmds[1], login_username, login_useraddress, false, "")

				}
				if block != nil {
					// fmt.Println("blockid:", block.BlockId)
					// fmt.Println("block_hash", block.Hash)
					// fmt.Println("pre_block_hash:", base64.RawStdEncoding.EncodeToString(block.PreBlockHash))
					// fmt.Println("block_hash:", base64.RawStdEncoding.EncodeToString(block.Hash))
					tx := block.TxInfos[index]
					var v []byte
					if !tx.Share {
						key := util.Yield16ByteKey([]byte(pass))
						v = util.AesDecrypt(tx.Value, key)
					} else {
						v = util.AesDecrypt(tx.Value, BC.LocalWallets.ShareChanMap[tx.ShareChan].Key)
					}
					switch tx.DataType {
					case Type.STRING:
						fmt.Println(string(v))
					case Type.INT32:
						fmt.Println(util.BytesToInt32(v))
					case Type.INT64:
						fmt.Println(util.BytesToInt64(v))
					case Type.STRING_ARRAY:
						fmt.Println(Type.ConvertToSTRING_ARRAY(v))
					case Type.INT32_ARRAY:
						fmt.Println(Type.ConvertToINT32_ARRAY(v))
					case Type.INT64_ARRAY:
						fmt.Println(Type.ConvertToINT64_ARRAY(v))
					case Type.STRING_SET:
						fmt.Println(Type.ConvertToSTRING_SET(v))
					case Type.INT32_SET:
						fmt.Println(Type.ConvertToINT32_SET(v))
					case Type.INT64_SET:
						fmt.Println(Type.ConvertToINT64_SET(v))
					}

					cur := time.Now().UnixNano()
					fmt.Println("耗时:", (cur-pre)/1000000, "(ms)")
				} else {
					fmt.Println("<nil>")
				}
			case "print_quorum":
				for _, node := range LocalNode.Quorum {
					if node.LocalIp == LocalNode.LocalIp {
						fmt.Println(node.LocalIp, "(本机)")
					}
				}
			case "detail":
				if len(cmds) < 2 {
					fmt.Println("格式错误  detail BlockId")
					break
				}
				LocalBlockChain.Traverse(func(block *BC.Block, err error) bool {
					if err != nil {
						return true
					}
					if fmt.Sprintf("%d", block.BlockId) == cmds[1] {
						fmt.Println("blockid", block.BlockId)
						fmt.Println("hash", base64.RawStdEncoding.EncodeToString(block.Hash))
						fmt.Println("prevhash", base64.RawStdEncoding.EncodeToString(block.PreBlockHash))
						fmt.Println("timestamp", block.Timestamp)
						fmt.Println("txnums", len(block.TxInfos))
						// for i, tx := range block.TxInfos {
						// 	fmt.Println("交易索引:", i)
						// 	fmt.Println("user_address:", BC.GenerateAddressFromPubkey(tx.PublicKey))
						// 	fmt.Println("key-value:", tx.Key, string(tx.Value))
						// 	fmt.Println("sharemode:", tx.Share)
						// 	fmt.Println("datatype:", tx.DataType)
						// 	fmt.Println("sharechan:", tx.ShareChan)
						// }
					}
					return true
				})
			case "isaccountant":
				fmt.Println(quorum.LocalNodeIsAccount())
			case "print":
				LocalBlockChain.Traverse(func(block *BC.Block, err error) bool {
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
				b, _ := LocalBlockChain.GetTailBlock()
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

				b, e := LocalBlockChain.GetBlockByHash(_hash)
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
					b, _ = LocalBlockChain.GetBlockByHash(_hash)
				}
			case "print_local_wallet":
				for addr, w := range BC.LocalWallets.WalletsMap {
					fmt.Println(w.Username, addr)
				}
			case "systxrate":
				// fmt.Println("----------------------------------------")
				test.SystemInfo.PrintTxInfo()
				// fmt.Println("----------------------------------------")
			case "resetsystxi":
				test.SystemInfo.CleanTxInfo()
			case "sysblockrate":
				// fmt.Println("----------------------------------------")
				test.SystemInfo.PrintBlockInfo()
				// fmt.Println("----------------------------------------")
			case "resetsysbki":
				test.SystemInfo.CleanBlockInfo()
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
	LocalNode, err = quorum.LoadGenesisFile(genesis_file_name)
	fmt.Println("监听地址:", LocalNode.LocalIp, LocalNode.LocalPort)
	fmt.Println("localNode.Quorum:")
	for _, v := range LocalNode.Quorum {
		fmt.Println(v.LocalIp)
	}
	if err != nil {
		panic(err)
	}

	LocalBlockChain = BC.NewBlockChain(
		LocalNode.BCInfo.BlockTailHashKey,

		LocalNode.BCInfo.BlockChainDB,
		func(name string) {
			quorum.GetShareChan(name)
		},
	)

	BC.LoadLocalWallets()
	_, err = LocalBlockChain.GetAddressFromUsername("liangchen")
	if err != nil {
		wa := BC.NewWallet("liangchen", LocalNode.BCInfo.PassWorld)
		wa.Private = LocalNode.BCInfo.PriKey
		wa.PubKey = LocalNode.BCInfo.PubKey
		BC.LocalWallets.WalletsMap[wa.NewAddress()] = wa
		BC.LocalWallets.SaveToFile()
	}
	quorum.Broadcast(LocalBlockChain)

	newbllocks, e := quorum.BlockSynchronization()
	if e == nil {
		addblocks(newbllocks)
	}

	_tailblock, _ := LocalBlockChain.GetTailBlock()
	if _tailblock == nil {
		// 创建创世块
		genesis_block := BC.NewGenesisBlock()
		genesis_block.BlockId = 1

		rw := BC.LocalWallets.GetBlockChainRootWallet()
		LocalBlockChain.SignBlock(rw.Private, true, genesis_block)
		flag := LocalBlockChain.VerifyBlock(rw.PubKey, genesis_block)
		if !flag {
			fmt.Println("创世块校验失败")
		}
		err = LocalBlockChain.AddBlock(genesis_block)
		if err != nil {
			panic(err)
		}
		// BC.LocalWallets.TailBlockHashMap[rw.NewAddress()] = genesis_block.Hash
		BC.LocalWallets.SetTailBlockHash(rw.NewAddress(), genesis_block.Hash)
		BC.LocalWallets.SaveToFile()
	}

	go view.Run(LocalBlockChain, LocalNode)
	quorum.StartGrpcWork()
	StartDraftWork()

	db.Run(LocalBlockChain, LocalNode)
	go uc.Run()
	// fmt.Println("hello world")
	// Test.Test1()
	// Test.Test3()
	runLocalTestCli()
}
