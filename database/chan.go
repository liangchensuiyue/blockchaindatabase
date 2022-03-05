package database

import (
	"encoding/base64"
	"errors"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	"go_code/基于区块链的非关系型数据库/quorum"
	"go_code/基于区块链的非关系型数据库/util"
	"strings"
	"time"
)

func JoinChan(channame, username, user_address string, creator, joinkey string, key string) error {
	_joinkey, e1 := base64.RawStdEncoding.DecodeString(joinkey)

	_key, e2 := base64.RawStdEncoding.DecodeString(key)
	if e1 != nil || e2 != nil {
		return errors.New("密钥错误")
	}
	_, _ok := BC.LocalWallets.ShareChanMap[channame]
	if !_ok {
		BC.LocalWallets.ShareChanMap[channame] = &BC.ShareChan{
			Channame: channame,
			Key:      _key,
			JoinKey:  util.AesEncrypt(_joinkey, _key),
			Creator:  creator,
		}
		BC.LocalWallets.SaveToFile()
	}
	value := []byte(creator + " ")
	value = append(value, util.AesEncrypt(_joinkey, _key)...)
	tx, e := BC.NewTransaction(channame, value, BC.JOIN_CHAN, user_address, false, "")
	if e != nil {
		quorum.Request(user_address, true, &BC.Transaction{
			Key:       channame,
			Share:     false,
			Value:     value,
			DataType:  BC.JOIN_CHAN,
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
	return nil
}
func ExitChan(creator, channame, username, user_address string) {
	tx, e := BC.NewTransaction(channame, []byte(creator), BC.EXIT_CHAN, user_address, false, "")
	if e != nil {
		quorum.Request(user_address, true, &BC.Transaction{
			Key:       channame,
			Share:     false,
			Value:     []byte(creator),
			DataType:  BC.EXIT_CHAN,
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
}
func DelChan(creator, channame, username, user_address string) {
	tx, e := BC.NewTransaction(channame, []byte(creator), BC.DEL_CHAN, user_address, false, "")
	if e != nil {
		quorum.Request(user_address, true, &BC.Transaction{
			Key:       channame,
			Share:     false,
			Value:     []byte(creator),
			DataType:  BC.DEL_CHAN,
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
}
func NewChan(newchan *BC.ShareChan, username string, user_address string) error {
	tx, e := BC.NewTransaction(newchan.Channame, newchan.JoinKey, BC.NEW_CHAN, user_address, false, "")
	if e != nil {
		quorum.Request(user_address, true, &BC.Transaction{
			Key:       newchan.Channame,
			Value:     newchan.JoinKey,
			Share:     false,
			DataType:  BC.NEW_CHAN,
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
				BC.LocalWallets.ShareChanMap[newchan.Channame+"."+username] = newchan
				BC.LocalWallets.SaveToFile()
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
				if tx.DataType == BC.EXIT_CHAN && BC.GenerateAddressFromPubkey(tx.PublicKey) == addr {
					flag = false
					return false
				}
				if tx.DataType == BC.JOIN_CHAN {
					return false
				}
				if tx.DataType == BC.NEW_CHAN {
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
				if tx.DataType == BC.NEW_CHAN {
					flag = false
					return false
				} else if tx.DataType == BC.DEL_CHAN {
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
func GetChanUsers(channame string, address string) []string {
	if !IsExsistChan(channame, address) {
		return []string{}
	}
	users := []string{}
	localBlockChain.Traverse(func(block *BC.Block, err error) bool {
		for _, tx := range block.TxInfos {

			//
			if tx.Key == channame && BC.GenerateAddressFromPubkey(tx.PublicKey) == address {
				if tx.DataType == BC.NEW_CHAN {
					return false
				} else if tx.DataType == BC.DEL_CHAN {
					return false
				} else if tx.DataType == BC.JOIN_CHAN {
					arr := strings.Split(string(tx.Value), " ")
					if len(arr) < 2 {
						return false
					}
					users = append(users, arr[0])
				}
			}

		}
		return true
	})
	return users
}
