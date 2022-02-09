package blockchain

import (
	// "github.com/btcsuite/btcutil/base58"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"log"
	"math/big"
	"time"

	// "log"
	// "go_code/区块链/demo1/block"
	"bytes"
)

const reword = 12.5

//1. 定义交易结构
type Transaction struct {
	Key       string
	Value     []byte
	DataType  string
	Timestamp uint64
	DelMark   bool
	PublicKey []byte
	Hash      []byte

	// 当交易打包时在填上
	PreBlockHash string
	Signature    string
}

// 设置交易ID
func (tx *Transaction) SetHash() {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	if err != nil {
		panic(err)
	}
	data := buffer.Bytes()
	hash := sha256.Sum256(data)
	tx.Hash = hash[:]
}

// 创建普通的转账交易
func NewTransaction(method, key string, value []byte, datatype string, user_address string, sharemode string, shareuser []string) (*Transaction, error) {

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
			Key:       key,
			Value:     value,
			DataType:  datatype,
			Timestamp: uint64(time.Now().Unix()),
			DelMark:   false,
			PublicKey: wallet.PubKey,
		}
	case "delete":
		Tx = &Transaction{
			Key:       key,
			Value:     value,
			DataType:  datatype,
			Timestamp: uint64(time.Now().Unix()),
			DelMark:   true,
			PublicKey: wallet.PubKey,
		}
	default:
		return nil, errors.New("未知的操作")
	}

	// hash 在区块打包时建立
	// tx.SetHash()

	// bc.SignTransaction(&tx, privateKey)
	return Tx, nil
}

func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey, prevTxs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}
	txCopy := tx.TrimmedCopy()
	for i, input := range txCopy.TXInputs {
		prevtx := prevTxs[string(input.TXid)]
		if len(tx.TXID) == 0 {
			log.Panic("引用交易失败!")
		}
		txCopy.TXInputs[i].PubKey = prevtx.TXOutputs[input.Index].PubKeyHash
		txCopy.SetHash()
		signDataHash := txCopy.TXID
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, signDataHash)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.TXInputs[i].Signature = signature
	}
}
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	for _, input := range tx.TXInputs {
		inputs = append(inputs, TXInput{input.TXid, input.Index, nil, nil})
	}
	for _, output := range tx.TXOutputs {
		outputs = append(outputs, output)
	}
	return Transaction{tx.TXID, inputs, outputs}
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	txCopy := tx.TrimmedCopy()
	for i, input := range tx.TXInputs {
		prevTX := prevTXs[string(input.TXid)]
		if len(prevTX.TXID) == 0 {
			log.Panic("引用的交易无效!")
		}
		txCopy.TXInputs[i].PubKey = prevTX.TXOutputs[input.Index].PubKeyHash
		txCopy.SetHash()
		dataHash := txCopy.TXID
		signature := input.Signature
		pubKey := input.PubKey

		r := big.Int{}
		s := big.Int{}

		r.SetBytes(signature[:len(signature)/2])
		s.SetBytes(signature[len(signature)/2:])

		X := big.Int{}
		Y := big.Int{}

		X.SetBytes(pubKey[:len(pubKey)/2])
		Y.SetBytes(pubKey[len(pubKey)/2:])

		pubKeyOrigin := ecdsa.PublicKey{Curve: elliptic.P256(), X: &X, Y: &Y}
		if !ecdsa.Verify(&pubKeyOrigin, dataHash, &r, &s) {
			return false
		}
	}
	return true
}
