package quorum

import (
	"fmt"
	bcgrpc "go_code/基于区块链的非关系型数据库/proto"
	"net"

	"google.golang.org/grpc"
)

func _startServer() {
	//创建网络
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", localNode.LocalPort))
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
func _starDistributeBlock(blockQueue chan queueObject) {
	for {
		el := <-blockQueue
		block := el.TargetBlock
		total := 0
		fail := 0
		fmt.Printf("区块 %d 分发:\n", block.BlockId)
		for _, blockBlockChainNode := range localNode.quorum {
			total++
			DistributeBlock(block, blockBlockChainNode, func(res *bcgrpc.VerifyInfo, err error) {
				if err != nil {
					fmt.Println(err)
					fail++
					fmt.Printf("节点 %s:%d 接受失败", blockBlockChainNode.LocalIp, blockBlockChainNode.LocalPort)
					return
				}
				if !res.Status {
					fail++
					fmt.Printf("节点 %s:%d 校验失败", blockBlockChainNode.LocalIp, blockBlockChainNode.LocalPort)
					return
				}
				fmt.Printf("节点 %s:%d 接受成功", blockBlockChainNode.LocalIp, blockBlockChainNode.LocalPort)
			})
		}
		el.Handle(total, fail)

	}
}
