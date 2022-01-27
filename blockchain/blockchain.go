package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/boltdb/bolt"
)

// 引入区块链
type BlockChain struct {
	db   *bolt.DB
	tail []byte // 存储最后一个区块的hash
}

const blockChainDB = "blockChain.db"
const blockBucket = "blockBucket"

func NewBlockChain(address string) *BlockChain {
	// 创建一个创世块，并作为第一个区块添加到区块链中
	var lastHash []byte
	db, err := bolt.Open("blockChainDB", 0600, nil)
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
			genesisBlock := GenesisBlock(address)
			bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
			bucket.Put([]byte("LastHashKey"), genesisBlock.Hash)
			lastHash = genesisBlock.Hash
		} else {
			lastHash = bucket.Get([]byte("LastHashKey"))
		}
		return nil
	})
	return &BlockChain{
		db,
		lastHash,
	}
}

// 创世块
func GenesisBlock(address string) *Block {
	coinbase := NewCoinbaseTX(address, "go 区块链")
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// 添加区块
func (bc *BlockChain) AddBlock(txs []*Transaction) {
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
func (bc *BlockChain) Print() {
	lasthash := bc.tail
	db := bc.db
	var block Block
	for {
		db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(blockBucket))
			if bucket == nil {
				panic(errors.New("错误的db"))
			}
			value := bucket.Get(lasthash)
			block = Deserialize(value)
			fmt.Println("============================================")
			fmt.Printf("区块版本: %d\n", block.Version)
			fmt.Printf("前区块hash: %x\n", block.PrevHash)
			fmt.Printf("MerkelRoot: %x\n", block.MerkelRoot)
			createtime := (time.Unix(int64(block.TimeStamp), 0)).Format("2006-01-02 15:04:05")
			fmt.Printf("时间: %s\n", createtime)
			fmt.Printf("难度值: %x\n", block.Difficulty)
			fmt.Printf("Nonce: %d\n", block.Nonce)
			fmt.Printf("Hash: %x\n", block.Hash)
			fmt.Printf("数据: %s\n", block.Transactions[0].TXInputs[0].PubKey)
			fmt.Printf("挖矿奖励: %f\n", block.Transactions[0].TXOutputs[0].Value)
			fmt.Println("============================================")
			lasthash = block.PrevHash
			return nil
		})
		if len(block.PrevHash) == 0 {
			break
		}
	}
}
func (bc *BlockChain) Next() (Block, error) {
	hash := bc.tail
	var block Block
	if !bc.HasNext() {
		return Block{}, errors.New("empty!")
	}
	db := bc.db
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		value := bucket.Get(hash)
		block = Deserialize(value)
		bc.tail = block.PrevHash
		return nil
	})
	return block, nil
}
func (bc *BlockChain) HasNext() bool {
	hash := bc.tail
	if len(hash) == 0 {
		return false
	}
	return true
}

// 找到指定地址的所有utxo
func (bc *BlockChain) FindUTXOs(pubKeyHash []byte) []TXOutput {
	var UTXO []TXOutput
	txs := bc.FindUTXOTransactions(pubKeyHash)
	for _, tx := range txs {
		for _, output := range tx.TXOutputs {
			if bytes.Equal(pubKeyHash, output.PubKeyHash) {
				UTXO = append(UTXO, output)
			}
		}
	}
	return UTXO
}

func (bc *BlockChain) FindNeedUTXOs(senderPubKeyHash []byte, amount float64) (map[string][]uint64, float64) {
	utxos := make(map[string][]uint64)
	var calc float64
	txs := bc.FindUTXOTransactions(senderPubKeyHash)
	for _, tx := range txs {
		for i, output := range tx.TXOutputs {
			if bytes.Equal(senderPubKeyHash, output.PubKeyHash) {
				if calc < amount {
					utxos[string(tx.TXID)] = append(utxos[string(tx.TXID)], uint64(i))
					calc += output.Value

					if calc >= amount {
						fmt.Println("找到满足的金额")
						return utxos, calc
					}
				}
			}
		}
	}
	return utxos, calc
}
func (bc *BlockChain) FindUTXOTransactions(senderPubKeyHash []byte) []*Transaction {
	// var UTXO []TXOutput
	var txs []*Transaction
	spentOutputs := make(map[string][]int64)
	it := bc.NewIterator()
	for {
		//1. 遍历区块
		block := it.Next()

		//2. 遍历交易
		for _, tx := range block.Transactions {

			//3. 遍历output，找到和自己相关的utxo(再添加结果之前检查一下是否已经消耗过)
		OUTPUT:
			for i, output := range tx.TXOutputs {

				// 过滤已经消耗的 utxo
				if spentOutputs[string(tx.TXID)] != nil {
					for _, j := range spentOutputs[string(tx.TXID)] {
						if int64(i) == j {
							// 当前的 utxo 已经消耗过了
							continue OUTPUT
						}
					}
				}

				// 找到自己的 utxo
				if bytes.Equal(output.PubKeyHash, senderPubKeyHash) {
					txs = append(txs, tx)
				}
			}

			//4. 遍历input，找到自己花费过的utxo的集合(标识自己消耗过的)
			//存储已经消耗过的utxo
			// map[交易id][]int64
			if !tx.IsCoinbase() {
				for _, input := range tx.TXInputs {
					// 判断交易的输入是不是 给定的address
					if bytes.Equal(HashPubKey(input.PubKey), senderPubKeyHash) {
						indexArray := spentOutputs[string(input.TXid)]
						spentOutputs[string(input.TXid)] = append(indexArray, input.Index)
					}
				}
			}

		}

		if len(block.PrevHash) == 0 {
			// fmt.Println("区块遍历完成，退出!")
			break
		}
	}
	return txs
}

func (bc *BlockChain) FindTransactionByTXid(id []byte) (Transaction, error) {
	it := bc.NewIterator()
	for {
		block := it.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.TXID, id) {
				return *tx, nil
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("不存在的交易")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privateKey *ecdsa.PrivateKey) {
	prevTxs := make(map[string]Transaction)
	for _, input := range tx.TXInputs {
		tx, err := bc.FindTransactionByTXid(input.TXid)
		if err != nil {
			log.Panic(err)
		}
		prevTxs[string(input.TXid)] = tx
	}

	tx.Sign(privateKey, prevTxs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	prevTxs := make(map[string]Transaction)
	for _, input := range tx.TXInputs {
		tx, err := bc.FindTransactionByTXid(input.TXid)
		if err != nil {
			log.Panic(err)
		}
		prevTxs[string(input.TXid)] = tx

	}
	return tx.Verify(prevTxs)
}
