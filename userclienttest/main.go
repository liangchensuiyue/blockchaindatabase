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
func testput(uname, pass string) {
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
	for i := nums; i < 100; i++ {
		err := client.Send(&ucgrpc.PutBody{
			Passworld: pass,
			Value:     []byte(fmt.Sprintf("%d", i)),
			Datatype:  Type.INT32,
			Strict:    true,
			Key:       fmt.Sprintf("key_%d", i),
			Username:  uname,
			Share:     false,
			ShareChan: "",
		})
		if err != nil {
			fmt.Println(err)
		}
	}
	time.Sleep(time.Second * 140)
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
	go testput("gds", "123")
	// go testput("lc", "123")
	// go testput("zs", "123")
	// go testput("ls", "123")
	// go testput("ww", "123")
	time.Sleep(time.Second * 100)
}

// newuser gds 123
// newuser lc 123
// newuser zs 123
// newuser ls 123
// newuser ww 123
