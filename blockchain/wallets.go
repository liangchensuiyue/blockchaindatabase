package blockchain

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"

	"github.com/btcsuite/btcutil/base58"
	// "fmt"
)

const walletFile = "wallet.dat"

// 定义一个 Wallets 结构，它保存所有的wallet以及它的地址
type Wallets struct {
	WalletsMap map[string]*Wallet
}

func NewWallets() *Wallets {
	var ws Wallets
	ws.loadFile()
	return &ws
}

func (ws *Wallets) loadFile() {
	_, err := os.Stat(walletFile)
	if os.IsNotExist(err) {
		ws.WalletsMap = make(map[string]*Wallet)
		return
	}
	// 读取钱包
	content, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	// 解码
	gob.Register(elliptic.P256())

	decoder := gob.NewDecoder(bytes.NewReader(content))

	var wsLocal Wallets
	err = decoder.Decode(&wsLocal)
	if err != nil {
		log.Panic(err)
	}
	// ws = &wsLocal
	ws.WalletsMap = wsLocal.WalletsMap
}
func (ws *Wallets) GetAllAddresses() []string {
	var addresses []string
	for address := range ws.WalletsMap {
		addresses = append(addresses, address)
	}
	return addresses
}

func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := wallet.NewAddress()

	ws.WalletsMap[address] = wallet

	ws.saveToFile()
	return address
}
func (ws *Wallets) saveToFile() {
	/*
		如果 Encode/Decode 类型是interface或者struct中某些字段是interface{}的时候
		需要在gob中注册interface可能的所有实现或者可能类型
	*/
	var content bytes.Buffer

	// Curve 是一个接口类型
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

func GetPubKeyFromAddress(address string) []byte {
	// 1. 解码
	// 2. 截取出公钥哈希，取出version(1 Byte), 去除校验码(4 Byte)
	addressByte := base58.Decode(address)
	length := len(addressByte)
	pubKeyHash := addressByte[1 : length-4]
	return pubKeyHash
}
