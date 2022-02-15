package quorum

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	bcgrpc "go_code/基于区块链的非关系型数据库/proto"
)

type Server struct{}

func (this *Server) DistributeBlock(ctx context.Context, req *bcgrpc.Block) (info *bcgrpc.VerifyInfo, err error) {
	info.Info = "区块校验成功"
	err = nil
	rw := BC.LocalWallets.GetBlockChainRootWallet()
	flag := localBlockChain.VerifyBlock(rw.PubKey, CopyBlock2(req))
	if !flag {
		info.Info = "区块校验失败"
		info.Status = false
		return
	}
	return
}
func (this *Server) GetLatestBlock(ctx context.Context, req *bcgrpc.ReqBlock) (info *bcgrpc.Block, err error) {
	latestblock, err := localBlockChain.GetTailBlock()
	if err != nil {
		return nil, err
	}
	info = CopyBlock(latestblock)
	return
}
func (this *Server) JoinGroup(ctx context.Context, req *bcgrpc.NodeInfo) (info *bcgrpc.Nodes, err error) {
	err = JointoGroup(req.Passworld, req.LocalIp, int32(req.LocalPort))
	if err != nil {
		return
	}
	for _, node := range localNode.quorum {
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
func (this *Server) Request(ctx context.Context, req *bcgrpc.RequestBody) (info *bcgrpc.VerifyInfo, err error) {

	_, _e := BC.LocalWallets.GetUserWallet(req.UserAddress)
	if _e != nil {
		info.Status = false
		info.Info = "请求失败"
		return
	}

	tx := &BC.Transaction{
		Key:          req.Tx.Key,
		Value:        req.Tx.Value,
		DataType:     req.Tx.DataType,
		Timestamp:    req.Tx.Timestamp,
		DelMark:      req.Tx.DelMark,
		PublicKey:    req.Tx.PublicKey,
		ShareAddress: req.Tx.Shareuseraddress,
		Share:        req.Tx.Share,
	}

	info.Status = true
	info.Info = "请求成功"
	go func() {
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

				BC.LocalWallets.TailBlockHashMap[req.UserAddress] = newblock.Hash

				if req.Tx.Share {
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
	}()
	return
}
func GetUserAddressByUsername(username string) (string, error) {
	wa := BC.LocalWallets.GetBlockChainRootWallet()
	_hash, _ := BC.LocalWallets.GetUserTailBlockHash(wa.NewAddress())
	b, e := localBlockChain.GetBlockByHash(_hash)
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
		b, e = localBlockChain.GetBlockByHash(_hash)

	}
}
func (this *Server) BlockSynchronization(ctx context.Context, req *bcgrpc.ReqBlock) (out *bcgrpc.ResBlocks, err error) {
	if req.BlockId < 0 {
		return
	}
	var b *BC.Block
	var e error
	b, e = localBlockChain.GetTailBlock()
	if e != nil || b == nil {
		return
	}
	for {
		if b.BlockId > req.BlockId {
			out.Blocks = append(out.Blocks, CopyBlock(b))
		} else {
			return
		}
		if b.IsGenesisBlock() {
			return
		}
		b, e = localBlockChain.GetBlockByHash(b.PreBlockHash)
		if e != nil || b.BlockId == 0 {
			return
		}

	}
}
