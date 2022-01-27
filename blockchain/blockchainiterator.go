package blockchain

import (
	"github.com/boltdb/bolt"
	// 	"log"
	// 	"errors"
	// 	"fmt"
)

type BlockChainIterator struct {
	db                 *bolt.DB
	currentHashPointer []byte
}

func (bc *BlockChain) NewIterator() *BlockChainIterator {
	return &BlockChainIterator{
		bc.db,
		bc.tail,
	}
}
func (iterator *BlockChainIterator) Next() *Block {
	db := iterator.db
	block := Block{}
	hash := iterator.currentHashPointer
	if len(hash) == 0 {
		return nil

	}
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			panic("bucket is empty!")
		}
		value := bucket.Get(hash)
		block = Deserialize(value)
		iterator.currentHashPointer = block.PrevHash
		return nil
	})
	return &block
}
