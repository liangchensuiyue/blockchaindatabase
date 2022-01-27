package blockchain

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

type ProofOfWork struct {
	block *Block

	target *big.Int
}

func NewProofOfWork(block *Block) *ProofOfWork {
	pow := ProofOfWork{
		block: block,
	}
	targetStr := "0000f00000000000000000000000000000000000000000000000000000000000"
	tmpInt := big.Int{}
	tmpInt.SetString(targetStr, 16)
	pow.target = &tmpInt
	return &pow
}

func (pow *ProofOfWork) Run() ([]byte, uint64) {
	block := pow.block
	var nonce uint64
	var hash [32]byte
	for {
		sc := bytes.Join([][]byte{
			Uint64Tobyte(block.Version),
			block.MerkelRoot,
			Uint64Tobyte(block.TimeStamp),
			Uint64Tobyte(block.Difficulty),
			Uint64Tobyte(nonce),
			// 只对区块头做哈希值
			// block.Data,
		}, []byte(""))
		hash = sha256.Sum256(sc)
		tmpInt := big.Int{}
		tmpInt.SetBytes(hash[:])
		if tmpInt.Cmp(pow.target) == -1 {
			// 找到
			fmt.Printf("挖矿成功 %x\n", hash)
			break
		} else {
			nonce++
		}
	}
	return hash[:], nonce
}
