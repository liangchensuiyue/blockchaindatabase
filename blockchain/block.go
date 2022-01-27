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
	// 时间戳
	Timestamp uint64
	// 当前区块hash,正常比特币区块中没有当前区块hash，这里是为了方便做了简化
	Hash    string
	TxInfos []*Transaction
	//
}

// 实现一个辅助函数，uint64 -> []byte
func Uint64Tobyte(src uint64) []byte {
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, src)
	return buffer.Bytes()
}

// 创建区块
func NewBlock(txs []*Transaction, prevBlockHash []byte) *Block {
	block := Block{
		Version:      00,
		PrevHash:     prevBlockHash,
		MerkelRoot:   []byte{},
		TimeStamp:    uint64(time.Now().Unix()),
		Difficulty:   0,        // 随便填写的无效值
		Nonce:        0,        // 同上
		Hash:         []byte{}, // 先填空，后面计算
		Transactions: txs,
	}
	block.MerkelRoot = block.MakeMerkelRoot()
	// block.SetHash()
	pow := NewProofOfWork(&block)
	hash, nonce := pow.Run()
	block.Hash = hash
	block.Nonce = nonce
	return &block
}

// 生成 hash
// func(block *Block) SetHash(){
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
// 	},[]byte(""))
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
func Deserialize(data []byte) Block {
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
	for _, tx := range block.Transactions {
		info = append(info, tx.TXID...)
	}
	hash := sha256.Sum256(info)
	return hash[:]
}
