package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

type BlockChainInfo struct {
	PassWorld         string
	NodeId            []string
	BlockTailHashName string
}

func main() {
	var info BlockChainInfo
	for {
		fmt.Printf("passworld:")
		fmt.Scanf("%s", &info.PassWorld)
		if strings.TrimSpace(info.PassWorld) == "" {
			continue
		}
		break
	}
	info.BlockTailHashName = "key_" + fmt.Sprintf("%d", time.Now().Unix())
	info.NodeId = []string{"127.0.0.1"}
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(info)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("genesis", aesEncrypt(buffer.Bytes(), []byte("1234567812345678")), 0644)
	if err != nil {
		panic(err)
	}
}

func aesDecrypt(codeText, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// 创建一个使用 ctr 分组
	iv := []byte("1234567812345678") // 这不是初始化向量，而是给一个随机种子，大小必须与blocksize 相等
	stream := cipher.NewCTR(block, iv)
	// 加密
	dst := make([]byte, len(codeText))
	stream.XORKeyStream(dst, codeText)
	return dst
}

// AES  加解密
func aesEncrypt(plainText, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// 创建一个使用 ctr 分组
	iv := []byte("1234567812345678") // 这不是初始化向量，而是给一个随机种子，大小必须与blocksize 相等
	stream := cipher.NewCTR(block, iv)
	// 加密
	dst := make([]byte, len(plainText))
	a := make([]byte, len(plainText))
	stream.XORKeyStream(dst, plainText)
	stream.XORKeyStream(a, plainText) // dst != a
	return dst
}
func load() {
	var info BlockChainInfo

	_, err := os.Stat("genesis")
	if os.IsNotExist(err) {
		panic(err)
	}
	// 读取钱包
	content, err := ioutil.ReadFile("genesis")
	if err != nil {
		panic(err)
	}

	decoder := gob.NewDecoder(bytes.NewReader(aesDecrypt(content, []byte("1234567812345678"))))
	err = decoder.Decode(&info)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(info)
}
