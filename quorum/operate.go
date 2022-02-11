package quorum

import (
	"context"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	bcgrpc "go_code/基于区块链的非关系型数据库/proto"

	"google.golang.org/grpc"
)

func DistributeBlock(block *BC.Block, node *BlockChainNode, handle func(bcgrpc.VerifyInfo, error)) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", node.LocalIp, node.LocalPort), grpc.WithInsecure())
	if err != nil {
		fmt.Println("网络异常", err)
	}
	//网络延迟关闭
	defer conn.Close()

	//获得grpc句柄
	c := bcgrpc.NewBlockChainServiceClient(conn)

	//通过句柄调用函数
	fmt.Println("ffffffffff")
	re, err := c.DistributeBlock(context.Background(), CopyBlock(block))
	if err != nil {
		fmt.Println("sayhello 服务调用失败")
	}
	fmt.Println("调用sayhello的返回", re)
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

			// 当交易打包时在填上
			PreBlockHash: tx.PreBlockHash,
			Signature:    tx.Signature,
		})
	}
	return new_grpc_block
}
