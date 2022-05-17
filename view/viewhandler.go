package view

import (
	"encoding/base64"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	db "go_code/基于区块链的非关系型数据库/database"
	quorum "go_code/基于区块链的非关系型数据库/quorum"
	Type "go_code/基于区块链的非关系型数据库/type"
	"strings"
)

// "bufio"
// "os"
func GetPipeRelateBlock(pname string) []BlockInfo {
	blocks := []BlockInfo{}

	schn := BC.LocalWallets.ShareChanMap[pname]

	if !BC.LocalWallets.HasShareChan(pname) {
		return blocks
	}
	tailhash := schn.TailBlockHash

	for {
		b, e := LBC.GetBlockByHash(tailhash)
		if e != nil || b.IsGenesisBlock() {
			// fmt.Println(e, tailhash)
			return blocks
		}
		blocks = append(blocks, BlockInfo{
			BlockId:   b.BlockId,
			Hash:      base64.RawStdEncoding.EncodeToString(b.Hash),
			PrevHash:  base64.RawStdEncoding.EncodeToString(b.PreBlockHash),
			Timestamp: b.Timestamp,
			Txnums:    uint64(len(b.TxInfos)),
		})
		for i := len(b.TxInfos) - 1; i >= 0; i-- {
			if b.TxInfos[i].Share && b.TxInfos[i].ShareChan == pname {
				tailhash = b.TxInfos[i].PreBlockHash
			}

		}
	}
}
func GetUserRelateBlock(useraddress string) []BlockInfo {
	tailhash, _ := BC.LocalWallets.GetUserTailBlockHash(useraddress)
	blocks := []BlockInfo{}
	for {
		b, e := LBC.GetBlockByHash(tailhash)

		if e != nil || b.IsGenesisBlock() {
			// fmt.Println(e, tailhash)
			return blocks
		}
		blocks = append(blocks, BlockInfo{
			BlockId:   b.BlockId,
			Hash:      base64.RawStdEncoding.EncodeToString(b.Hash),
			PrevHash:  base64.RawStdEncoding.EncodeToString(b.PreBlockHash),
			Timestamp: b.Timestamp,
			Txnums:    uint64(len(b.TxInfos)),
		})
		for i := len(b.TxInfos) - 1; i >= 0; i-- {
			if BC.GenerateAddressFromPubkey(b.TxInfos[i].PublicKey) == useraddress {
				tailhash = b.TxInfos[i].PreBlockHash
			}
		}
	}
}

// type ViewHander struct {
// 	GetQuorum    func() []Quorum
// 	GetNodeInfo  func() NodeInfo
// 	GetPipeInfo  func(string) PipeInfo
// 	GetUserInfo  func(string) UserInfo
// 	GetBlockInfo func(uint64) BlockInfo
// }
func GetPipeUserinfo(pname, creaddress string) []UserInfo {
	userinfos := []UserInfo{}
	if !BC.IsExsistChan(pname, creaddress) {
		return userinfos
	}
	LBC.Traverse(func(block *BC.Block, err error) bool {
		if block != nil {
			for i := len(block.TxInfos) - 1; i >= 0; i-- {
				tx := block.TxInfos[i]
				if tx.DataType == Type.DEL_CHAN && tx.Key == pname {
					return false
				}
				if tx.DataType == Type.NEW_CHAN && tx.Key == pname {
					useradd := BC.GenerateAddressFromPubkey(tx.PublicKey)
					username, _ := BC.GetUsernameFromAddress(useradd)
					userinfos = append(userinfos, GlobalViewHandler.GetUserInfo(username))
					continue
				}
				if tx.DataType == Type.JOIN_CHAN && tx.Key == pname {
					useradd := BC.GenerateAddressFromPubkey(tx.PublicKey)
					username, _ := BC.GetUsernameFromAddress(useradd)
					userinfos = append(userinfos, GlobalViewHandler.GetUserInfo(username))
				}
			}

		}
		return true

	})
	return userinfos
}
func init() {
	GlobalViewHandler.GetUserInfo = func(username string) UserInfo {
		if !db.IsExsistUser(username) {
			return UserInfo{}
		}
		userinfo := UserInfo{}
		uaddress, err := db.GetAddressFromUsername(username)
		if err != nil {
			return UserInfo{}
		}
		userinfo.UserAddress = uaddress
		userinfo.Username = username
		Bks := GetUserRelateBlock(uaddress)
		userinfo.UserRelateBlock = Bks
		userinfo.UserTxNums = 0
		for i := 0; i < len(Bks); i++ {
			userinfo.UserTxNums += Bks[i].Txnums
		}

		return userinfo
	}
	GlobalViewHandler.GetNodeInfo = func() NodeInfo {
		nodeinfo := NodeInfo{}
		nodeinfo.Ip = LocalNodes.LocalIp
		b, _ := LBC.GetTailBlock()
		nodeinfo.LatestblockId = b.BlockId
		nodeinfo.Isaccountant = quorum.LocalNodeIsAccount()
		nodeinfo.SystxRateInfo = fmt.Sprintf("%d 笔交易耗时 %d (ms) \n 待同步的区块数量 %d", db.NUM, db.Total/1000000, BC.BlockQueue.Len())
		nodeinfo.SysblockRateInfo = fmt.Sprintf("%d 个区块同步耗时 %d (ms)", quorum.NUM, quorum.Total/1000000)
		nodeinfo.LocalWalletAddress = []string{}
		for addr, _ := range BC.LocalWallets.WalletsMap {
			nodeinfo.LocalWalletAddress = append(nodeinfo.LocalWalletAddress, addr)
		}

		return nodeinfo
	}

	GlobalViewHandler.GetPipeInfo = func(pname string) PipeInfo {
		pipeinfo := PipeInfo{}
		arr := strings.Split(pname, ".")
		if len(arr) < 2 {
			return pipeinfo
		}
		creator := arr[0]

		craddress, err := db.GetAddressFromUsername(creator)
		if err != nil {
			return pipeinfo
		}
		if !db.IsExsistChan(pname, craddress) {
			return pipeinfo
		}
		pipeinfo.Pipename = pname
		pipeinfo.PipeUser = GetPipeUserinfo(pname, craddress)
		pipeinfo.PipeRelateBlock = GetPipeRelateBlock(pname)
		pipeinfo.PipeTxNums = 0
		for i := 0; i < len(pipeinfo.PipeRelateBlock); i++ {
			pipeinfo.PipeTxNums += pipeinfo.PipeRelateBlock[i].Txnums
		}
		return pipeinfo
	}
	GlobalViewHandler.GetQuorum = func() []Quorum {
		qnode := []Quorum{}
		for _, v := range LocalNodes.Quorum {
			qnode = append(qnode, Quorum{
				Ip: v.LocalIp,
			})
		}
		return qnode
	}
	GlobalViewHandler.GetBlockInfo = func(bid uint64) BlockInfo {
		binfo := BlockInfo{}
		LBC.Traverse(func(block *BC.Block, err error) bool {
			if err != nil {
				return true
			}
			if block.BlockId == bid {
				binfo.BlockId = bid
				binfo.Hash = base64.RawStdEncoding.EncodeToString(block.Hash)
				binfo.PrevHash = base64.RawStdEncoding.EncodeToString(block.PreBlockHash)
				binfo.Timestamp = block.Timestamp
				binfo.Txnums = uint64(len(block.TxInfos))
			}
			return true
		})
		return binfo
	}
	GlobalViewHandler.GetAllBlock = func() []BlockInfo {
		blocks := []BlockInfo{}
		LBC.Traverse(func(b *BC.Block, err error) bool {
			if err != nil {
				return true
			}
			blocks = append(blocks, BlockInfo{
				BlockId:   b.BlockId,
				Hash:      base64.RawStdEncoding.EncodeToString(b.Hash),
				PrevHash:  base64.RawStdEncoding.EncodeToString(b.PreBlockHash),
				Timestamp: b.Timestamp,
				Txnums:    uint64(len(b.TxInfos)),
			})

			return true
		})
		return blocks
	}
}
