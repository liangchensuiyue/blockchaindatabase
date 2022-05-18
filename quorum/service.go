package quorum

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	bcgrpc "go_code/基于区块链的非关系型数据库/proto/blockchain"
	Type "go_code/基于区块链的非关系型数据库/type"
	util "go_code/基于区块链的非关系型数据库/util"
	"time"
)

type Server struct{}

func (this *Server) GetAccountant(ctx context.Context, req *bcgrpc.Heartbeat) (info *bcgrpc.VerifyInfo, err error) {
	info = &bcgrpc.VerifyInfo{}
	info.Status = false
	if int(req.BlockNums) > BC.BlockQueue.Len() && isAccountant {
		isAccountant = false
		info.Status = true
		return info, nil
	} else {
		return info, errors.New("获取记账权失败")
	}
}
func (this *Server) QuorumHeartbeat(ctx context.Context, req *bcgrpc.NodeInfo) (info *bcgrpc.Heartbeat, err error) {
	info = &bcgrpc.Heartbeat{}
	// fmt.Println("心跳检测", req.Passworld == LocalNode.BCInfo.PassWorld)
	if req.Passworld != LocalNode.BCInfo.PassWorld {
		return info, errors.New("没有访问权限")
	}
	info.IsAccountant = isAccountant
	info.BlockNums = int32(BC.BlockQueue.Len())

	return info, nil
}
func (this *Server) DistributeBlock(ctx context.Context, req *bcgrpc.Block) (info *bcgrpc.VerifyInfo, err error) {
	info = &bcgrpc.VerifyInfo{}
	info.Info = "区块校验成功"
	info.Status = true
	err = nil
	rw := BC.LocalWallets.GetBlockChainRootWallet()
	newblock := CopyBlock2(req)
	flag := LocalBlockChain.VerifyBlock(rw.PubKey, newblock)
	if !flag {
		info.Info = "区块校验失败"
		info.Status = false
		return
	}
	e := LocalBlockChain.AddBlock(newblock)
	if e != nil {
		fmt.Println(e)
		return
	}
	if newblock.IsGenesisBlock() {
		// BC.WalletsLock.Lock()
		// BC.LocalWallets.TailBlockHashMap[rw.NewAddress()] = newblock.Hash
		// BC.WalletsLock.Unlock()
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
	return
}

func (this *Server) GetLatestBlock(ctx context.Context, req *bcgrpc.ReqBlock) (info *bcgrpc.Block, err error) {
	latestblock, err := LocalBlockChain.GetTailBlock()
	if err != nil {
		return nil, err
	}
	info = CopyBlock(latestblock)
	return
}
func (this *Server) GetShareChan(ctx context.Context, req *bcgrpc.ShareChanName) (body *bcgrpc.ShareChanBody, err error) {
	body = &bcgrpc.ShareChanBody{}
	cn, ok := BC.LocalWallets.ShareChanMap[req.Name]
	if !ok {
		return nil, errors.New("not found")
	}
	body.Key = cn.Key
	body.JoinKey = cn.JoinKey
	body.Creator = cn.Creator
	body.CreatorAddress = cn.CreatorAddress
	body.Channame = cn.Channame
	return
}
func (this *Server) JoinGroup(ctx context.Context, req *bcgrpc.NodeInfo) (info *bcgrpc.Nodes, err error) {
	info = &bcgrpc.Nodes{}
	err = JointoGroup(req.Passworld, req.LocalIp, int32(req.LocalPort))

	_nodes := []string{}
	for _, v := range LocalNode.Quorum {
		_nodes = append(_nodes, v.LocalIp)
	}
	if err != nil {
		return
	}
	for _, node := range LocalNode.Quorum {
		info.Nodes = append(info.Nodes, &bcgrpc.NodeInfo{
			LocalIp:   node.LocalIp,
			LocalPort: int32(node.LocalPort),
		})
	}
	return
}
func VeriftUser(username string, passworld string) error {
	user_address := BC.LocalWallets.GetBlockChainRootWallet().NewAddress()

	// 判断用户是否创建
	_hash, _ := BC.LocalWallets.GetUserTailBlockHash(user_address)

	b, e := LocalBlockChain.GetBlockByHash(_hash)
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
		b, _ = LocalBlockChain.GetBlockByHash(_hash)
	}
	return errors.New("未知的用户")
}
func (this *Server) Request(ctx context.Context, req *bcgrpc.RequestBody) (info *bcgrpc.VerifyInfo, err error) {
	info = &bcgrpc.VerifyInfo{}
	uw, _e := BC.LocalWallets.GetUserWallet(req.UserAddress)
	if _e != nil {
		info.Status = false
		info.Info = "请求失败"
		return
	}

	info.Status = true
	info.Info = "请求成功"
	user_name, err := BC.GetUsernameFromAddress(BC.GenerateAddressFromPubkey(uw.PubKey))
	if err != nil {
		return
	}
	newchan := &BC.ShareChan{
		Channame: req.Tx.Key,
	}
	newchan.YieldKey()
	newchan.Creator = user_name
	newchan.CreatorAddress = BC.GenerateAddressFromPubkey(uw.PubKey)

	newchan.JoinKey = util.AesEncrypt([]byte(base64.RawStdEncoding.EncodeToString([]byte(fmt.Sprintf("%d%s", time.Now().UnixNano(), newchan.Channame)))), newchan.Key)
	tx := &BC.Transaction{
		Key:       req.Tx.Key,
		Value:     req.Tx.Value,
		DataType:  req.Tx.DataType,
		Timestamp: req.Tx.Timestamp,
		PublicKey: uw.PubKey,
		ShareChan: req.Tx.ShareChan,
		Share:     req.Tx.Share,
	}
	// if !tx.VerifySimple() {
	// 	return
	// }
	if req.Tx.DataType == Type.NEW_CHAN {
		ok := BC.UserIsChanCreator(newchan.Channame, BC.GenerateAddressFromPubkey(uw.PubKey))
		if ok {
			// fmt.Println("改chan已存在")
			return
		}
		tx = &BC.Transaction{
			Key:       req.Tx.Key,
			Value:     newchan.JoinKey,
			DataType:  req.Tx.DataType,
			Timestamp: req.Tx.Timestamp,
			PublicKey: uw.PubKey,
			ShareChan: req.Tx.ShareChan,
			Share:     req.Tx.Share,
		}
	}
	if req.Strict {
		lcdraft := BC.GetLocalDraft()
		newblock, _ := lcdraft.PackBlock(tx)
		rw := BC.LocalWallets.GetBlockChainRootWallet()
		// LocalBlockChain.SignBlock(rw.Private, false, newblock)
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
					if tx.DataType == Type.NEW_CHAN {

						BC.LocalWallets.ShareChanMap[newchan.Creator+"."+newchan.Channame] = newchan
						BC.LocalWallets.SaveToFile()
					}
					// BC.LocalWallets.TailBlockHashMap[req.UserAddress] = newblock.Hash

					// if req.Tx.Share {
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
					// }
					BC.LocalWallets.SaveToFile()
					// fmt.Println("block:", newblock.BlockId, "校验成功")
					return
				}
				// fmt.Println("block:", newblock.BlockId, "校验失败")
			},
		})
	} else {
		draft := BC.GetLocalDraft()
		draft.PutTx(tx)
	}

	return
}
func GetUserAddressByUsername(username string) (string, error) {
	wa := BC.LocalWallets.GetBlockChainRootWallet()
	_hash, _ := BC.LocalWallets.GetUserTailBlockHash(wa.NewAddress())
	b, e := LocalBlockChain.GetBlockByHash(_hash)
	for {
		if e != nil {
			return "", e
		}
		for _, tx := range b.TxInfos {
			_hash = tx.PreBlockHash
			if tx.Key == username {
				return string(BC.GenerateAddressFromPubkey(tx.PublicKey)), nil
			}
		}
		b, e = LocalBlockChain.GetBlockByHash(_hash)

	}
}
func (this *Server) BlockSynchronization(ctx context.Context, req *bcgrpc.ReqBlock) (out *bcgrpc.ResBlocks, err error) {
	// fmt.Println("同步服务调用 blockid", req.BlockId)
	out = &bcgrpc.ResBlocks{}
	if req.BlockId < 0 {
		return
	}
	var b *BC.Block
	var e error
	b, e = LocalBlockChain.GetTailBlock()
	if e != nil || b == nil {
		return
	}
	for {
		if b.BlockId > req.BlockId {
			out.Blocks = append(out.Blocks, CopyBlock(b))
		} else {
			// fmt.Println("同步区块数量:", len(out.Blocks))
			return
		}
		if b.IsGenesisBlock() {
			// fmt.Println("同步区块数量:", len(out.Blocks))
			return
		}
		b, e = LocalBlockChain.GetBlockByHash(b.PreBlockHash)
		if e != nil || b.BlockId == 0 {
			// fmt.Println("同步区块数量:", len(out.Blocks))
			return
		}

	}
}
