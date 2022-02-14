package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

type Wallet struct {
	Private   *ecdsa.PrivateKey
	Username  string
	Passworld string
	PubKey    []byte

	// 用户最后一个操作所在的区块hash
	TailBlockHash []byte
}

// 创建钱包
func NewWallet(username string, passworld string) *Wallet {
	curve := elliptic.P256()

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic()
	}
	pubkeyOrig := privateKey.PublicKey
	pubKey := append(pubkeyOrig.X.Bytes(), pubkeyOrig.Y.Bytes()...)
	return &Wallet{Private: privateKey, PubKey: pubKey, Username: username, Passworld: passworld}
}

// 生成地址
func (w *Wallet) NewAddress() string {
	pubKey := w.PubKey

	rip160HashValue := HashPubKey(pubKey)

	version := byte(00)
	payload := append([]byte{version}, rip160HashValue...)

	checkCode := CheckSum(payload)
	payload = append(payload, checkCode...)

	// go 语言有一个库 btcd 这个是go语言实现的比特币全节点源码
	address := base58.Encode(payload)
	return address
}

func GenerateAddressFromPubkey(pubKey []byte) string {
	rip160HashValue := HashPubKey(pubKey)

	version := byte(00)
	payload := append([]byte{version}, rip160HashValue...)

	checkCode := CheckSum(payload)
	payload = append(payload, checkCode...)

	// go 语言有一个库 btcd 这个是go语言实现的比特币全节点源码
	address := base58.Encode(payload)
	return address
}

func HashPubKey(data []byte) []byte {
	hash := sha256.Sum256(data)
	rip160hasher := ripemd160.New()
	_, err := rip160hasher.Write(hash[:])
	if err != nil {
		log.Panic(err)
	}
	rip160HashValue := rip160hasher.Sum(nil)
	return rip160HashValue
}

func CheckSum(data []byte) []byte {
	hash1 := sha256.Sum256(data)
	hash2 := sha256.Sum256(hash1[:])

	checkCode := hash2[:4]
	return checkCode

}

func IsValidAddress(address string) bool {
	addressByte := base58.Decode(address)
	if len(addressByte) < 4 {
		return false
	}
	payload := addressByte[:len(addressByte)-4]
	checksum1 := addressByte[len(addressByte)-4:]
	checksum2 := CheckSum(payload)
	return bytes.Equal(checksum1, checksum2)
}
