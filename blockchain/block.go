package blockchain

import (
	// "crypto/sha256"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"time"
)

type Block struct {
	// 前区块hash
	PreBlockHash []byte
	// 块 序号
	BlockId uint64
	// 时间戳
	Timestamp uint64
	// 当前区块hash,正常比特币区块中没有当前区块hash，这里是为了方便做了简化
	Hash []byte

	MerkelRoot []byte

	TxInfos []*Transaction

	// 由集群私钥加密
	Signature []byte
	//
}

// 实现一个辅助函数，uint64 -> []byte
func Uint64Tobyte(src uint64) []byte {
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, src)
	return buffer.Bytes()
}

// 创建区块
func NewBlock(txinfos []*Transaction) *Block {
	block := Block{
		PreBlockHash: []byte{},
		MerkelRoot:   []byte{},
		Signature:    []byte{},
		// 前区块hash
		// PreBlockHash: pre_block_hash,
		// 块 序号
		// BlockId: block_id,
		// 时间戳
		Timestamp: uint64(time.Now().Unix()),
		// 当前区块hash,正常比特币区块中没有当前区块hash，这里是为了方便做了简化
		TxInfos: txinfos,
	}
	// block.MerkelRoot = block.MakeMerkelRoot()
	// hash := sha256.Sum256(block.Serialize())
	// block.Hash = hash[:]
	return &block
}

// 生成 hash
// func (block *Block) SetHash() {
// 	// blockInfo := append(block.PrevHash, Uint64Tobyte(block.Version)...)
// 	// blockInfo  = append(blockInfo, block.MerkelRoot...)
// 	// blockInfo  = append(blockInfo, Uint64Tobyte(block.TimeStamp)...)
// 	// blockInfo  = append(blockInfo, Uint64Tobyte(block.Difficulty)...)
// 	// blockInfo  = append(blockInfo, Uint64Tobyte(block.Nonce)...)
// 	// blockInfo  = append(blockInfo, block.Data...)
// 	sc := bytes.Join([][]byte{
// 		Uint64Tobyte(block.Version),
// 		block.MerkelRoot,
// 		Uint64Tobyte(block.TimeStamp),
// 		Uint64Tobyte(block.Difficulty),
// 		Uint64Tobyte(block.Nonce),
// 		block.Data,
// 	}, []byte(""))
// 	hash := sha256.Sum256(sc)
// 	block.Hash = hash[:]
// }
func (block *Block) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(block)
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
	// return []byte{}
}

func BlockDeserialize(data []byte) Block {
	var buffer *bytes.Buffer = bytes.NewBuffer(data)
	decore := gob.NewDecoder(bytes.NewReader(buffer.Bytes()))
	var block Block
	err := decore.Decode(&block)
	if err != nil {
		panic(err)
	}
	return block
}

// 模拟梅克尔根，这里只是堆交易的数据做简单的拼接，而不错二叉树
func (block *Block) MakeMerkelRoot() []byte {
	var info []byte
	for _, tx := range block.TxInfos {
		info = append(info, tx.Hash...)
	}
	hash := sha256.Sum256(info)
	return hash[:]
}

// 判断是否为创世区块
func (block *Block) IsGenesisBlock() bool {
	if block.BlockId == 1 {
		return true
	} else {
		return false
	}
}
