package database

import (
	"bytes"
	"errors"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
)

func CreateUser(username string, passworld string) {
	wa := BC.NewWallet(username, passworld)

	BC.LocalWallets.WalletsMap[wa.NewAddress()] = wa
	BC.LocalWallets.SaveToFile()
}
func VeriftUser(username string, passworld string) (string, error) {
	for _, wa := range BC.LocalWallets.WalletsMap {
		if wa.Username == username && wa.Passworld == passworld {
			return wa.NewAddress(), nil
		}
	}
	return "", errors.New("未知的用户")
}
func Put(key string, value []byte, datatype string, user_address string, share bool, shareuser []string, strict bool) error {
	if share {
		for i := 0; i < len(shareuser); i++ {
			addr, e := BC.LocalWallets.GetAddressFromUsername(shareuser[i])
			if e != nil {
				return e
			}
			shareuser[i] = addr
		}
	} else {
		shareuser = []string{}
	}
	tx, e := BC.NewTransaction("put", key, value, datatype, user_address, shareuser)
	if e != nil {
		return e
	}
	if strict {
		lcdraft := BC.GetLocalDraft()
		newblock, _ := lcdraft.PackBlock(tx)
		flag := localBlockChain.VerifyBlock(localNode.BCInfo.PubKey, newblock)
		if flag {
			localBlockChain.AddBlock(newblock)
			for _, tx := range newblock.TxInfos {
				for _, addr := range tx.ShareAddress {
					wa, e := BC.LocalWallets.GetUserWallet(addr)
					if e != nil {
						wa.TailBlockHash = newblock.Hash
					}
				}

			}
			return nil
		} else {
			return errors.New("区块校验失败")
		}
	}
	draft := BC.GetLocalDraft()
	draft.PutTx(tx)
	return nil
}
func Del(key string, user_address string, share bool, shareuser []string, strict bool) {
	if share {
		for i := 0; i < len(shareuser); i++ {
			addr, e := BC.LocalWallets.GetAddressFromUsername(shareuser[i])
			if e != nil {
				return
			}
			shareuser[i] = addr
		}
	} else {
		shareuser = []string{}
	}
	tx, e := BC.NewTransaction("del", key, []byte{}, "", user_address, shareuser)
	if e != nil {
		return
	}
	if strict {
		lcdraft := BC.GetLocalDraft()
		newblock, _ := lcdraft.PackBlock(tx)
		flag := localBlockChain.VerifyBlock(localNode.BCInfo.PubKey, newblock)
		if flag {
			localBlockChain.AddBlock(newblock)
			for _, tx := range newblock.TxInfos {
				for _, addr := range tx.ShareAddress {
					wa, e := BC.LocalWallets.GetUserWallet(addr)
					if e != nil {
						wa.TailBlockHash = newblock.Hash
					}
				}

			}
			return
		} else {
			fmt.Println("database.operate 97 区块校验失败")
			return
		}
	}
	draft := BC.GetLocalDraft()
	draft.PutTx(tx)
	return
}
func Get(key string, value []byte, datatype string, user_address string, sharemode bool, shareuser []string) []byte {
	user_wa, _ := BC.LocalWallets.GetUserWallet(user_address)
	tailhash := BC.LocalWallets.WalletsMap[user_address].TailBlockHash
	user_addrs_map := map[string]bool{}
	for j := 0; j < len(shareuser); j++ {
		addr, e := BC.LocalWallets.GetAddressFromUsername(shareuser[j])
		if e == nil {
			user_addrs_map[addr] = true
		}
	}
	for {
		b, e := localBlockChain.GetBlockByHash(tailhash)
		if e != nil {
			return []byte{}
		}
		for i := len(b.TxInfos) - 1; i >= 0; i-- {
			if sharemode {
				if bytes.Equal(b.TxInfos[i].PublicKey, user_wa.PubKey) {
					if b.TxInfos[i].Key == key {

						flag := false
						for _, _addr := range b.TxInfos[i].ShareAddress {
							_, ok := user_addrs_map[_addr]
							if !ok {
								flag = true
							}
						}
						if flag {
							continue
						}

						if !b.TxInfos[i].DelMark {
							return b.TxInfos[i].Value
						} else {
							return []byte{}
						}
					}
					tailhash = b.TxInfos[i].PreBlockHash
				}
			} else {
				if bytes.Equal(b.TxInfos[i].PublicKey, user_wa.PubKey) {
					if b.TxInfos[i].Key == key {
						if !b.TxInfos[i].DelMark {
							return b.TxInfos[i].Value
						} else {
							return []byte{}
						}
					}

					tailhash = b.TxInfos[i].PreBlockHash
				}
				continue
			}
		}
	}
}
