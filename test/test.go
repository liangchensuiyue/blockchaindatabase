package test

import (
	"fmt"
	db "go_code/基于区块链的非关系型数据库/database"
	"time"
)

var nums chan int = make(chan int, 2)

func Test2() {

}
func Test3() {
	pre := time.Now().UnixNano()
	nums := 200
	for i := 0; i < nums; i++ {
		// time.Sleep(time.Second)

		_, n := db.Get(fmt.Sprintf("key_%d", i), "1KqpoNxAdGkzeEuWeAvntM2k38tNVA7TVK", false, []string{})
		if n == -1 {
			fmt.Println("失败", fmt.Sprintf("key_%d", i))
		} else {
			// fmt.Println("成功", fmt.Sprintf("key_%d", i))
		}
	}
	cur := time.Now().UnixNano()
	fmt.Printf("耗时:%d ms\n", (cur-pre)/1000000)
}
func Test1() {
	nums := 200
	pre := time.Now().UnixNano()
	for i := 0; i < nums; i++ {
		db.PutTest(fmt.Sprintf("key_%d", i), []byte(fmt.Sprintf("%d", i)), "int", "1KqpoNxAdGkzeEuWeAvntM2k38tNVA7TVK", false, []string{}, true, func() {
			nums--
			if nums <= 0 {
				cur := time.Now().UnixNano()
				fmt.Printf("%d 笔交易 耗时 %d ms\n", 200, (cur-pre)/1000000)
			}
		})
	}
}
