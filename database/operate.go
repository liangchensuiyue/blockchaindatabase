package database

import (
	"errors"
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
		for i := 0; i < len(shareuser)-1; i++ {
			addr, e := BC.LocalWallets.GetAddressFromUsername(shareuser[i])
			if e != nil {
				return e
			}
			shareuser[i] = addr
		}
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
		} else {
			return errors.New("区块校验失败")
		}
	}
	draft := BC.GetLocalDraft()
	draft.PutTx(tx)
	return nil
}
