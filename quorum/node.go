package quorum

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"errors"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	util "go_code/基于区块链的非关系型数据库/util"

	"io/ioutil"
	"os"
)

type BlockChainInfo struct {
	PassWorld        string
	BlockTailHashKey string
	TailBlockId      uint64
	BlockChainDB     string

	//用于区块的校验；同一个集群的通信
	PriKey *ecdsa.PrivateKey
	PubKey []byte
}
type BlockChainNode struct {
	LocalIp   string
	LocalPort int

	BCInfo *BlockChainInfo
	Quorum []*BlockChainNode
}

var isAccountant bool = false

func StartGrpcWork() {
	go _startServer()

	go _starDistributeBlock()
	go _startHeartbeat()
	// go func() {
	// 	for {
	// 		time.Sleep(time.Second * 2)
	// 		fmt.Println(NUM, "笔交易耗时", Total/1000000, "(ms)")
	// 	}
	// }()
}

func (node *BlockChainNode) DistribuBlock(newblock *BC.Block, handle func(int, int)) {
	BC.BlockQueue.Insert(BC.QueueObject{
		TargetBlock: newblock,
		Handle:      handle,
	})
	// BlockQueue <-
}
func LocalNodeIsAccount() bool {
	return isAccountant
}

func SaveGenesisFileToDisk() {
	var buffer bytes.Buffer
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(localNode)
	if err != nil {
		fmt.Println("SaveGenesisFileToDisk 失败")
		panic(err)
	}
	err = ioutil.WriteFile("genesis", util.AesEncrypt(buffer.Bytes(), []byte("1234567812345678")), 0644)
	if err != nil {
		panic(err)
	}
}
func LoadGenesisFile(filename string) (*BlockChainNode, error) {
	var info BlockChainNode

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil, err
	}
	// 读取钱包
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// 解码
	gob.Register(elliptic.P256())

	decoder := gob.NewDecoder(bytes.NewReader(util.AesDecrypt(content, []byte("1234567812345678"))))
	err = decoder.Decode(&info)
	if err != nil {
		panic(err)
	}
	localNode = &info
	ip, e := util.GetLocalIp()
	if e != nil {
		return nil, e
	}
	// localNode.LocalIp = "10.0.0.1"
	localNode.LocalIp = ip.String()
	JointoGroup(localNode.BCInfo.PassWorld, ip.String(), int32(localNode.LocalPort))
	// JointoGroup(localNode.BCInfo.PassWorld, "10.0.0.1", int32(localNode.LocalPort))
	return localNode, nil
}
func JointoGroup(passworld, local_ip string, local_port int32) error {
	if passworld != localNode.BCInfo.PassWorld {
		return errors.New("访问密码错误")
	}
	flag := false
	for _, node := range localNode.Quorum {
		if node.LocalIp == local_ip {
			node.LocalPort = int(local_port)
			flag = true
		}
	}
	if !flag {
		localNode.Quorum = append(localNode.Quorum, &BlockChainNode{
			LocalIp:   local_ip,
			LocalPort: int(local_port),
		})
	}
	SaveGenesisFileToDisk()
	return nil
}
