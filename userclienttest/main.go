package main

import (
	"context"
	"fmt"
	ucgrpc "go_code/基于区块链的非关系型数据库/proto/userclient"
	"time"

	"google.golang.org/grpc"
)

func put(key string, value []byte, datatype string, pass string, user_address string, share bool, sharechan string, strict bool) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", "10.0.0.1", 3600), grpc.WithInsecure())
	if err != nil {
		fmt.Println("网络异常")
	}
	//网络延迟关闭
	defer conn.Close()

	//获得grpc句柄
	c := ucgrpc.NewUserClientServiceClient(conn)

	//通过句柄调用函数
	_, err = c.Put(context.Background(), &ucgrpc.PutBody{
		Passworld:   pass,
		Value:       value,
		Datatype:    datatype,
		Strict:      strict,
		Key:         key,
		UserAddress: user_address,
		Share:       share,
		ShareChan:   sharechan,
	})
	if err != nil {
		fmt.Println("put", err.Error())
	}
}
func get(key string, username string, user_address string, share bool, sharechan string) bool {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", "10.0.0.1", 3600), grpc.WithInsecure())
	if err != nil {
		fmt.Println("网络异常")
	}
	//网络延迟关闭
	defer conn.Close()

	//获得grpc句柄
	c := ucgrpc.NewUserClientServiceClient(conn)

	//通过句柄调用函数
	re, err := c.Get(context.Background(), &ucgrpc.GetBody{
		Key:         key,
		UserAddress: user_address,
		Sharemode:   share,
		Username:    username,
		ShareChan:   sharechan,
	})
	if err != nil {
		return false
	}
	return re.Status
}
func main() {

	nums := 1
	pre := time.Now().UnixNano()
	for i := nums; i < 500; i++ {
		put(fmt.Sprintf("key_%d", i), []byte(fmt.Sprintf("%d", i)), "int", "123", "1BZsJu1amTo2f5F3DZRYihk6Xjq3k7u4AD", false, "", true)
		flag := get(fmt.Sprintf("key_%d", i), "gds", "1BZsJu1amTo2f5F3DZRYihk6Xjq3k7u4AD", false, "")
		if !flag {
			fmt.Println("失败", fmt.Sprintf("key_%d", i))
		} else {
			// fmt.Println("成功", fmt.Sprintf("key_%d", i))
		}
	}
	cur := time.Now().UnixNano()
	fmt.Println("耗时", (cur-pre)/1000000, "(ms)")
}