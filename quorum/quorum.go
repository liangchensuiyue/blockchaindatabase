package quorum

import (
	"context"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	bcgrpc "go_code/基于区块链的非关系型数据库/proto/blockchain"
	"net"
	"time"

	"google.golang.org/grpc"
)

func _startServer() {
	//创建网络
	// fmt.Println("网络")
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", LocalNode.LocalPort))
	if err != nil {
		fmt.Println("网络错误", err)
	}

	//创建grpc的服务
	srv := grpc.NewServer()

	//注册服务
	bcgrpc.RegisterBlockChainServiceServer(srv, &Server{})

	//等待网络连接
	err = srv.Serve(ln)
	if err != nil {
		fmt.Println("网络错误", err)
	}
}
func _startHeartbeat() {
	// return
	for {
		time.Sleep(time.Second * 2)

		nodes := []*bcgrpc.Heartbeat{}
		flag := false
		for _, node := range LocalNode.Quorum {
			if node.LocalIp == LocalNode.LocalIp {
				continue
			}
			conn, err := grpc.Dial(fmt.Sprintf("%s:%d", node.LocalIp, node.LocalPort), grpc.WithInsecure())
			if err != nil {
				// fmt.Println("网络异常", err)
			}

			//获得grpc句柄
			c := bcgrpc.NewBlockChainServiceClient(conn)

			//通过句柄调用函数
			re, err := c.QuorumHeartbeat(context.Background(), &bcgrpc.NodeInfo{
				Passworld: LocalNode.BCInfo.PassWorld,
			})
			if err != nil {
				// fmt.Println("DistributeBlock 服务调用失败", err.Error())
			} else {
				if re.IsAccountant {
					flag = true
				}
				nodes = append(nodes, re)
			}
			//网络延迟关闭
			conn.Close()
		}

		if !flag && !isAccountant {
			// 重新选择会计
			nodes = append(nodes, &bcgrpc.Heartbeat{
				LocalIp:   LocalNode.LocalIp,
				LocalPort: int32(LocalNode.LocalPort),
				BlockNums: int32(BC.BlockQueue.Len()),
			})
			winner := nodes[0]
			for i := 1; i < len(nodes); i++ {
				if nodes[i].BlockNums > winner.BlockNums {
					winner = nodes[i]
					continue
				}
				if nodes[i].BlockNums == winner.BlockNums {
					if nodes[i].LocalIp > winner.LocalIp {
						winner = nodes[i]
					}
				}
			}

			if winner.LocalIp == LocalNode.LocalIp {
				isAccountant = true
			}
		}
	}
}
func getAccountant() bool {
	for _, node := range LocalNode.Quorum {
		if node.LocalIp == LocalNode.LocalIp {
			continue
		}
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", node.LocalIp, node.LocalPort), grpc.WithInsecure())
		if err != nil {
			// fmt.Println("网络异常", err)
		}

		//获得grpc句柄
		c := bcgrpc.NewBlockChainServiceClient(conn)

		//通过句柄调用函数
		re, err := c.GetAccountant(context.Background(), &bcgrpc.Heartbeat{
			BlockNums: int32(BC.BlockQueue.Len()),
		})
		if err == nil && re.Status {
			return true
		}
		//网络延迟关闭
		conn.Close()
	}
	return false
}

var NUM = 0
var Total int64 = 0
var pre int64

func _starDistributeBlock() {
	for {
		if BC.BlockQueue.Len() > 0 && !isAccountant {
			// time.Sleep(time.Second)
			flag := getAccountant()
			if flag {
				isAccountant = true
			} else {
				continue
			}
		}
		if !isAccountant {
			time.Sleep(time.Second)
			continue
		}
		if BC.BlockQueue.Len() == 0 {
			time.Sleep(time.Second)
			continue
		}
		el, _ := BC.BlockQueue.Front()
		// BC.BlockQueue.SaveToDisk()
		block := el.TargetBlock
		// if len(block.TxInfos) > 0 {
		// 	fmt.Println("打包", block.TxInfos[0].Key)

		// }
		pre = time.Now().UnixNano()
		NUM++

		total := 0
		fail := 0
		rw := BC.LocalWallets.GetBlockChainRootWallet()
		if block.IsGenesisBlock() {
			LocalBlockChain.SignBlock(rw.Private, true, block)

		} else {
			LocalBlockChain.SignBlock(rw.Private, false, block)

		}
		// fmt.Printf("区块 %d 分发\n", block.BlockId)
		for _, blockBlockChainNode := range LocalNode.Quorum {
			if blockBlockChainNode.LocalIp == LocalNode.LocalIp {
				continue
			}
			total++
			DistributeBlock(block, blockBlockChainNode, func(res *bcgrpc.VerifyInfo, err error) {
				if err != nil {
					// fmt.Println(err)
					fail++
					// fmt.Printf("节点 %s:%d 接受失败\n", blockBlockChainNode.LocalIp, blockBlockChainNode.LocalPort)
					return
				}
				if !res.Status {
					fail++
					fmt.Printf("节点 %s:%d 校验失败\n", blockBlockChainNode.LocalIp, blockBlockChainNode.LocalPort)
					return
				}
				// fmt.Printf("节点 %s:%d 接受成功\n", blockBlockChainNode.LocalIp, blockBlockChainNode.LocalPort)
			})
		}
		Total += time.Now().UnixNano() - pre
		if el.Handle == nil {
			rw := BC.LocalWallets.GetBlockChainRootWallet()
			flag := LocalBlockChain.VerifyBlock(rw.PubKey, block)
			if flag {
				e := LocalBlockChain.AddBlock(block)
				if e != nil {
					// fmt.Println(e)
					return
				}
				for _, tx := range block.TxInfos {
					// fmt.Println(tx.Share, tx.ShareAddress, base64.RawStdEncoding.EncodeToString(newblock.Hash))
					if tx.Share {
						schn := BC.LocalWallets.ShareChanMap[tx.ShareChan]
						schn.TailBlockHash = block.Hash

					} else {
						// BC.WalletsLock.Lock()
						// BC.LocalWallets.TailBlockHashMap[BC.GenerateAddressFromPubkey(tx.PublicKey)] = block.Hash
						// BC.WalletsLock.Unlock()
						BC.LocalWallets.SetTailBlockHash(BC.GenerateAddressFromPubkey(tx.PublicKey), block.Hash)
					}
				}
				BC.LocalWallets.SaveToFile()
			}
			// fmt.Println("block:", block.BlockId, "校验失败")
		} else {
			el.Handle(total, fail)
		}
		BC.BlockQueue.Delete()

	}
}
