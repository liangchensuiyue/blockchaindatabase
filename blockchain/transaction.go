package blockchain

import (
	// "github.com/btcsuite/btcutil/base58"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"

	// "log"
	// "go_code/区块链/demo1/block"
	"bytes"
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
	Hash         string
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
	tx.TXID = hash[:]
}

// 判断是否为挖矿交易
func (tx *Transaction) IsCoinbase() bool {
	// 交易的input只有一个
	// 交易id为空
	// 交易的index为-1
	input := tx.TXInputs[0]
	if len(tx.TXInputs) == 1 && len(input.TXid) == 0 && input.Index == -1 {
		return true
	}
	return false
}

//2. 提供创建交易方法(挖矿交易)
func NewCoinbaseTX(address string, data string) *Transaction {
	//	挖矿交易的特点
	//1. 只有一个input
	//2. 无需引用交易id
	//3. 无需引用index
	// 矿工由于挖矿时无需指定签名，所以这个PubKey字段可以由矿工自由填写数据，一般是填写矿池的名字
	input := TXInput{[]byte{}, -1, nil, []byte(data)}
	// output := TXOutput{reword, address}
	output := NewTXOutput(reword, address)
	tx := Transaction{[]byte{}, []TXInput{input}, []TXOutput{*output}}
	tx.SetHash()
	return &tx
}

// 创建普通的转账交易
func NewTransaction(key string, value []byte, user string, sharemode string, shareuser []string) *Transaction {
	// 创建交易之后要进行数字签名,所以需要通过地址打开对应钱包获取私钥公钥
	ws := NewWallets()
	wallet := ws.WalletsMap[user]
	if wallet == nil {
		fmt.Println("没有找到对应钱包!")
		return nil
	}

	pubKey := wallet.PubKey
	pubKeyHash := HashPubKey(pubKey)
	privateKey := wallet.Private

	utxos, resValue := bc.FindNeedUTXOs(pubKeyHash, amount)
	if resValue < amount {
		fmt.Println("余额不足，交易失败!", resValue, amount)
		return nil
	}
	var inputs []TXInput
	var outputs []TXOutput

	// 将这些 utxo 转换成 input
	for id, indexArray := range utxos {
		for _, i := range indexArray {
			input := TXInput{[]byte(id), int64(i), nil, pubKey}
			inputs = append(inputs, input)
		}
	}
	// output := TXOutput{amount, to}
	output := NewTXOutput(amount, to)
	outputs = append(outputs, *output)
	if resValue > amount {
		// 找零
		// outputs = append(outputs, TXOutput{resValue - amount, from})
		outputs = append(outputs, *(NewTXOutput(resValue-amount, from)))
	}
	tx := Transaction{[]byte{}, inputs, outputs}
	tx.SetHash()

	bc.SignTransaction(&tx, privateKey)
	return &tx
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
