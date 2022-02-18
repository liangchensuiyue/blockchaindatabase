package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	quorum "go_code/基于区块链的非关系型数据库/quorum"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	var localNode quorum.BlockChainNode
	localNode.BCInfo = &quorum.BlockChainInfo{}
	localNode.LocalPort = 3300
	localNode.BCInfo.TailBlockId = 0
	localNode.BCInfo.PassWorld = "pass_" + fmt.Sprintf("%d", time.Now().Unix())
	localNode.BCInfo.BlockChainDB = "block"
	// ip, err := util.GetLocalIp()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// localNode.LocalIp = ip.String()
	localNode.LocalIp = "10.0.0.1"
	curve := elliptic.P256()

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic()
	}
	pubkeyOrig := privateKey.PublicKey
	pubKey := append(pubkeyOrig.X.Bytes(), pubkeyOrig.Y.Bytes()...)
	localNode.BCInfo.PriKey = privateKey
	localNode.BCInfo.PubKey = pubKey
	localNode.BCInfo.BlockTailHashKey = "key_" + fmt.Sprintf("%d", time.Now().Unix())
	localNode.Quorum = append(localNode.Quorum, &quorum.BlockChainNode{
		LocalIp:   localNode.LocalIp,
		LocalPort: localNode.LocalPort,
	})
	var buffer bytes.Buffer
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&buffer)
	err = encoder.Encode(localNode)
	if err != nil {
		fmt.Println("密钥生成失败")
		panic(err)
	}
	err = ioutil.WriteFile("genesis", quorum.AesEncrypt(buffer.Bytes(), []byte("1234567812345678")), 0644)
	if err != nil {
		panic(err)
	}
}

func load() {
	var info quorum.BlockChainInfo

	_, err := os.Stat("genesis")
	if os.IsNotExist(err) {
		panic(err)
	}
	// 读取钱包
	content, err := ioutil.ReadFile("genesis")
	if err != nil {
		panic(err)
	}

	decoder := gob.NewDecoder(bytes.NewReader(quorum.AesDecrypt(content, []byte("1234567812345678"))))
	err = decoder.Decode(&info)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(info)
}
