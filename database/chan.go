package database

import (
	"encoding/base64"
	"errors"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	Type "go_code/基于区块链的非关系型数据库/type"
	"strings"

	"go_code/基于区块链的非关系型数据库/quorum"
	"go_code/基于区块链的非关系型数据库/util"
	"time"
)

func JoinChan(channame, username, user_address string, creator, joinkey string, key string) error {
	_joinkey, e1 := base64.RawStdEncoding.DecodeString(joinkey)

	_key, e2 := base64.RawStdEncoding.DecodeString(key)
	if e1 != nil || e2 != nil {
		return errors.New("密钥错误")
	}
	_, _ok := BC.LocalWallets.ShareChanMap[creator+"."+channame]
	if !_ok {
		BC.LocalWallets.ShareChanMap[creator+"."+channame] = &BC.ShareChan{
			Channame: channame,
			Key:      _key,
			JoinKey:  util.AesEncrypt(_joinkey, _key),
			Creator:  creator,
		}
		BC.LocalWallets.SaveToFile()
	}
	value := []byte(username + " ")
	value = append(value, util.AesEncrypt(_joinkey, _key)...)
	tx, e := BC.NewTransaction(creator+"."+channame, value, Type.JOIN_CHAN, user_address, false, "")
	if e != nil {
		quorum.Request(user_address, true, &BC.Transaction{
			Key:       creator + "." + channame,
			Share:     false,
			Value:     value,
			DataType:  Type.JOIN_CHAN,
			Timestamp: uint64(time.Now().Unix()),
			ShareChan: "",
		})
		return nil
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
				for _, tx := range newblock.TxInfos {

					// fmt.Println(tx.Share, tx.ShareAddress, base64.RawStdEncoding.EncodeToString(newblock.Hash))
					if tx.Share {
						scn := BC.LocalWallets.ShareChanMap[tx.ShareChan]
						scn.TailBlockHash = newblock.Hash
						// BC.LocalWallets.ShareTailBlockHashMap[BC.GenerateUserShareKey(tx.ShareAddress)] = newblock.Hash

					} else {
						// BC.WalletsLock.Lock()
						// BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
						// BC.WalletsLock.Unlock()
						BC.LocalWallets.SetTailBlockHash(BC.GenerateAddressFromPubkey(tx.PublicKey), newblock.Hash)

					}

				}
				BC.LocalWallets.SaveToFile()
				// fmt.Println("block:", newblock.BlockId, "校验成功")
				return
			}
			fmt.Println("block:", newblock.BlockId, "校验失败")
		},
	})
	return nil
}
func ExitChan(creator, channame, username, user_address string) {
	if BC.UserIsChanCreator(channame, user_address) {
		return
	}
	tx, e := BC.NewTransaction(creator+"."+channame, []byte(username), Type.EXIT_CHAN, user_address, false, "")
	if e != nil {
		quorum.Request(user_address, true, &BC.Transaction{
			Key:       creator + "." + channame,
			Share:     false,
			Value:     []byte(username),
			DataType:  Type.EXIT_CHAN,
			Timestamp: uint64(time.Now().Unix()),
			ShareChan: "",
		})
		return
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
				for _, tx := range newblock.TxInfos {

					// fmt.Println(tx.Share, tx.ShareAddress, base64.RawStdEncoding.EncodeToString(newblock.Hash))
					if tx.Share {
						scn := BC.LocalWallets.ShareChanMap[tx.ShareChan]
						scn.TailBlockHash = newblock.Hash
						// BC.LocalWallets.ShareTailBlockHashMap[BC.GenerateUserShareKey(tx.ShareAddress)] = newblock.Hash

					} else {
						// BC.WalletsLock.Lock()
						// BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
						// BC.WalletsLock.Unlock()
						BC.LocalWallets.SetTailBlockHash(BC.GenerateAddressFromPubkey(tx.PublicKey), newblock.Hash)

					}

				}
				BC.LocalWallets.SaveToFile()
				// fmt.Println("block:", newblock.BlockId, "校验成功")
				return
			}
			fmt.Println("block:", newblock.BlockId, "校验失败")
		},
	})
}
func DelChan(creator, channame, username, user_address string) {
	tx, e := BC.NewTransaction(creator+"."+channame, []byte(username), Type.DEL_CHAN, user_address, false, "")
	if e != nil {
		quorum.Request(user_address, true, &BC.Transaction{
			Key:       creator + "." + channame,
			Share:     false,
			Value:     []byte(username),
			DataType:  Type.DEL_CHAN,
			Timestamp: uint64(time.Now().Unix()),
			ShareChan: "",
		})
		return
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
				for _, tx := range newblock.TxInfos {

					// fmt.Println(tx.Share, tx.ShareAddress, base64.RawStdEncoding.EncodeToString(newblock.Hash))
					if tx.Share {
						scn := BC.LocalWallets.ShareChanMap[tx.ShareChan]
						scn.TailBlockHash = newblock.Hash
						// BC.LocalWallets.ShareTailBlockHashMap[BC.GenerateUserShareKey(tx.ShareAddress)] = newblock.Hash

					} else {
						// BC.WalletsLock.Lock()
						// BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
						// BC.WalletsLock.Unlock()
						BC.LocalWallets.SetTailBlockHash(BC.GenerateAddressFromPubkey(tx.PublicKey), newblock.Hash)
					}

				}
				BC.LocalWallets.SaveToFile()
				// fmt.Println("block:", newblock.BlockId, "校验成功")
				return
			}
			fmt.Println("block:", newblock.BlockId, "校验失败")
		},
	})
}
func NewChan(newchan *BC.ShareChan, username string, user_address string) error {
	if BC.IsExsistChan(newchan.Channame, user_address) {
		return errors.New("已经存在的chan")
	}
	tx, e := BC.NewTransaction(username+"."+newchan.Channame, newchan.JoinKey, Type.NEW_CHAN, user_address, false, "")
	if e != nil {
		quorum.Request(user_address, true, &BC.Transaction{
			Key:       username + "." + newchan.Channame,
			Value:     newchan.JoinKey,
			Share:     false,
			DataType:  Type.NEW_CHAN,
			Timestamp: uint64(time.Now().Unix()),
			ShareChan: "",
		})
		return nil
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
				BC.LocalWallets.ShareChanMap[username+"."+newchan.Channame] = newchan
				BC.LocalWallets.SaveToFile()
				for _, tx := range newblock.TxInfos {

					// fmt.Println(tx.Share, tx.ShareAddress, base64.RawStdEncoding.EncodeToString(newblock.Hash))
					if tx.Share {
						scn := BC.LocalWallets.ShareChanMap[tx.ShareChan]
						scn.TailBlockHash = newblock.Hash
						// BC.LocalWallets.ShareTailBlockHashMap[BC.GenerateUserShareKey(tx.ShareAddress)] = newblock.Hash

					} else {
						// BC.WalletsLock.Lock()
						// BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = newblock.Hash
						// BC.WalletsLock.Unlock()
						BC.LocalWallets.SetTailBlockHash(BC.GenerateAddressFromPubkey(tx.PublicKey), newblock.Hash)

					}

				}
				BC.LocalWallets.SaveToFile()
				// fmt.Println("block:", newblock.BlockId, "校验成功")
				return
			}
			fmt.Println("block:", newblock.BlockId, "校验失败")
		},
	})
	return nil
}
func UserIsInChan(username, channame string) bool {
	flag := true
	addr, err := GetAddressFromUsername(username)
	if err != nil {
		return false
	}
	localBlockChain.Traverse(func(block *BC.Block, err error) bool {
		for _, tx := range block.TxInfos {

			//
			if tx.Key == channame {
				if tx.DataType == Type.EXIT_CHAN && BC.GenerateAddressFromPubkey(tx.PublicKey) == addr {
					flag = false
					return false
				}
				if tx.DataType == Type.JOIN_CHAN {
					return false
				}
				if tx.DataType == Type.NEW_CHAN {
					flag = false
					return false
				}

			}

		}
		return true
	})
	if !flag {
		return false
	}
	return true
}
func IsExsistChan(name string, address string) bool {
	flag := true
	localBlockChain.Traverse(func(block *BC.Block, err error) bool {
		for _, tx := range block.TxInfos {

			//
			if tx.Key == name && BC.GenerateAddressFromPubkey(tx.PublicKey) == address {
				if tx.DataType == Type.NEW_CHAN {
					flag = false
					return false
				} else if tx.DataType == Type.DEL_CHAN {
					return false
				}
			}

		}
		return true
	})
	if !flag {
		// 找到
		return true
	}
	return false
}
func GetChanUsers(channame string, username string, address string) []string {
	status := map[string]bool{}
	if !IsExsistChan(channame, address) {
		return []string{}
	}
	users := []string{}
	localBlockChain.Traverse(func(block *BC.Block, err error) bool {
		for i := len(block.TxInfos) - 1; i >= 0; i-- {
			tx := block.TxInfos[i]
			if tx.Key == channame {
				if tx.DataType == Type.NEW_CHAN && BC.GenerateAddressFromPubkey(tx.PublicKey) == address {
					users = append(users, username)
					return false
				} else if tx.DataType == Type.DEL_CHAN {
					return false
				} else if tx.DataType == Type.JOIN_CHAN {
					arr := strings.Split(string(tx.Value), " ")
					if len(arr) < 2 {
						return false
					}
					// if arr[0] == username {
					un, e := BC.GetUsernameFromAddress(BC.GenerateAddressFromPubkey(tx.PublicKey))
					if e != nil {
						continue
					}
					if un != username {
						_, ok := status[un]
						if !ok {
							users = append(users, un)
							status[un] = true
						}

					}

					// }
				} else if tx.DataType == Type.EXIT_CHAN {
					uname, e := BC.GetUsernameFromAddress(BC.GenerateAddressFromPubkey(tx.PublicKey))
					if e != nil {
						continue
					} else {
						_, ok := status[uname]
						if !ok {
							status[uname] = false
						}
					}
				}
			}
		}
		return true
	})
	return users
}
