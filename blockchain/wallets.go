package blockchain

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"sync"

	"github.com/btcsuite/btcutil/base58"
	// "fmt"
)

const walletFile = "wallet.dat"

var LocalWallets *Wallets
var _lock *sync.Mutex

// 定义一个 Wallets 结构，它保存所有的wallet以及它的地址
type Wallets struct {
	WalletsMap map[string]*Wallet
}

func LoadLocalWallets() {
	LocalWallets.loadFile()
}

func (ws *Wallets) GetUserWallet(user_address string) (*Wallet, error) {
	LoadLocalWallets()
	wa, flag := ws.WalletsMap[user_address]
	if !flag {
		return wa, errors.New("未知的用户")
	}
	return wa, nil
}
func (ws *Wallets) GetAddressFromUsername(username string) (string, error) {
	for addr, v := range ws.WalletsMap {
		if v.Username == username {
			return addr, nil
		}
	}
	return "", errors.New("未知用户")
}
func (ws *Wallets) GetUserTailBlockHash(user_address string) ([]byte, error) {
	LoadLocalWallets()
	wa, flag := ws.WalletsMap[user_address]
	if !flag {
		return []byte{}, errors.New("未知的用户")
	}
	return wa.TailBlockHash, nil
}
func (ws *Wallets) PutTailBlockHash(user_address string, blockhash []byte) error {
	LoadLocalWallets()
	_, flag := ws.WalletsMap[user_address]
	if !flag {
		return errors.New("未知的用户")
	}
	ws.WalletsMap[user_address].TailBlockHash = blockhash
	ws.SaveToFile()
	return nil
}
func (ws *Wallets) loadFile() {
	_, err := os.Stat(walletFile)
	if os.IsNotExist(err) {
		ws.WalletsMap = make(map[string]*Wallet)
		ws.SaveToFile()
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

func (ws *Wallets) SaveToFile() {
	defer _lock.Unlock()
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
	_lock.Lock()
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
