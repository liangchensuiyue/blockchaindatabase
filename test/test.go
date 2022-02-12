package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"os"
	"strings"
)

func init() {
	fmt.Println("init")
}

var blockBucket string = "gds"

type node struct {
	Id   int
	Name string
	Tx   []*int
}

func g(node1 *node) []byte {

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(node1)
	if err != nil {
		panic(err)
	}
	data := buffer.Bytes()
	hash1 := sha256.Sum256(data)
	return hash1[:]
}
func main() {
	for {
		reader := bufio.NewReader(os.Stdin)
		s, _, _ := reader.ReadLine()
		strs := strings.Split(string(s), " ")
		for _, v := range strs {
			if v == "" {
				continue
			}
			fmt.Printf("--%s--\n", v)
		}
	}
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
