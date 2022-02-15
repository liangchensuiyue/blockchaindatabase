package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	quorum "go_code/基于区块链的非关系型数据库/quorum"
	"go_code/基于区块链的非关系型数据库/util"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

func main1() {
	var localNode quorum.BlockChainNode
	localNode.BCInfo = &quorum.BlockChainInfo{}
	localNode.BCInfo.TailBlockId = 0
	for {
		fmt.Printf("绑定端口:")
		fmt.Scanf("%d\n", &localNode.LocalPort)
		// 校验端口是否合规
		if localNode.LocalPort >= 65535 {
			continue
		}
		break
	}
	ip, err := util.GetLocalIp()
	if err != nil {
		fmt.Println(err)
	}
	localNode.LocalIp = ip.String()
	for {
		fmt.Printf("集群访问密码:")
		fmt.Scanf("%s\n", &localNode.BCInfo.PassWorld)
		if strings.TrimSpace(localNode.BCInfo.PassWorld) == "" {
			continue
		}
		break
	}
	for {
		fmt.Printf("块数据存储文件名称:")
		fmt.Scanf("%s", &localNode.BCInfo.BlockChainDB)
		if strings.TrimSpace(localNode.BCInfo.BlockChainDB) == "" {
			continue
		}
		break
	}

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
