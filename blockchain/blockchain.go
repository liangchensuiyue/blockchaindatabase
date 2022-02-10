package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

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
	return NewBlock([]*Transaction{})
}

// 添加区块
func (bc *BlockChain) AddBlock(newblock *Block) error {
	db := bc.Db //区块链数据库

	e := db.Update(func(tx *bolt.Tx) error {

		//完成数据添加
		bucket := tx.Bucket([]byte(bc.BlockBucket))
		if bucket == nil {
			return errors.New("bucket 不应该为空，请检查!")
		}

		bucket.Put(newblock.Hash, newblock.Serialize())
		bucket.Put([]byte(bc.BlockTailHashKey), newblock.Hash)

		return nil
	})
	return e
}
func (bc *BlockChain) GetTailBlock() *Block {
	hash := bc.GetTailBlockHash()
	bl, _ := bc.GetBlockByHash(hash)
	return bl
}

// 遍历区块链
func (bc *BlockChain) traverse(handle func(*Block, error)) {
	cur_block := bc.GetTailBlock()
	var e error = nil
	handle(cur_block, nil)
	for {
		cur_block, e = bc.GetBlockByHash(cur_block.PreBlockHash)
		handle(cur_block, e)
		if e != nil {
			break
		}
	}
}

// 使用集群成员共有的私钥对区块进行签名
func (bc *BlockChain) SignBlock(groupPriKey *ecdsa.PrivateKey, user_address string, newblock *Block) {
	bc_pre_block := bc.GetTailBlock()
	for _, tx := range newblock.TxInfos {
		tx.Sign(user_address)
	}
	newblock.BlockId = bc_pre_block.BlockId + 1
	newblock.PreBlockHash = bc_pre_block.Hash

	newblock.MerkelRoot = newblock.MakeMerkelRoot()
	hash := sha256.Sum256(newblock.Serialize())
	newblock.Hash = hash[:]

	r, s, err := ecdsa.Sign(rand.Reader, groupPriKey, newblock.Hash)
	if err != nil {
		panic(err)
	}
	newblock.Signature = append(r.Bytes(), s.Bytes()...)
}

// 对区块进行验证
func (bc *BlockChain) VerifyBlock(groupPubKey *ecdsa.PublicKey, user_address string, newblock *Block) bool {
	bc_pre_block := bc.GetTailBlock()
	for _, tx := range newblock.TxInfos {
		flag := tx.Verify(user_address)
		if !flag {
			return flag
		}
	}
	// 区块是否连续
	if bc_pre_block.BlockId != newblock.BlockId-1 {
		fmt.Println("区块验证: 区块不连续")
		return false
	}
	// 前区块是否存在
	_, e := bc.GetBlockByHash(newblock.PreBlockHash)
	if e != nil {
		fmt.Println("区块验证: 区块前hash不存在")
		return false
	}

	newblock.BlockId = bc_pre_block.BlockId + 1
	newblock.PreBlockHash = bc_pre_block.Hash

	if !bytes.Equal(newblock.MerkelRoot, newblock.MakeMerkelRoot()) {
		fmt.Println("区块验证: 默克尔根错误")
		return false
	}
	r := big.Int{}
	s := big.Int{}

	r.SetBytes(newblock.Signature[:len(newblock.Signature)/2])
	s.SetBytes(newblock.Signature[len(newblock.Signature)/2:])

	if !ecdsa.Verify(groupPubKey, newblock.Hash, &r, &s) {
		return false
	}
	return true
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
