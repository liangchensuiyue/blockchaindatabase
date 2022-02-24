package database

import (
	"encoding/base64"
	"errors"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	"go_code/基于区块链的非关系型数据库/quorum"
	"strings"
	"time"
)

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
	tx, e := BC.NewTransaction("create_user", username, []byte(base64.RawStdEncoding.EncodeToString([]byte(passworld))+" "+wa.NewAddress()), "string", user_address, false, []string{})
	if e != nil {
		return errors.New("创建用户失败")
	}

	lcdraft := BC.GetLocalDraft()
	newblock, _ := lcdraft.PackBlock(tx)
	rw := BC.LocalWallets.GetBlockChainRootWallet()
	// localBlockChain.SignBlock(rw.Private, false, newblock)

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

				BC.LocalWallets.TailBlockHashMap[user_address] = newblock.Hash
				BC.LocalWallets.WalletsMap[wa.NewAddress()] = wa

				BC.LocalWallets.SaveToFile()
				fmt.Println("校验成功")
				return
			}
			fmt.Println("区块校验失败")
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
func PutTest(key string, value []byte, datatype string, user_address string, share bool, shareuser []string, strict bool, TestHandler func()) error {
	shareaddress := []string{}
	if share {
		for i := 0; i < len(shareuser); i++ {
			addr, e := GetAddressFromUsername(shareuser[i])
			if addr == user_address {
				continue
			}
			if e != nil {
				return e
			}
			shareaddress = append(shareaddress, addr)
		}
	} else {
		shareuser = []string{}
	}
	if len(shareaddress) == 0 {
		share = false
	} else {
		shareaddress = append(shareaddress, user_address)
	}
	tx, e := BC.NewTransaction("put", key, value, datatype, user_address, share, shareaddress)
	if e != nil {
		TestHandler()
		go quorum.Request(user_address, false, &BC.Transaction{
			Key:          key,
			Value:        value,
			DelMark:      false,
			Share:        share,
			DataType:     datatype,
			Timestamp:    uint64(time.Now().Unix()),
			ShareAddress: shareuser,
		})
		return nil
	}
	if strict {
		lcdraft := BC.GetLocalDraft()
		newblock, _ := lcdraft.PackBlock(tx)
		rw := BC.LocalWallets.GetBlockChainRootWallet()
		// localBlockChain.SignBlock(rw.Private, false, newblock)
		BC.BlockQueue.Insert(BC.QueueObject{
			TargetBlock: newblock,
			Handle: func(total, fail int) {
				flag := localBlockChain.VerifyBlock(rw.PubKey, newblock)
				if flag {
					e := localBlockChain.AddBlock(newblock)
					TestHandler()
					if e != nil {
						fmt.Println(e)
						return
					}

					for _, tx := range newblock.TxInfos {

						// fmt.Println(tx.Share, tx.ShareAddress, base64.RawStdEncoding.EncodeToString(newblock.Hash))
						if tx.Share {
							BC.LocalWallets.ShareTailBlockHashMap[BC.GenerateUserShareKey(tx.ShareAddress)] = newblock.Hash

						} else {
							BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
						}

					}
					BC.LocalWallets.SaveToFile()
					fmt.Println("block:", newblock.BlockId, "校验成功")
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

func Put(key string, value []byte, datatype string, user_address string, share bool, shareuser []string, strict bool) error {
	shareaddress := []string{}
	if share {
		for i := 0; i < len(shareuser); i++ {
			addr, e := GetAddressFromUsername(shareuser[i])
			if addr == user_address {
				continue
			}
			if e != nil {
				return e
			}
			shareaddress = append(shareaddress, addr)
		}
	} else {
		shareuser = []string{}
	}
	if len(shareaddress) == 0 {
		share = false
	} else {
		shareaddress = append(shareaddress, user_address)
	}
	tx, e := BC.NewTransaction("put", key, value, datatype, user_address, share, shareaddress)
	if e != nil {
		go quorum.Request(user_address, false, &BC.Transaction{
			Key:          key,
			Value:        value,
			DelMark:      false,
			Share:        share,
			DataType:     datatype,
			Timestamp:    uint64(time.Now().Unix()),
			ShareAddress: shareuser,
		})
		return nil
	}
	if strict {
		lcdraft := BC.GetLocalDraft()
		newblock, _ := lcdraft.PackBlock(tx)
		rw := BC.LocalWallets.GetBlockChainRootWallet()
		// localBlockChain.SignBlock(rw.Private, false, newblock)
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
							BC.LocalWallets.ShareTailBlockHashMap[BC.GenerateUserShareKey(tx.ShareAddress)] = newblock.Hash

						} else {
							BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
						}

					}
					BC.LocalWallets.SaveToFile()
					fmt.Println("block:", newblock.BlockId, "校验成功")
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
func Del(key string, user_address string, share bool, shareuser []string, strict bool) {
	shareaddress := []string{}
	if share {
		for i := 0; i < len(shareuser); i++ {
			addr, e := GetAddressFromUsername(shareuser[i])
			if user_address == addr {
				continue
			}
			if e != nil {
				return
			}
			shareaddress = append(shareaddress, addr)
		}
	} else {
		shareuser = []string{}
	}
	if len(shareaddress) == 0 {
		share = false
	} else {
		shareaddress = append(shareaddress, user_address)
	}
	tx, e := BC.NewTransaction("del", key, []byte{}, "", user_address, share, shareaddress)
	if e != nil {
		go quorum.Request(user_address, false, &BC.Transaction{
			Key:          key,
			DelMark:      true,
			Share:        share,
			Timestamp:    uint64(time.Now().Unix()),
			ShareAddress: shareuser,
		})
		return
	}
	if strict {
		lcdraft := BC.GetLocalDraft()
		newblock, _ := lcdraft.PackBlock(tx)
		rw := BC.LocalWallets.GetBlockChainRootWallet()

		// localBlockChain.SignBlock(rw.Private, false, newblock)
		BC.BlockQueue.Insert(BC.QueueObject{
			TargetBlock: newblock,
			Handle: func(total, fail int) {
				flag := localBlockChain.VerifyBlock(rw.PubKey, newblock)
				if flag {
					localBlockChain.AddBlock(newblock)

					for _, tx := range newblock.TxInfos {

						// fmt.Println(tx.Share, tx.ShareAddress, base64.RawStdEncoding.EncodeToString(newblock.Hash))
						if tx.Share {
							BC.LocalWallets.ShareTailBlockHashMap[BC.GenerateUserShareKey(tx.ShareAddress)] = newblock.Hash

						} else {
							BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
						}

					}
					BC.LocalWallets.SaveToFile()
					fmt.Println("block:", newblock.BlockId, "校验成功")
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
func Get(key string, user_address string, sharemode bool, shareuser []string) (*BC.Block, int) {
	// user_wa, _ := BC.LocalWallets.GetUserWallet(user_address)
	var tailhash []byte
	shareaddress := []string{}
	for j := 0; j < len(shareuser); j++ {
		addr, e := GetAddressFromUsername(shareuser[j])
		if addr == user_address {
			continue
		}
		if e == nil {
			shareaddress = append(shareaddress, addr)
		}
	}
	if len(shareaddress) == 0 {
		sharemode = false
	} else {
		shareaddress = append(shareaddress, user_address)
	}
	shareaddressKey := BC.GenerateUserShareKey(shareaddress)
	if sharemode {
		tailhash, _ = BC.LocalWallets.ShareTailBlockHashMap[shareaddressKey]
	} else {
		tailhash, _ = BC.LocalWallets.TailBlockHashMap[user_address]
	}
	// fmt.Println("tailhash", base64.RawStdEncoding.EncodeToString(tailhash))
	// for k, v := range BC.LocalWallets.ShareTailBlockHashMap {
	// 	fmt.Println(k, base64.RawStdEncoding.EncodeToString(v))
	// }
	_index := -1
	fmt.Println("length", BC.BlockQueue.Len())
	_b, e := BC.BlockQueue.Find(func(b *BC.Block) bool {
		for i := len(b.TxInfos) - 1; i >= 0; i-- {
			if sharemode {
				if b.TxInfos[i].Share && BC.GenerateUserShareKey(b.TxInfos[i].ShareAddress) == shareaddressKey {
					if b.TxInfos[i].Key == key {
						if !b.TxInfos[i].DelMark {
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
						if !b.TxInfos[i].DelMark {
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
			fmt.Println(e, tailhash)
			return nil, -1
		}
		for i := len(b.TxInfos) - 1; i >= 0; i-- {
			if sharemode {
				if b.TxInfos[i].Share && BC.GenerateUserShareKey(b.TxInfos[i].ShareAddress) == shareaddressKey {
					if b.TxInfos[i].Key == key {
						if !b.TxInfos[i].DelMark {
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
						if !b.TxInfos[i].DelMark {
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
			if tx.Key == username {
				addr := strings.Split(string(tx.Value), " ")[1]
				return addr, nil
			}
		}
		b, _ = localBlockChain.GetBlockByHash(_hash)
	}
	return "", errors.New("未知的用户")
}
