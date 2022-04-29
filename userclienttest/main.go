package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	ucgrpc "go_code/基于区块链的非关系型数据库/proto/userclient"
	Type "go_code/基于区块链的非关系型数据库/type"
	"time"

	"google.golang.org/grpc"
)

func Int16Tobytes(n int) []byte {
	data := int16(n)
	bytebuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytebuffer, binary.BigEndian, data)
	return bytebuffer.Bytes()
}
func BytesToInt16(bts []byte) int {
	bytebuffer := bytes.NewBuffer(bts)
	var data int16
	binary.Read(bytebuffer, binary.BigEndian, &data)

	return int(data)
}
func testput() {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", "10.0.0.1", 3600), grpc.WithInsecure())
	if err != nil {
		fmt.Println("网络异常")
	}
	//网络延迟关闭
	defer conn.Close()

	//获得grpc句柄
	c := ucgrpc.NewUserClientServiceClient(conn)

	//通过句柄调用函数
	client, e := c.Put(context.Background())
	if e != nil {
		fmt.Println("put", err.Error())
	}

	nums := 1
	// pre := time.Now().UnixNano()
	for i := nums; i < 1000; i++ {
		err := client.Send(&ucgrpc.PutBody{
			Passworld: "123",
			Value:     []byte(fmt.Sprintf("%d", i)),
			Datatype:  Type.INT32,
			Strict:    true,
			Key:       fmt.Sprintf("key_%d", i),
			Username:  "gds",
			Share:     false,
			ShareChan: "",
		})
		if err != nil {
			fmt.Println(err)
		}
	}
	time.Sleep(time.Second * 14)
	// cur := time.Now().UnixNano()
	// fmt.Println("耗时", (cur-pre)/1000000, "(ms)")
}
func testget() {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", "10.0.0.1", 3600), grpc.WithInsecure())
	if err != nil {
		fmt.Println("网络异常")
	}
	//网络延迟关闭
	defer conn.Close()

	//获得grpc句柄
	c := ucgrpc.NewUserClientServiceClient(conn)

	//通过句柄调用函数
	client, err := c.Get(context.Background())
	if err != nil {
		panic(err)
	}
	nums := 1
	pre := time.Now().UnixNano()
	for i := nums; i < 500; i++ {
		client.Send(&ucgrpc.GetBody{
			Key:       fmt.Sprintf("key_%d", i),
			Passworld: "123",
			Sharemode: false,
			Username:  "gds",
			ShareChan: "",
		})
		re, e := client.Recv()
		if !re.Status || e != nil {
			fmt.Println("失败", fmt.Sprintf("key_%d", i))
		} else {
			// fmt.Println("成功", fmt.Sprintf("key_%d", i))
		}
	}
	cur := time.Now().UnixNano()
	fmt.Println("耗时", (cur-pre)/1000000, "(ms)")
}
func main() {
	testput()
}
