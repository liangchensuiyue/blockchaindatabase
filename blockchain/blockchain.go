package blockchain

import (
	"errors"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

// 引入区块链
type BlockChain struct {
	Db                   *bolt.DB
	PassWorld            string
	NodeId               []string
	BlockBucket          string
	BlockTailHashKey     string
	BlockChainDBFileName string
	// Tail                 []byte // 存储最后一个区块的hash
}

const blockChainDB = "blockChain.db"
const blockBucket = "blockBucket"
const LastHashKey = "lastkey"

func NewBlockChain(passworld string, NodeId []string, blockTailHashKey, blockChainDBFileName string) *BlockChain {
	// 创建一个创世块，并作为第一个区块添加到区块链中
	db, err := bolt.Open(blockChainDBFileName, 0600, nil)
	if err != nil {
		panic(err)
	}
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			// 没有该bucket,需要创建
			bucket, err = tx.CreateBucket([]byte(blockBucket))
			if err != nil {
				panic(err)
			}
			genesisBlock := GenesisBlock()
			bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
			bucket.Put([]byte(blockTailHashKey), genesisBlock.Hash)
		}
		return nil
	})
	return &BlockChain{
		Db:                   db,
		PassWorld:            passworld,
		BlockBucket:          blockBucket,
		BlockChainDBFileName: blockChainDBFileName,
		BlockTailHashKey:     blockTailHashKey,
	}
}

// 创世块
func GenesisBlock() *Block {
	// coinbase := NewCoinbaseTX(block_hash, block_id, pre_block_hash)
	return NewBlock(1, []byte(""), []*Transaction{})
}

// 添加区块
func (bc *BlockChain) AddBlock(newblock *Block) {
	for _, tx := range txs {
		if !bc.VerifyTransaction(tx) {
			fmt.Println("矿工发现无效交易")
			return
		}
	}
	//如何获取前区块的哈希呢？？
	db := bc.db         //区块链数据库
	lastHash := bc.tail //最后一个区块的哈希

	db.Update(func(tx *bolt.Tx) error {

		//完成数据添加
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			log.Panic("bucket 不应该为空，请检查!")
		}

		//a. 创建新的区块
		block := NewBlock(txs, lastHash)

		//b. 添加到区块链db中
		//hash作为key， block的字节流作为value，尚未实现
		bucket.Put(block.Hash, block.Serialize())
		bucket.Put([]byte("LastHashKey"), block.Hash)

		//c. 更新一下内存中的区块链，指的是把最后的小尾巴tail更新一下
		bc.tail = block.Hash

		return nil
	})
}
func (bc *BlockChain) GetTailBlockHash() []byte {
	var blockhash []byte
	bc.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bc.BlockBucket))
		if bucket == nil {
			panic("bucket is empty!")
		}
		hashkey := bucket.Get([]byte(bc.BlockTailHashKey))
		blockhash = bucket.Get(hashkey)
		return nil
	})
	return blockhash
}
func (bc *BlockChain) GetBlockByHash(hash []byte) (*Block, error) {
	block := Block{}
	if len(hash) == 0 {
		return nil, errors.New("错误的hash")

	}
	bc.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bc.BlockBucket))
		if bucket == nil {
			panic("bucket is empty!")
		}
		value := bucket.Get(hash)
		block = BlockDeserialize(value)
		return nil
	})
	return &block, nil
}
