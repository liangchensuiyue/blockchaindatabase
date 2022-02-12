package quorum

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"errors"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
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

var blockQueue chan *BC.Block
var localBlockChain *BC.BlockChain
var localnode *BlockChainNode

func StartGrpcWork(localnode *BlockChainNode, localBlockChain *BC.BlockChain) {
	blockQueue = make(chan *BC.Block, 100)
	go _starDistributeBlock(blockQueue)
	go _startServer()
}

func (node *BlockChainNode) DistribuBlock(newblock *BC.Block) {
	blockQueue <- newblock
}
func JointoGroup(passworld, local_ip string, local_port int32) error {
	if passworld != localnode.BCInfo.PassWorld {
		return errors.New("访问密码错误")
	}
	flag := false
	for _, node := range localnode.quorum {
		if node.LocalIp == local_ip {
			node.LocalPort = int(local_port)
			flag = true
		}
	}
	if flag {
		localnode.quorum = append(localnode.quorum, &BlockChainNode{
			LocalIp:   local_ip,
			LocalPort: int(local_port),
		})
	}
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
