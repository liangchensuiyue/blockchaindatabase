package database

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
)

func CreateUser(username string, passworld string) error {
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
	tx, e := BC.NewTransaction("create_user", username, []byte(base64.RawStdEncoding.EncodeToString([]byte(passworld))), "string", user_address, false, []string{})
	if e != nil {
		return errors.New("创建用户失败")
	}

	lcdraft := BC.GetLocalDraft()
	newblock, _ := lcdraft.PackBlock(tx)
	rw := BC.LocalWallets.GetBlockChainRootWallet()
	localBlockChain.SignBlock(rw.Private, false, newblock)

	localNode.DistribuBlock(newblock, func(total, fail int) {
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
	})
	_, e = GetAddressFromUsername(username)
	if e != nil {
		return errors.New("创建失败")
	}
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
				if bytes.Equal(tx.Value, []byte(base64.RawStdEncoding.EncodeToString([]byte(passworld)))) {
					return nil
				}
				return errors.New("密码错误")
			}
		}
		b, _ = localBlockChain.GetBlockByHash(_hash)
	}
	return errors.New("未知的用户")
}

func Put(key string, value []byte, datatype string, user_address string, share bool, shareuser []string, strict bool) error {
	if share {
		for i := 0; i < len(shareuser); i++ {
			addr, e := GetAddressFromUsername(shareuser[i])
			if e != nil {
				return e
			}
			shareuser[i] = addr
		}
	} else {
		shareuser = []string{}
	}
	tx, e := BC.NewTransaction("put", key, value, datatype, user_address, share, shareuser)
	if e != nil {
		return e
	}
	if strict {
		lcdraft := BC.GetLocalDraft()
		newblock, _ := lcdraft.PackBlock(tx)
		rw := BC.LocalWallets.GetBlockChainRootWallet()
		localBlockChain.SignBlock(rw.Private, false, newblock)

		localNode.DistribuBlock(newblock, func(total, fail int) {
			flag := localBlockChain.VerifyBlock(rw.PubKey, newblock)
			if flag {
				e := localBlockChain.AddBlock(newblock)
				if e != nil {
					fmt.Println(e)
					return
				}

				BC.LocalWallets.TailBlockHashMap[user_address] = newblock.Hash

				if share {
					for _, tx := range newblock.TxInfos {

						for _, addr := range tx.ShareAddress {
							BC.LocalWallets.TailBlockHashMap[addr] = newblock.Hash
						}

					}
				}
				BC.LocalWallets.SaveToFile()
				fmt.Println("校验成功")
				return
			}
			fmt.Println("区块校验失败")
		})

	}
	draft := BC.GetLocalDraft()
	draft.PutTx(tx)
	return nil
}
func Del(key string, user_address string, share bool, shareuser []string, strict bool) {
	if share {
		for i := 0; i < len(shareuser); i++ {
			addr, e := GetAddressFromUsername(shareuser[i])
			if e != nil {
				return
			}
			shareuser[i] = addr
		}
	} else {
		shareuser = []string{}
	}
	tx, e := BC.NewTransaction("del", key, []byte{}, "", user_address, share, shareuser)
	if e != nil {
		return
	}
	if strict {
		lcdraft := BC.GetLocalDraft()
		newblock, _ := lcdraft.PackBlock(tx)
		rw := BC.LocalWallets.GetBlockChainRootWallet()

		localBlockChain.SignBlock(rw.Private, false, newblock)
		localNode.DistribuBlock(newblock, func(total, fail int) {
			flag := localBlockChain.VerifyBlock(rw.PubKey, newblock)
			if flag {
				localBlockChain.AddBlock(newblock)

				BC.LocalWallets.TailBlockHashMap[user_address] = newblock.Hash

				if share {
					for _, tx := range newblock.TxInfos {

						for _, addr := range tx.ShareAddress {
							BC.LocalWallets.TailBlockHashMap[addr] = newblock.Hash
						}

					}
				}
				BC.LocalWallets.SaveToFile()
				fmt.Println("区块校验成功")
				return
			}
			fmt.Println("database.operate 121 区块校验失败")
			return

		})

	}
	draft := BC.GetLocalDraft()
	draft.PutTx(tx)
	return
}
func Get(key string, user_address string, sharemode bool, shareuser []string) (*BC.Block, int) {
	user_wa, _ := BC.LocalWallets.GetUserWallet(user_address)
	tailhash, ok := BC.LocalWallets.TailBlockHashMap[user_address]
	if !ok {
		return nil, -1
	}
	user_addrs_map := map[string]bool{}
	for j := 0; j < len(shareuser); j++ {
		addr, e := GetAddressFromUsername(shareuser[j])
		if e == nil {
			user_addrs_map[addr] = true
		}
	}
	for {
		b, e := localBlockChain.GetBlockByHash(tailhash)
		if e != nil {
			fmt.Println(e, tailhash)
			return nil, -1
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
							return b, i
						} else {
							return nil, -1
						}
					}
					tailhash = b.TxInfos[i].PreBlockHash
				}
			} else {
				if bytes.Equal(b.TxInfos[i].PublicKey, user_wa.PubKey) {
					if b.TxInfos[i].Key == key {
						if !b.TxInfos[i].DelMark {
							return b, i
						} else {
							return nil, -1
						}
					}

					tailhash = b.TxInfos[i].PreBlockHash
				}
				continue
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
			_hash = tx.PreBlockHash
			if tx.Key == username {
				return BC.GenerateAddressFromPubkey(tx.PublicKey), nil
			}
		}
		b, _ = localBlockChain.GetBlockByHash(_hash)
	}
	return "", errors.New("未知的用户")
}
