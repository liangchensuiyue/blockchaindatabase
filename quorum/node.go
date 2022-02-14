package quorum

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"errors"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	"go_code/基于区块链的非关系型数据库/util"

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
	quorum []*BlockChainNode
}
type queueObject struct {
	TargetBlock *BC.Block
	Handle      func(int, int)
}

var blockQueue chan queueObject

func StartGrpcWork() {
	blockQueue = make(chan queueObject, 100)
	go _starDistributeBlock(blockQueue)
	go _startServer()
}

func (node *BlockChainNode) DistribuBlock(newblock *BC.Block, handle func(int, int)) {
	blockQueue <- queueObject{
		TargetBlock: newblock,
		Handle:      handle,
	}
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
	err = ioutil.WriteFile("genesis", AesEncrypt(buffer.Bytes(), []byte("1234567812345678")), 0644)
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

	decoder := gob.NewDecoder(bytes.NewReader(AesDecrypt(content, []byte("1234567812345678"))))
	err = decoder.Decode(&info)
	if err != nil {
		panic(err)
	}
	localNode = &info
	ip, e := util.GetLocalIp()
	if e != nil {
		return nil, e
	}
	JointoGroup(localBlockChain.PassWorld, ip.String(), int32(localNode.LocalPort))
	return localNode, nil
}
func JointoGroup(passworld, local_ip string, local_port int32) error {
	if passworld != localNode.BCInfo.PassWorld {
		return errors.New("访问密码错误")
	}
	flag := false
	for _, node := range localNode.quorum {
		if node.LocalIp == local_ip {
			node.LocalPort = int(local_port)
			flag = true
		}
	}
	if flag {
		localNode.quorum = append(localNode.quorum, &BlockChainNode{
			LocalIp:   local_ip,
			LocalPort: int(local_port),
		})
	}
	SaveGenesisFileToDisk()
	return nil
}

func AesDecrypt(codeText, key []byte) []byte {
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
func AesEncrypt(plainText, key []byte) []byte {
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
