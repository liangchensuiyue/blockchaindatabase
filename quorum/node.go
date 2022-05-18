package quorum

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"errors"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	"go_code/基于区块链的非关系型数据库/test"
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
var LocalBlockChain *BC.BlockChain
var LocalNode *BlockChainNode

func StartGrpcWork(lbc *BC.BlockChain) {
	LocalBlockChain = lbc
	go _startServer()

	go _starDistributeBlock()
	go _startHeartbeat()
	test.SystemInfo.PrintBlockInfo = func() {
		fmt.Println(NUM, "个区块同步耗时", Total/1000000, "(ms)")
	}
	test.SystemInfo.CleanBlockInfo = func() {
		NUM = 0
		Total = 0
	}
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
	err := encoder.Encode(LocalNode)
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
	LocalNode = &info
	// ip, e := util.GetLocalIp()
	// if e != nil {
	// 	return nil, e
	// }
	LocalNode.LocalIp = "10.0.0.1"
	// localNode.LocalIp = ip.String()
	// JointoGroup(localNode.BCInfo.PassWorld, ip.String(), int32(localNode.LocalPort))
	JointoGroup(LocalNode.BCInfo.PassWorld, "10.0.0.1", int32(LocalNode.LocalPort))
	return LocalNode, nil
}
func JointoGroup(passworld, local_ip string, local_port int32) error {
	if passworld != LocalNode.BCInfo.PassWorld {
		return errors.New("访问密码错误")
	}
	flag := false
	for _, node := range LocalNode.Quorum {
		if node.LocalIp == local_ip {
			node.LocalPort = int(local_port)
			flag = true
		}
	}
	if !flag {
		LocalNode.Quorum = append(LocalNode.Quorum, &BlockChainNode{
			LocalIp:   local_ip,
			LocalPort: int(local_port),
		})
	}
	SaveGenesisFileToDisk()
	return nil
}
