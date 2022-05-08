package database

import (
	"encoding/base64"
	"errors"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	"go_code/基于区块链的非关系型数据库/quorum"
	Type "go_code/基于区块链的非关系型数据库/type"
	"go_code/基于区块链的非关系型数据库/util"
	"strings"
	"time"
)

func DelUser(username string) {
	root_address := BC.LocalWallets.GetBlockChainRootWallet().NewAddress()
	addr, err := BC.GetAddressFromUsername(username)
	if err != nil {
		return
	}
	tx, e := BC.NewTransaction(username, []byte(addr), Type.NEW_USER, root_address, false, "")
	if e != nil {
		return
	}

	lcdraft := BC.GetLocalDraft()
	newblock, _ := lcdraft.PackBlock(tx)
	rw := BC.LocalWallets.GetBlockChainRootWallet()
	// localBlockChain.SignBlock(rw.Private, false, newblock)
	BC.Global_DICT.Put(tx.Key, tx)
	BC.BlockQueue.Insert(BC.QueueObject{
		TargetBlock: newblock,
		Handle: func(total, fail int) {
			flag := localBlockChain.VerifyBlock(rw.PubKey, newblock)
			if flag {
				e := localBlockChain.AddBlock(newblock)
				if e != nil {
					// fmt.Println(e)
					return
				}

				BC.LocalWallets.TailBlockHashMap[root_address] = newblock.Hash

				BC.LocalWallets.SaveToFile()
				// fmt.Println("校验成功")
				return
			}
			fmt.Println("block:", newblock.BlockId, "校验失败")
		},
	})

}
func CreateUser(username string, passworld string) error {
	fmt.Println("正在创建用户...")
	user_address := BC.LocalWallets.GetBlockChainRootWallet().NewAddress()

	// 判断用户是否创建
	_hash, _ := BC.LocalWallets.GetUserTailBlockHash(user_address)

	b, e := localBlockChain.GetBlockByHash(_hash)
	if e != nil {
		return e
	}
	for {
		if b.IsGenesisBlock() {
			break
		}
		for _, tx := range b.TxInfos {
			_hash = tx.PreBlockHash
			if tx.Key == username {
				return errors.New("该用户已被创建")
			}
		}
		b, _ = localBlockChain.GetBlockByHash(_hash)
	}

	wa := BC.NewWallet(username, passworld)
	tx, e := BC.NewTransaction(username, []byte(base64.RawStdEncoding.EncodeToString([]byte(passworld))+" "+wa.NewAddress()), Type.NEW_USER, user_address, false, "")
	if e != nil {
		return errors.New("创建用户失败")
	}

	lcdraft := BC.GetLocalDraft()
	newblock, _ := lcdraft.PackBlock(tx)
	rw := BC.LocalWallets.GetBlockChainRootWallet()
	// localBlockChain.SignBlock(rw.Private, false, newblock)
	BC.Global_DICT.Put(tx.Key, tx)
	BC.BlockQueue.Insert(BC.QueueObject{
		TargetBlock: newblock,
		Handle: func(total, fail int) {
			flag := localBlockChain.VerifyBlock(rw.PubKey, newblock)
			if flag {
				e := localBlockChain.AddBlock(newblock)
				if e != nil {
					// fmt.Println(e)
					return
				}

				BC.LocalWallets.TailBlockHashMap[user_address] = newblock.Hash
				BC.LocalWallets.WalletsMap[wa.NewAddress()] = wa

				BC.LocalWallets.SaveToFile()
				// fmt.Println("校验成功")
				return
			}
			fmt.Println("block:", newblock.BlockId, "校验失败")
		},
	})

	return nil
}

func VeriftUser(username string, passworld string) error {
	user_address := BC.LocalWallets.GetBlockChainRootWallet().NewAddress()

	// 判断用户是否创建
	_hash, _ := BC.LocalWallets.GetUserTailBlockHash(user_address)

	b, e := localBlockChain.GetBlockByHash(_hash)
	if e != nil {
		return e
	}
	for {
		if b.IsGenesisBlock() {
			break
		}
		for _, tx := range b.TxInfos {
			_hash = tx.PreBlockHash
			if tx.Key == username {
				passw := strings.Split(string(tx.Value), " ")[0]
				if passw == base64.RawStdEncoding.EncodeToString([]byte(passworld)) {
					return nil
				}
				return errors.New("密码错误")
			}
		}
		b, _ = localBlockChain.GetBlockByHash(_hash)
	}
	return errors.New("未知的用户")
}
func PutTest(key string, value []byte, datatype int32, user_address string, share bool, shareChan string, strict bool, TestHandler func()) error {
	if !(share && BC.LocalWallets.HasShareChan(shareChan)) {
		return errors.New("指定的 sharechan 不存在")
	}
	tx, e := BC.NewTransaction(key, util.AesEncrypt(value, BC.LocalWallets.ShareChanMap[shareChan].Key), datatype, user_address, share, shareChan)
	if e != nil {
		TestHandler()
		quorum.Request(user_address, true, &BC.Transaction{
			Key:       key,
			Value:     value,
			Share:     share,
			DataType:  datatype,
			Timestamp: uint64(time.Now().Unix()),
			ShareChan: shareChan,
		})
		return nil
	}
	if strict {
		lcdraft := BC.GetLocalDraft()
		newblock, _ := lcdraft.PackBlock(tx)
		rw := BC.LocalWallets.GetBlockChainRootWallet()
		// localBlockChain.SignBlock(rw.Private, false, newblock)
		BC.Global_DICT.Put(tx.Key, tx)
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

						// fmt.Println(tx.Share, tx.ShareAddress, base64.RawStdEncoding.EncodeToString(newblock.Hash))
						if tx.Share {
							scn := BC.LocalWallets.ShareChanMap[tx.ShareChan]
							scn.TailBlockHash = newblock.Hash
							// BC.LocalWallets.ShareTailBlockHashMap[BC.GenerateUserShareKey(tx.ShareAddress)] = newblock.Hash

						} else {
							BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
						}

					}
					BC.LocalWallets.SaveToFile()
					// fmt.Println("block:", newblock.BlockId, "校验成功")
					return
				}
				fmt.Println("block:", newblock.BlockId, "校验失败")
			},
		})

	} else {
		draft := BC.GetLocalDraft()
		draft.PutTx(tx)
	}

	return nil
}

// var NUM = 0
// var Total int64 = 0
// var pre int64

func Put(key string, value []byte, datatype int32, user_address string, share bool, shareChan string, strict bool) error {
	// NUM++
	// pre = time.Now().UnixNano()
	if share {
		if !BC.LocalWallets.HasShareChan(shareChan) {
			quorum.GetShareChan(shareChan)
			if !BC.LocalWallets.HasShareChan(shareChan) {
				return errors.New("指定的 sharechan 不存在")
			}
		}
		value = util.AesEncrypt(value, BC.LocalWallets.ShareChanMap[shareChan].Key)
	} else {
		shareChan = ""
	}
	tx, e := BC.NewTransaction(key, value, datatype, user_address, share, shareChan)
	if e == nil && !tx.VerifySimple() {
		return errors.New("操作有误")
		// 交易校验失败
	}
	if e != nil {
		quorum.Request(user_address, true, &BC.Transaction{
			Key:       key,
			Value:     value,
			Share:     share,
			DataType:  datatype,
			Timestamp: uint64(time.Now().Unix()),
			ShareChan: shareChan,
		})
		return nil
	}
	if strict {
		lcdraft := BC.GetLocalDraft()
		newblock, _ := lcdraft.PackBlock(tx)
		rw := BC.LocalWallets.GetBlockChainRootWallet()
		// localBlockChain.SignBlock(rw.Private, false, newblock)
		BC.Global_DICT.Put(tx.Key, tx)
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

						// fmt.Println(tx.Share, tx.ShareAddress, base64.RawStdEncoding.EncodeToString(newblock.Hash))
						if tx.Share {
							scn := BC.LocalWallets.ShareChanMap[tx.ShareChan]
							scn.TailBlockHash = newblock.Hash
							// BC.LocalWallets.ShareTailBlockHashMap[BC.GenerateUserShareKey(tx.ShareAddress)] = newblock.Hash

						} else {
							BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
						}

					}
					BC.LocalWallets.SaveToFile()
					// fmt.Println("block:", newblock.BlockId, "校验成功")
					return
				}
				fmt.Println("block:", newblock.BlockId, "校验失败")
			},
		})
		// Total += (time.Now().UnixNano() - pre)

	} else {
		draft := BC.GetLocalDraft()
		draft.PutTx(tx)
	}

	return nil
}

// func Put(key string, value []byte, datatype string, user_address string, share bool, shareuser []string, strict bool) error {
// 	shareaddress := []string{}
// 	if share {
// 		for i := 0; i < len(shareuser); i++ {
// 			addr, e := GetAddressFromUsername(shareuser[i])
// 			if addr == user_address {
// 				continue
// 			}
// 			if e != nil {
// 				return e
// 			}
// 			shareaddress = append(shareaddress, addr)
// 		}
// 	} else {
// 		shareuser = []string{}
// 	}
// 	if len(shareaddress) == 0 {
// 		share = false
// 	} else {
// 		shareaddress = append(shareaddress, user_address)
// 	}
// 	tx, e := BC.NewTransaction("put", key, value, datatype, user_address, share, shareaddress)
// 	if e != nil {
// 		quorum.Request(user_address, true, &BC.Transaction{
// 			Key:          key,
// 			Value:        value,
// 			DelMark:      false,
// 			Share:        share,
// 			DataType:     datatype,
// 			Timestamp:    uint64(time.Now().Unix()),
// 			ShareAddress: shareuser,
// 		})
// 		return nil
// 	}
// 	if strict {
// 		lcdraft := BC.GetLocalDraft()
// 		newblock, _ := lcdraft.PackBlock(tx)
// 		rw := BC.LocalWallets.GetBlockChainRootWallet()
// 		// localBlockChain.SignBlock(rw.Private, false, newblock)
// 		BC.BlockQueue.Insert(BC.QueueObject{
// 			TargetBlock: newblock,
// 			Handle: func(total, fail int) {
// 				flag := localBlockChain.VerifyBlock(rw.PubKey, newblock)
// 				if flag {
// 					e := localBlockChain.AddBlock(newblock)
// 					if e != nil {
// 						fmt.Println(e)
// 						return
// 					}

// 					for _, tx := range newblock.TxInfos {

// 						// fmt.Println(tx.Share, tx.ShareAddress, base64.RawStdEncoding.EncodeToString(newblock.Hash))
// 						if tx.Share {
// 							BC.LocalWallets.ShareTailBlockHashMap[BC.GenerateUserShareKey(tx.ShareAddress)] = newblock.Hash

// 						} else {
// 							BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
// 						}

// 					}
// 					BC.LocalWallets.SaveToFile()
// 					// fmt.Println("block:", newblock.BlockId, "校验成功")
// 					return
// 				}
// 				fmt.Println("block:", newblock.BlockId, "校验失败")
// 			},
// 		})

// 	} else {
// 		draft := BC.GetLocalDraft()
// 		draft.PutTx(tx)
// 	}

// 	return nil
// }
func Del(key string, user_address string, share bool, sharechan string, strict bool) {
	if share {
		if !BC.LocalWallets.HasShareChan(sharechan) {
			quorum.GetShareChan(sharechan)
			if !BC.LocalWallets.HasShareChan(sharechan) {
				return
			}
			return
		}
	} else {
		sharechan = ""
	}
	tx, e := BC.NewTransaction(key, []byte{}, Type.DEL_KEY, user_address, share, sharechan)
	if e != nil {
		go quorum.Request(user_address, true, &BC.Transaction{
			Key:       key,
			Share:     share,
			Timestamp: uint64(time.Now().Unix()),
			ShareChan: sharechan,
		})
		return
	}
	if strict {
		lcdraft := BC.GetLocalDraft()
		newblock, _ := lcdraft.PackBlock(tx)
		rw := BC.LocalWallets.GetBlockChainRootWallet()

		// localBlockChain.SignBlock(rw.Private, false, newblock)
		BC.Global_DICT.Put(tx.Key, tx)

		BC.BlockQueue.Insert(BC.QueueObject{
			TargetBlock: newblock,
			Handle: func(total, fail int) {
				flag := localBlockChain.VerifyBlock(rw.PubKey, newblock)
				if flag {
					localBlockChain.AddBlock(newblock)

					for _, tx := range newblock.TxInfos {

						// fmt.Println(tx.Share, tx.ShareAddress, base64.RawStdEncoding.EncodeToString(newblock.Hash))
						if tx.Share {
							scn := BC.LocalWallets.ShareChanMap[tx.ShareChan]
							scn.TailBlockHash = newblock.Hash

						} else {
							BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
						}

					}
					BC.LocalWallets.SaveToFile()
					// fmt.Println("block:", newblock.BlockId, "校验成功")
					return
				}
				fmt.Println("block:", newblock.BlockId, "校验失败")
				return

			},
		})

	} else {
		draft := BC.GetLocalDraft()
		draft.PutTx(tx)
	}

	return
}
func Get(key string, username string, user_address string, sharemode bool, sharechan string) (*BC.Block, int) {
	// user_wa, _ := BC.LocalWallets.GetUserWallet(user_address)
	var tailhash []byte

	schn := BC.LocalWallets.ShareChanMap[sharechan]
	if sharemode {
		if !BC.LocalWallets.HasShareChan(sharechan) {
			return nil, -1
		}
		if !schn.HasUser(username) {
			return nil, -1
		}
		tailhash = schn.TailBlockHash
	} else {
		tailhash, _ = BC.LocalWallets.TailBlockHashMap[user_address]
	}
	// fmt.Println("tailhash", base64.RawStdEncoding.EncodeToString(tailhash))
	// for k, v := range BC.LocalWallets.ShareTailBlockHashMap {
	// 	fmt.Println(k, base64.RawStdEncoding.EncodeToString(v))
	// }
	txs := BC.Global_DICT.Get(key)

	// if len(txs) == 0 {
	// 	return nil, -1
	// }

	// 从缓存里面去找
	for _, TxInfo := range txs {
		if sharemode {
			if TxInfo.Share && TxInfo.ShareChan == sharechan {
				if TxInfo.Key == key {
					if TxInfo.DataType != Type.DEL_KEY {
						return &BC.Block{TxInfos: []*BC.Transaction{TxInfo}}, 0

					} else {
						return nil, -1
					}
				}

			}
		} else {
			if BC.GenerateAddressFromPubkey(TxInfo.PublicKey) == user_address {
				if TxInfo.Key == key {
					if TxInfo.DataType != Type.DEL_KEY {
						return &BC.Block{TxInfos: []*BC.Transaction{TxInfo}}, 0
					} else {
						return nil, -1
					}
				}

			}
		}
	}

	// 从区块队列上去找
	_index := -1
	_b, e := BC.BlockQueue.Find(func(b *BC.Block) bool {
		for i := len(b.TxInfos) - 1; i >= 0; i-- {
			if sharemode {
				if b.TxInfos[i].Share && b.TxInfos[i].ShareChan == sharechan {
					if b.TxInfos[i].Key == key {
						if b.TxInfos[i].DataType != Type.DEL_KEY {
							_index = i
							return false
						} else {
							_index = -1
							return false
						}
					}

				}
			} else {
				if BC.GenerateAddressFromPubkey(b.TxInfos[i].PublicKey) == user_address {
					if b.TxInfos[i].Key == key {
						if b.TxInfos[i].DataType != Type.DEL_KEY {
							_index = i
							return false
						} else {
							_index = -1
							return false
						}
					}

				}
			}
		}
		return true
	})
	if e == nil {
		return _b, _index
	}
	for {
		b, e := localBlockChain.GetBlockByHash(tailhash)
		if e != nil || b.IsGenesisBlock() {
			// fmt.Println(e, tailhash)
			return nil, -1
		}
		for i := len(b.TxInfos) - 1; i >= 0; i-- {
			if sharemode {
				if b.TxInfos[i].Share && b.TxInfos[i].ShareChan == sharechan {
					if b.TxInfos[i].Key == key {
						if b.TxInfos[i].DataType != Type.DEL_KEY {
							return b, i
						} else {
							return nil, -1
						}
					}

					tailhash = b.TxInfos[i].PreBlockHash
				}
			} else {
				if BC.GenerateAddressFromPubkey(b.TxInfos[i].PublicKey) == user_address {
					if b.TxInfos[i].Key == key {
						if b.TxInfos[i].DataType != Type.DEL_KEY {
							return b, i
						} else {
							return nil, -1
						}
					}

					tailhash = b.TxInfos[i].PreBlockHash
				}
			}
		}
	}
}

func GetAddressFromUsername(username string) (string, error) {
	user_address := BC.LocalWallets.GetBlockChainRootWallet().NewAddress()

	// 判断用户是否创建
	_hash, _ := BC.LocalWallets.GetUserTailBlockHash(user_address)

	b, e := localBlockChain.GetBlockByHash(_hash)
	if e != nil {
		return "", e
	}
	for {
		if b.IsGenesisBlock() {
			break
		}
		for _, tx := range b.TxInfos {
			// fmt.Println("t.Key", tx.Key, username)
			_hash = tx.PreBlockHash
			if tx.Key == username && tx.DataType == Type.NEW_USER {
				addr := strings.Split(string(tx.Value), " ")[1]
				return addr, nil
			}
			if tx.Key == username && tx.DataType == Type.DEL_USER {
				break
			}
		}
		b, _ = localBlockChain.GetBlockByHash(_hash)
	}
	return "", errors.New("未知的用户")
}
