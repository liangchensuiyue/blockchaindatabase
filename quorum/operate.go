package quorum

import (
	"context"
	"errors"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	bcgrpc "go_code/基于区块链的非关系型数据库/proto"

	"google.golang.org/grpc"
)

func Broadcast() {
	for _, node := range localnode.quorum {
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", node.LocalIp, node.LocalPort), grpc.WithInsecure())
		if err != nil {
			fmt.Printf("%s:%d 网络异常", node.LocalIp, node.LocalPort)
		}
		//网络延迟关闭
		defer conn.Close()

		//获得grpc句柄
		c := bcgrpc.NewBlockChainServiceClient(conn)

		//通过句柄调用函数
		re, err := c.JoinGroup(context.Background(), &bcgrpc.NodeInfo{
			Passworld: localnode.BCInfo.PassWorld,
			LocalIp:   localnode.LocalIp,
			LocalPort: int32(localnode.LocalPort),
		})
		if err != nil {
			fmt.Println("JoinGroup 服务调用失败")
		} else {
			for _, n := range re.Nodes {
				JointoGroup(localnode.BCInfo.PassWorld, n.LocalIp, n.LocalPort)
			}
		}
	}

}

func BlockSynchronization() ([]*BC.Block, error) {
	var blockId_map map[int]uint64 = make(map[int]uint64)
	for index, rnode := range localnode.quorum {
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", rnode.LocalIp, rnode.LocalPort), grpc.WithInsecure())
		if err != nil {
			fmt.Println("网络异常", err)
		}
		//网络延迟关闭
		defer conn.Close()

		//获得grpc句柄
		c := bcgrpc.NewBlockChainServiceClient(conn)

		//通过句柄调用函数
		re, err := c.GetLatestBlock(context.Background(), &bcgrpc.ReqBlock{})
		if err != nil {
			fmt.Printf("%s:%d  BlockSynchronization 服务调用失败\n", rnode.LocalIp, rnode.LocalPort)
			continue
		}
		blockId_map[index] = re.BlockId

	}
	local_tailblock, _e := localBlockChain.GetTailBlock()
	if _e != nil {
		local_tailblock.BlockId = 0
	}
	var _index uint64 = 0
	var _id uint64 = 0
	for i, v := range blockId_map {
		if v > _id {
			_index = uint64(i)
			_id = v
		}
	}
	if _id <= local_tailblock.BlockId {
		return []*BC.Block{}, nil
	}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", localnode.quorum[_index].LocalIp, localnode.quorum[_index].LocalPort), grpc.WithInsecure())
	if err != nil {
		fmt.Println("网络异常", err)
	}
	//网络延迟关闭
	defer conn.Close()

	//获得grpc句柄
	c := bcgrpc.NewBlockChainServiceClient(conn)

	//通过句柄调用函数

	req, err := c.BlockSynchronization(context.Background(), &bcgrpc.ReqBlock{
		BlockId: local_tailblock.BlockId,
		Hash:    local_tailblock.Hash,
	})
	if err != nil {
		fmt.Printf("%s:%d  BlockSynchronization 服务调用失败\n", localnode.quorum[_index].LocalIp, localnode.quorum[_index].LocalPort)
		return nil, errors.New("区块同步失败")
	}
	Blocks := []*BC.Block{}
	if len(req.Blocks) == 0 {
		return Blocks, nil
	}
	for j := len(req.Blocks) - 1; j >= 0; j-- {
		Blocks = append(Blocks, CopyBlock2(req.Blocks[j]))
	}
	return Blocks, nil
}
func DistributeBlock(block *BC.Block, node *BlockChainNode, handle func(*bcgrpc.VerifyInfo, error)) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", node.LocalIp, node.LocalPort), grpc.WithInsecure())
	if err != nil {
		fmt.Println("网络异常", err)
	}
	//网络延迟关闭
	defer conn.Close()

	//获得grpc句柄
	c := bcgrpc.NewBlockChainServiceClient(conn)

	//通过句柄调用函数
	re, err := c.DistributeBlock(context.Background(), CopyBlock(block))
	if err != nil {
		fmt.Println("DistributeBlock 服务调用失败")
	}
	handle(re, err)
}
func CopyBlock(block *BC.Block) *bcgrpc.Block {
	new_grpc_block := &bcgrpc.Block{
		// 前区块hash
		PreBlockHash: block.PreBlockHash,
		// 块 序号
		BlockId: block.BlockId,
		// 时间戳
		Timestamp: block.Timestamp,
		// 当前区块hash,正常比特币区块中没有当前区块hash，这里是为了方便做了简化
		Hash: block.Hash,

		MerkelRoot: block.MerkelRoot,

		// 由集群私钥加密
		Signature: block.Signature,
	}
	new_grpc_block.TxInfos = []*bcgrpc.Transaction{}
	for _, tx := range block.TxInfos {
		new_grpc_block.TxInfos = append(new_grpc_block.TxInfos, &bcgrpc.Transaction{
			Key:       tx.Key,
			Value:     tx.Value,
			DataType:  tx.DataType,
			Timestamp: tx.Timestamp,
			DelMark:   tx.DelMark,
			PublicKey: tx.PublicKey,
			Hash:      tx.Hash,

			PreBlockHash: tx.PreBlockHash,
			Signature:    tx.Signature,
		})
	}
	return new_grpc_block
}
func CopyBlock2(block *bcgrpc.Block) *BC.Block {
	new_grpc_block := &BC.Block{
		// 前区块hash
		PreBlockHash: block.PreBlockHash,
		// 块 序号
		BlockId: block.BlockId,
		// 时间戳
		Timestamp: block.Timestamp,
		// 当前区块hash,正常比特币区块中没有当前区块hash，这里是为了方便做了简化
		Hash: block.Hash,

		MerkelRoot: block.MerkelRoot,

		// 由集群私钥加密
		Signature: block.Signature,
	}
	new_grpc_block.TxInfos = []*BC.Transaction{}
	for _, tx := range block.TxInfos {
		new_grpc_block.TxInfos = append(new_grpc_block.TxInfos, &BC.Transaction{
			Key:       tx.Key,
			Value:     tx.Value,
			DataType:  tx.DataType,
			Timestamp: tx.Timestamp,
			DelMark:   tx.DelMark,
			PublicKey: tx.PublicKey,
			Hash:      tx.Hash,

			// 当交易打包时在填上
			PreBlockHash: tx.PreBlockHash,
			Signature:    tx.Signature,
		})
	}
	return new_grpc_block
}
