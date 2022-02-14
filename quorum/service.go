package quorum

import (
	"context"
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
