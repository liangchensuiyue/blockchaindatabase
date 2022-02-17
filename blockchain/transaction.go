package blockchain

import (
	// "github.com/btcsuite/btcutil/base58"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"
	// "log"
	// "go_code/区块链/demo1/block"
)

const reword = 12.5

//1. 定义交易结构
type Transaction struct {
	Key          string
	Value        []byte
	DataType     string
	Timestamp    uint64
	DelMark      bool
	PublicKey    []byte
	Hash         []byte
	Share        bool
	ShareAddress []string
	// 当交易打包时在填上
	PreBlockHash []byte
	Signature    []byte
}

// 设置交易ID
// func (tx *Transaction) SetHash() {
// 	var buffer bytes.Buffer
// 	encoder := gob.NewEncoder(&buffer)
// 	err := encoder.Encode(tx)
// 	if err != nil {
// 		panic(err)
// 	}
// 	data := buffer.Bytes()
// 	// hash := sha256.Sum256(data)
// 	tx.Hash = data
// }

func (tx *Transaction) SetHash() {
	data := []byte{}
	data = append(data, []byte(tx.Key)...)
	data = append(data, tx.Value...)
	data = append(data, []byte(tx.DataType)...)
	data = append(data, []byte(fmt.Sprintf("%d", tx.Timestamp))...)
	data = append(data, tx.PublicKey...)
	data = append(data, tx.PreBlockHash...)
	data = append(data, []byte(fmt.Sprintf("%d", tx.Share))...)
	data = append(data, []byte(fmt.Sprintf("%d", tx.DelMark))...)
	for _, addr := range tx.ShareAddress {
		data = append(data, []byte(addr)...)
	}
	hash := sha256.Sum256(data)
	tx.Hash = hash[:]
}

// 创建普通的转账交易
func NewTransaction(method, key string, value []byte, datatype string, user_address string, share bool, shareuser_address []string) (*Transaction, error) {

	var Tx *Transaction

	// 创建交易之后要进行数字签名,所以需要通过地址打开对应钱包获取私钥公钥

	wallet, e := LocalWallets.GetUserWallet(user_address)
	if wallet == nil || e != nil {
		return nil, e
	}

	// pubKey := wallet.PubKey
	// pubKeyHash := HashPubKey(pubKey)
	// privateKey := wallet.Private
	switch method {
	case "put":
		Tx = &Transaction{
			Key:          key,
			Value:        value,
			Share:        share,
			DataType:     datatype,
			Timestamp:    uint64(time.Now().Unix()),
			DelMark:      false,
			PublicKey:    wallet.PubKey,
			ShareAddress: shareuser_address,
		}
	case "del":
		Tx = &Transaction{
			Key:          key,
			Value:        value,
			Share:        share,
			DataType:     datatype,
			Timestamp:    uint64(time.Now().Unix()),
			DelMark:      true,
			PublicKey:    wallet.PubKey,
			ShareAddress: shareuser_address,
		}
	case "create_user":
		Tx = &Transaction{
			Key:       key,
			Value:     value,
			Share:     share,
			DataType:  datatype,
			Timestamp: uint64(time.Now().Unix()),
			DelMark:   false,
			PublicKey: wallet.PubKey,
		}
	default:
		return nil, errors.New(" 未知的操作")
	}

	// hash 在区块打包时建立
	// tx.SetHash()

	// bc.SignTransaction(&tx, privateKey)
	return Tx, nil
}

func (tx *Transaction) Sign() {
	user_address := GenerateAddressFromPubkey(tx.PublicKey)
	var privateKey *ecdsa.PrivateKey
	wa, _ := LocalWallets.GetUserWallet(user_address)
	privateKey = wa.Private
	if tx.Share {
		tx.PreBlockHash, _ = LocalWallets.GetUserTailBlockHash(GenerateUserShareKey(tx.ShareAddress))
	} else {
		tx.PreBlockHash, _ = LocalWallets.GetUserTailBlockHash(user_address)
	}
	tx.Hash = []byte{}
	tx.SetHash()
	// fmt.Println("sign", string(tx.Hash), tx)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, tx.Hash)
	if err != nil {
		log.Panic(err)
	}
	tx.Signature = append(r.Bytes(), s.Bytes()...)
}
func Printx(hash []byte) {
	var buffer *bytes.Buffer = bytes.NewBuffer(hash)
	decore := gob.NewDecoder(bytes.NewReader(buffer.Bytes()))
	var otx Transaction
	err := decore.Decode(&otx)
	if err != nil {
		panic(err)
	}
	fmt.Println(otx.Key)
	fmt.Println(otx.Value)
	fmt.Println(otx.DataType)
	fmt.Println(otx.DelMark)
	fmt.Println("hash", base64.RawStdEncoding.EncodeToString(otx.Hash))
	fmt.Println("pubkey", base64.RawStdEncoding.EncodeToString(otx.PublicKey))
	fmt.Println(otx.Share)
	fmt.Println(otx.ShareAddress)
	fmt.Println(base64.RawStdEncoding.EncodeToString(otx.Signature))
	fmt.Println(otx.Timestamp)
	fmt.Println(base64.RawStdEncoding.EncodeToString(otx.PreBlockHash))
}
func (tx *Transaction) Verify() bool {
	signature := tx.Signature
	tx.Signature = []byte{}
	user_address := GenerateAddressFromPubkey(tx.PublicKey)
	// fmt.Println(user_address)
	if tx.Share {
		tx.PreBlockHash, _ = LocalWallets.GetUserTailBlockHash(GenerateUserShareKey(tx.ShareAddress))
	} else {
		tx.PreBlockHash, _ = LocalWallets.GetUserTailBlockHash(user_address)
	}
	tx.Hash = []byte{}
	tx.SetHash()

	// fmt.Println("verift", string(tx.Hash), tx)

	tx.Signature = signature
	r := big.Int{}
	s := big.Int{}

	r.SetBytes(tx.Signature[:len(tx.Signature)/2])
	s.SetBytes(tx.Signature[len(tx.Signature)/2:])

	X := big.Int{}
	Y := big.Int{}

	X.SetBytes(tx.PublicKey[:len(tx.PublicKey)/2])
	Y.SetBytes(tx.PublicKey[len(tx.PublicKey)/2:])

	pubKeyOrigin := ecdsa.PublicKey{Curve: elliptic.P256(), X: &X, Y: &Y}
	if !ecdsa.Verify(&pubKeyOrigin, tx.Hash, &r, &s) {
		fmt.Println(tx.Key, "校验失败")
		return false
	}
	return true
}
