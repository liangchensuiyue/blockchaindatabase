package quorum

import (
	"context"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"

	"google.golang.org/grpc"
)

func _startWork(blockQueue chan *BC.Block) {
	for {
		block := <-blockQueue

		conn, err := grpc.Dial("127.0.0.1:8899", grpc.WithInsecure())
		if err != nil {
			fmt.Println("网络异常", err)
		}
		//网络延迟关闭
		defer conn.Close()

		//获得grpc句柄
		c := proto.NewStudentServiceClient(conn)

		//通过句柄调用函数
		fmt.Println("ffffffffff")
		re, err := c.GetRealNameByUsername(context.Background(), &pd.MyRequest{Username: "熊猫"})
		if err != nil {
			fmt.Println("sayhello 服务调用失败")
		}
		fmt.Println("调用sayhello的返回", re)

	}
}
