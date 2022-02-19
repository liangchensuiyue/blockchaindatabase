package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/boltdb/bolt"
)

// 引入区块链
type BlockChain struct {
	Db *bolt.DB
	// PassWorld            string
	BlockBucket          string
	BlockTailHashKey     string
	TailUserBlockHashkey string
	BlockChainDBFileName string
	// Tail                 []byte // 存储最后一个区块的hash
}

const blockChainDB = "blockChain.db"
const blockBucket = "blockBucket"
const LastHashKey = "lastkey"

func NewBlockChain(blockTailHashKey, blockChainDBFileName string) *BlockChain {
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
			// genesisBlock := GenesisBlock()
			// bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
			// bucket.Put([]byte(blockTailHashKey), genesisBlock.Hash)
		}
		return nil
	})
	return &BlockChain{
		Db:                   db,
		BlockBucket:          blockBucket,
		BlockChainDBFileName: blockChainDBFileName,
		TailUserBlockHashkey: "user",
		BlockTailHashKey:     blockTailHashKey,
	}
}

// 创世块
func NewGenesisBlock() *Block {
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
		// fmt.Println(newblock.BlockId, newblock.Hash)
		bucket.Put([]byte(bc.BlockTailHashKey), newblock.Hash)

		return nil
	})
	return e
}

// 获取区块链中最后一个区块
func (bc *BlockChain) GetTailBlock() (*Block, error) {
	hash := bc.GetTailBlockHash()
	bl, err := bc.GetBlockByHash(hash)
	if err != nil {
		// fmt.Println(err)
		return bl, err
	}
	return bl, nil
}

// 获取关于用户信息的最后一个区块信息
func (bc *BlockChain) GetTailUserBlock() (*Block, error) {
	hash := bc.GetTailBlockHash()
	bl, err := bc.GetBlockByHash(hash)
	if err != nil {
		// fmt.Println(err)
		return bl, err
	}
	return bl, nil
}

// 遍历区块链
func (bc *BlockChain) Traverse(handle func(*Block, error)) {
	cur_block, err := bc.GetTailBlock()
	if err != nil {
		handle(cur_block, nil)
		return
	}
	var e error = nil
	handle(cur_block, nil)
	for {
		cur_block, e = bc.GetBlockByHash(cur_block.PreBlockHash)
		handle(cur_block, e)
		if e != nil || cur_block.IsGenesisBlock() {
			break
		}
	}
}

// 使用集群成员共有的私钥对区块进行签名
func (bc *BlockChain) SignBlock(groupPriKey *ecdsa.PrivateKey, IsGenesisBlock bool, newblock *Block) {
	bc_pre_block, _ := bc.GetTailBlock()

	if IsGenesisBlock {
		// 创世块
		newblock.BlockId = 1
		newblock.PreBlockHash = []byte{}
		newblock.MerkelRoot = []byte{}
		hash := sha256.Sum256(newblock.Serialize())
		newblock.Hash = hash[:]
	} else {
		for _, tx := range newblock.TxInfos {
			tx.Sign()
		}
		newblock.BlockId = bc_pre_block.BlockId + 1
		newblock.PreBlockHash = bc_pre_block.Hash

		newblock.MerkelRoot = newblock.MakeMerkelRoot()
		hash := sha256.Sum256(newblock.Serialize())
		newblock.Hash = hash[:]
	}

	r, s, err := ecdsa.Sign(rand.Reader, groupPriKey, newblock.Hash)
	if err != nil {
		panic(err)
	}
	newblock.Signature = append(r.Bytes(), s.Bytes()...)
}

// 对区块进行验证
func (bc *BlockChain) VerifyBlock(groupPubKey []byte, newblock *Block) bool {
	X := big.Int{}
	Y := big.Int{}

	X.SetBytes(groupPubKey[:len(groupPubKey)/2])
	Y.SetBytes(groupPubKey[len(groupPubKey)/2:])

	pubKeyOrigin := ecdsa.PublicKey{Curve: elliptic.P256(), X: &X, Y: &Y}
	bc_pre_block, _ := bc.GetTailBlock()
	if newblock.BlockId == 1 {
		r := big.Int{}
		s := big.Int{}

		r.SetBytes(newblock.Signature[:len(newblock.Signature)/2])
		s.SetBytes(newblock.Signature[len(newblock.Signature)/2:])

		_sig := newblock.Signature
		_hash := newblock.Hash
		newblock.Signature = []byte{}
		newblock.Hash = []byte{}
		hash := sha256.Sum256(newblock.Serialize())

		if !ecdsa.Verify(&pubKeyOrigin, hash[:], &r, &s) {
			return false
		}
		newblock.Signature = _sig
		newblock.Hash = _hash
		return true
	}
	for _, tx := range newblock.TxInfos {
		flag := tx.Verify()
		if !flag {
			fmt.Println("区块中交易校验失败")
			return flag
		}
	}
	// 区块是否连续
	if bc_pre_block.BlockId != newblock.BlockId-1 || !bytes.Equal(bc_pre_block.Hash, newblock.PreBlockHash) {
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
		log.Println("区块验证: 默克尔根错误")
		return false
	}
	r := big.Int{}
	s := big.Int{}

	r.SetBytes(newblock.Signature[:len(newblock.Signature)/2])
	s.SetBytes(newblock.Signature[len(newblock.Signature)/2:])

	_sig := newblock.Signature
	newblock.Signature = []byte{}
	newblock.Hash = []byte{}
	_hash := sha256.Sum256(newblock.Serialize())
	if !ecdsa.Verify(&pubKeyOrigin, _hash[:], &r, &s) {
		return false
	}
	newblock.Signature = _sig
	newblock.Hash = _hash[:]
	return true
}

func (bc *BlockChain) GetTailBlockHash() []byte {
	var hashkey []byte
	bc.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bc.BlockBucket))
		if bucket == nil {
			panic("bucket is empty!")
		}
		hashkey = bucket.Get([]byte(bc.BlockTailHashKey))
		return nil
	})
	return hashkey
}

func (bc *BlockChain) GetTailUserBlockHash() []byte {
	var hashkey []byte
	bc.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bc.BlockBucket))
		if bucket == nil {
			panic("bucket is empty!")
		}
		hashkey = bucket.Get([]byte(bc.TailUserBlockHashkey))
		return nil
	})
	return hashkey
}
func (bc *BlockChain) GetBlockByHash(hash []byte) (*Block, error) {
	block := Block{BlockId: 0}
	if len(hash) == 0 {
		return nil, errors.New("错误的hash")

	}
	err := bc.Db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bc.BlockBucket))
		if bucket == nil {
			return errors.New("not found bucket")
		}
		value := bucket.Get(hash)
		if len(value) == 0 {
			return errors.New(fmt.Sprintf(" not the key"))
		}
		block = BlockDeserialize(value)
		return nil
	})
	return &block, err
}
func (bc *BlockChain) GetAddressFromUsername(username string) (string, error) {
	rw := LocalWallets.GetBlockChainRootWallet()
	if rw == nil {
		return "", errors.New("root用户未创建")
	}
	user_address := rw.NewAddress()

	// 判断用户是否创建
	_hash, _ := LocalWallets.GetUserTailBlockHash(user_address)

	b, e := bc.GetBlockByHash(_hash)
	if e != nil {
		return "", e
	}
	for {
		if b.IsGenesisBlock() {
			break
		}
		for _, tx := range b.TxInfos {
			_hash = tx.PreBlockHash
			if tx.Key == username {
				addr := strings.Split(string(tx.Value), " ")[1]
				return addr, nil
			}
		}
		b, _ = bc.GetBlockByHash(_hash)
	}
	return "", errors.New("未知的用户")
}
