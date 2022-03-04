package quorum

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	bcgrpc "go_code/基于区块链的非关系型数据库/proto/blockchain"
	view "go_code/基于区块链的非关系型数据库/view"

	"google.golang.org/grpc"
)

var localBlockChain *BC.BlockChain
var localNode *BlockChainNode

func Broadcast(lbc *BC.BlockChain) {
	localBlockChain = lbc
	for _, node := range localNode.Quorum {
		if node.LocalIp == localNode.LocalIp && node.LocalPort == localNode.LocalPort {
			continue
		}
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
			Passworld: localNode.BCInfo.PassWorld,
			LocalIp:   localNode.LocalIp,
			LocalPort: int32(localNode.LocalPort),
		})
		if err != nil {
			fmt.Println("JoinGroup 服务调用失败")
		} else {
			for _, n := range re.Nodes {
				JointoGroup(localNode.BCInfo.PassWorld, n.LocalIp, n.LocalPort)
			}
		}
	}

}
func GetShareChan(name string) {
	for _, node := range localNode.Quorum {
		if node.LocalIp == localNode.LocalIp && node.LocalPort == localNode.LocalPort {
			continue
		}
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", node.LocalIp, node.LocalPort), grpc.WithInsecure())
		if err != nil {
			// fmt.Printf("%s:%d 网络异常", node.LocalIp, node.LocalPort)
		}

		//获得grpc句柄
		c := bcgrpc.NewBlockChainServiceClient(conn)

		//通过句柄调用函数
		re, err := c.GetShareChan(context.Background(), &bcgrpc.ShareChanName{
			Name: name,
		})
		if err != nil {
			// fmt.Println("JoinGroup 服务调用失败")
		} else {
			BC.LocalWallets.ShareChanMap[name] = &BC.ShareChan{
				Key:       re.Key,
				ShareUser: re.Users,
			}
		}
		//网络延迟关闭
		conn.Close()
	}

}
func Request(useraddress string, strict bool, tx *BC.Transaction) error {
	for _, rnode := range localNode.Quorum {
		if rnode.LocalIp == localNode.LocalIp && rnode.LocalPort == localNode.LocalPort {
			continue
		}
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", rnode.LocalIp, rnode.LocalPort), grpc.WithInsecure())
		if err != nil {
			fmt.Println("网络异常", err)
		}
		//网络延迟关闭
		defer conn.Close()

		//获得grpc句柄
		c := bcgrpc.NewBlockChainServiceClient(conn)

		//通过句柄调用函数
		re, err := c.Request(context.Background(), &bcgrpc.RequestBody{
			UserAddress: useraddress,
			Strict:      strict,
			Tx: &bcgrpc.Transaction{
				Key:       tx.Key,
				Value:     tx.Value, // []byte
				DataType:  tx.DataType,
				Timestamp: tx.Timestamp, // 时间错
				ShareChan: tx.ShareChan,
				Share:     tx.Share,
			},
		})
		if err != nil {
			fmt.Printf("%s:%d  Request 服务调用失败\n", rnode.LocalIp, rnode.LocalPort)
			continue
		}
		if re.Status {
			return nil
		}

	}
	return errors.New("请求失败")
}

func BlockSynchronization() ([]*BC.Block, error) {
	var blockId_map map[int]uint64 = make(map[int]uint64)
	if len(localNode.Quorum) == 0 {
		return []*BC.Block{}, nil
	}
	for index, rnode := range localNode.Quorum {
		if rnode.LocalIp == localNode.LocalIp && rnode.LocalPort == localNode.LocalPort {
			continue
		}
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
		local_tailblock = &BC.Block{}
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
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", localNode.Quorum[_index].LocalIp, localNode.Quorum[_index].LocalPort), grpc.WithInsecure())
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
		fmt.Printf("%s:%d  BlockSynchronization 服务调用失败\n", localNode.Quorum[_index].LocalIp, localNode.Quorum[_index].LocalPort)
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
	if err != nil || !re.Status {
		// fmt.Println("DistributeBlock 服务调用失败")
	}

	infos := []map[string]interface{}{}
	for i, tx := range block.TxInfos {

		// fmt.Println("交易索引:", i)
		// fmt.Println("user_address:", BC.GenerateAddressFromPubkey(tx.PublicKey))
		// fmt.Println("key-value:", tx.Key, string(tx.Value))
		// fmt.Println("sharemode:", tx.Share)
		// fmt.Println("delmark:", tx.DelMark)
		// fmt.Println("sharechan:", tx.ShareChan)

		infos = append(infos, map[string]interface{}{
			"Index":         i,
			"UserAddress":   BC.GenerateAddressFromPubkey(tx.PublicKey),
			"Hash":          base64.RawStdEncoding.EncodeToString(tx.Hash),
			"Timestamp":     tx.Timestamp,
			"Sharechan":     tx.ShareChan,
			"PrevBlockHash": base64.RawStdEncoding.EncodeToString(tx.PreBlockHash),
		})
	}
	datastr, _ := json.Marshal(map[string]interface{}{
		"BlockId":       block.BlockId,
		"PrevBlockHash": base64.RawStdEncoding.EncodeToString(block.PreBlockHash),
		"Hash":          base64.RawStdEncoding.EncodeToString(block.Hash),
		"Timestamp":     block.Timestamp,
		"TxInfos":       infos,
	})
	view.MsgQueue <- view.Message{
		Type:       "DistributeBlock",
		MsgJsonStr: string(datastr),
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
			Key:          tx.Key,
			Value:        tx.Value,
			DataType:     tx.DataType,
			Timestamp:    tx.Timestamp,
			PublicKey:    tx.PublicKey,
			Hash:         tx.Hash,
			Share:        tx.Share,
			ShareChan:    tx.ShareChan,
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
			PublicKey: tx.PublicKey,
			Hash:      tx.Hash,
			Share:     tx.Share,
			ShareChan: tx.ShareChan,
			// 当交易打包时在填上
			PreBlockHash: tx.PreBlockHash,
			Signature:    tx.Signature,
		})
	}
	return new_grpc_block
}
