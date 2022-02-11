package quorum

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
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

func StartGrpcWork(localnode *BlockChainNode) {
	blockQueue = make(chan *BC.Block, 100)
	_startWork(blockQueue, localnode)
}

func (node *BlockChainNode) DistribuBlock(newblock *BC.Block) {
	blockQueue <- newblock
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
