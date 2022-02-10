package main

import (
	"fmt"
)

func init() {
	fmt.Println("init")
}

var blockBucket string = "gds"

func main() {
	a := []byte("123")
	fmt.Println(a[:2])
	// var lastHash []byte
	// db, err := bolt.Open("blockChainDB", 0600, nil)
	// if err != nil {
	// 	panic(err)
	// }
	// db.Update(func(tx *bolt.Tx) error {
	// 	bucket := tx.Bucket([]byte(blockBucket))
	// 	if bucket == nil {
	// 		fmt.Println("create")
	// 		// 没有该bucket,需要创建
	// 		bucket, err = tx.CreateBucket([]byte(blockBucket))
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		bucket.Put([]byte("LastHashKey"), []byte("gdddddddd"))
	// 		lastHash = []byte("LastHashKey")
	// 	} else {
	// 		lastHash = bucket.Get([]byte("LastHashKeyff"))
	// 	}
	// 	return nil
	// })
	// fmt.Println(lastHash)
	// fmt.Println(fmt.Sprintf("%d", time.Now().Unix()))
}
