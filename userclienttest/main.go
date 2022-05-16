package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	ucgrpc "go_code/基于区块链的非关系型数据库/proto/userclient"
	Type "go_code/基于区块链的非关系型数据库/type"
	"strconv"
	"sync"
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
func getMd5(str string) string {
	m := md5.New()
	m.Write([]byte(str))
	c := m.Sum(nil)
	return hex.EncodeToString(c)
}

var _lock1 *sync.Mutex = &sync.Mutex{}
var Total int = 0

func testget(name, pass string) {
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
	succ := 0
	fail := 0
	pre := time.Now().UnixNano()
	for i := 0; i < 100000; i++ {
		__e := client.Send(&ucgrpc.GetBody{
			Key:       fmt.Sprintf("key_%d", i),
			Passworld: pass,
			Sharemode: false,
			Username:  name,
			ShareChan: "",
		})
		if __e != nil {
			panic(__e)
		}
		re, e := client.Recv()
		if e != nil || !re.Status {
			fail++
			fmt.Printf("\rtotal: %d success:%d  fail:%d  %s not found", 100000, succ, fail, fmt.Sprintf("key_%d", i))
		} else {
			succ++
			fmt.Printf("\rtotal: %d success:%d  fail:%d", 100000, succ, fail)
			// fmt.Println("成功", fmt.Sprintf("key_%d", i))
		}
	}
	cur := time.Now().UnixNano()
	fmt.Println("\n耗时", (cur-pre)/1000000, "(ms)")
	fmt.Println("失败数量", fail)
}
func createuser(uname, pass string) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", "10.0.0.1", 3600), grpc.WithInsecure())
	if err != nil {
		fmt.Println("网络异常")
	}
	//网络延迟关闭
	defer conn.Close()

	//获得grpc句柄
	c := ucgrpc.NewUserClientServiceClient(conn)

	//通过句柄调用函数
	_, err = c.Newuser(context.Background(), &ucgrpc.UserInfo{
		Username:  uname,
		Passworld: pass,
	})
	if err != nil {
		// panic(err)
	}
}
func testput(uname, pass string, start, end int64) {
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

	// pre := time.Now().UnixNano()
	for i := start; i < end; i++ {

		err := client.Send(&ucgrpc.PutBody{
			Passworld: pass,
			Value:     []byte(fmt.Sprintf("%d", i)),
			Datatype:  Type.STRING,
			Strict:    true,
			Key:       fmt.Sprintf("key_%d", i),
			Username:  uname,
			Share:     false,
			ShareChan: "",
		})
		_lock1.Lock()
		Total++
		fmt.Printf("\r提交次数 %d", Total)
		_lock1.Unlock()
		if err != nil {
		}
	}
	time.Sleep(time.Second * 3)
	// cur := time.Now().UnixNano()
	// fmt.Println("耗时", (cur-pre)/1000000, "(ms)")
}
func Put(u_num int64, tx_num int64) {
	var i int64
	for i = 0; i < u_num; i++ {
		createuser(fmt.Sprintf("user%d", i), "123")
	}
	time.Sleep(time.Second * 3)
	ava_num := tx_num / u_num
	for i = 0; i < u_num; i++ {
		if i == u_num-1 {
			go testput(fmt.Sprintf("user%d", i), "123", i*ava_num, tx_num)
			break
		}
		go testput(fmt.Sprintf("user%d", i), "123", i*ava_num, i*ava_num+ava_num)
	}
}
func main() {
	var u_num int64
	var txnum int64
	var mode string
	flag.StringVar(&mode, "mod", "put", "模式put/get")
	flag.Int64Var(&u_num, "unum", 1, "并发数量")
	flag.Int64Var(&txnum, "tnum", 100000, "请求数量")
	flag.Parse()
	switch mode {
	case "put":
		Put(u_num, txnum)
	case "get":
		testget("user0", "123")
	}
	// createuser("lc", "123")
	// createuser("gds", "123")
	// createuser("zs", "123")
	// time.Sleep(time.Second * 2)
	// go testput("lc", "123", 0, 100000)
	// go testput("zs", "123")
	// go testput("ls", "123")
	// go testput("ww", "123")
	time.Sleep(time.Second * 100)
}
func main1() {
	md5Str := getMd5("key") //取得md5
	tempsubstr := md5Str[:16]
	hexVal, err := strconv.ParseInt(tempsubstr, 16, 64) //生成
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(tempsubstr)
	fmt.Println(hexVal)
}

// newuser gds 123
// newuser lc 123
// newuser zs 123
// newuser ls 123
// newuser ww 123
