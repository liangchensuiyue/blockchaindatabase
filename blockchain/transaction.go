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
	Type "go_code/基于区块链的非关系型数据库/type"
	"go_code/基于区块链的非关系型数据库/util"
	"log"
	"math/big"
	"strings"
	"time"
	// "log"
	// "go_code/区块链/demo1/block"
)

// const reword = 12.5

//

//1. 定义交易结构
type Transaction struct {
	Key       string
	Value     []byte
	DataType  int32 // newchan join_chan del_chan exit_chan
	Timestamp uint64
	PublicKey []byte
	Hash      []byte
	Share     bool
	ShareChan string
	DsValue   interface{}
	// ShareAddress []string
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

// 生成交易的hash
func (tx *Transaction) SetHash() {
	data := []byte{}
	data = append(data, []byte(tx.Key)...)
	data = append(data, tx.Value...)
	data = append(data, util.Int32ToBytes(tx.DataType)...)
	data = append(data, []byte(tx.ShareChan)...)
	data = append(data, []byte(fmt.Sprintf("%d", tx.Timestamp))...)
	data = append(data, tx.PublicKey...)
	data = append(data, tx.PreBlockHash...)
	data = append(data, []byte(fmt.Sprintf("%d", tx.Share))...)
	// for _, addr := range tx.ShareAddress {
	// 	data = append(data, []byte(addr)...)
	// }
	hash := sha256.Sum256(data)
	tx.Hash = hash[:]
}

// 创建普通的转账交易
func NewTransaction(key string, value []byte, datatype int32, user_address string, share bool, ShareChanName string) (*Transaction, error) {

	var Tx *Transaction

	// 创建交易之后要进行数字签名,所以需要通过地址打开对应钱包获取私钥公钥

	wallet, e := LocalWallets.GetUserWallet(user_address)
	if wallet == nil || e != nil {
		return nil, e
	}
	Tx = &Transaction{
		Key:       key,
		Value:     value,
		Share:     share,
		DataType:  datatype,
		Timestamp: uint64(time.Now().Unix()),
		PublicKey: wallet.PubKey,
		ShareChan: ShareChanName,
	}
	return Tx, nil
}

// 对交易签名
func (tx *Transaction) Sign() {
	user_address := GenerateAddressFromPubkey(tx.PublicKey)
	var privateKey *ecdsa.PrivateKey
	wa, _ := LocalWallets.GetUserWallet(user_address)
	privateKey = wa.Private
	if tx.Share {
		if !LocalWallets.HasShareChan(tx.ShareChan) {
			_GetShareChan(tx.ShareChan)
		}
		schn := LocalWallets.ShareChanMap[tx.ShareChan]
		tx.PreBlockHash = schn.TailBlockHash
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
	fmt.Println("hash", base64.RawStdEncoding.EncodeToString(otx.Hash))
	fmt.Println("pubkey", base64.RawStdEncoding.EncodeToString(otx.PublicKey))
	fmt.Println(otx.Share)
	fmt.Println(otx.ShareChan)
	fmt.Println(base64.RawStdEncoding.EncodeToString(otx.Signature))
	fmt.Println(otx.Timestamp)
	fmt.Println(base64.RawStdEncoding.EncodeToString(otx.PreBlockHash))
}
func GetAddressFromUsername(username string) (string, error) {
	user_address := LocalWallets.GetBlockChainRootWallet().NewAddress()

	// 判断用户是否创建
	_hash, _ := LocalWallets.GetUserTailBlockHash(user_address)

	b, e := _localblockchain.GetBlockByHash(_hash)
	if e != nil {
		return "", e
	}
	for {
		if b.IsGenesisBlock() {
			break
		}
		for _, tx := range b.TxInfos {
			// fmt.Println("t.Key", tx.Key, username)
			_hash = tx.PreBlockHash
			if tx.Key == username {
				addr := strings.Split(string(tx.Value), " ")[1]
				return addr, nil
			}
		}
		b, _ = _localblockchain.GetBlockByHash(_hash)
	}
	return "", errors.New("未知的用户")
}

// 验证用户加入管道的密钥
func VerifyKeyChan(key []byte, channame string) bool {
	flag := true
	var okey []byte
	_localblockchain.Traverse(func(block *Block, err error) bool {
		for i := len(block.TxInfos) - 1; i >= 0; i-- {
			tx := block.TxInfos[i]
			//
			if tx.Key == channame {
				if tx.DataType == Type.NEW_CHAN {
					flag = false
					okey = tx.Value
					return false
				} else if tx.DataType == Type.DEL_CHAN {
					return false
				}
			}

		}
		return true
	})
	if !flag && bytes.Equal(key, okey) {
		// 找到

		return true
	}
	fmt.Println(key, okey)
	return false
}

// 判断用户是否在管道内
func UserIsInChan(addr, creator, channame string) bool {
	retrue := false
	refalse := false
	creatoraddr, err := GetAddressFromUsername(creator)
	if err != nil {
		return false
	}
	if creatoraddr == addr {
		return true
	}
	_localblockchain.Traverse(func(block *Block, err error) bool {
		for i := len(block.TxInfos) - 1; i >= 0; i-- {
			tx := block.TxInfos[i]
			//
			if tx.Key == creator+"."+channame {
				// arr := strings.Split(string(tx.Value), " ")
				if tx.DataType == Type.DEL_CHAN && GenerateAddressFromPubkey(tx.PublicKey) == creatoraddr {
					refalse = true
					return false
				}
				if tx.DataType == Type.EXIT_CHAN && GenerateAddressFromPubkey(tx.PublicKey) == addr {
					refalse = true
					return false
				}
				if tx.DataType == Type.JOIN_CHAN && GenerateAddressFromPubkey(tx.PublicKey) == addr {
					retrue = true
					return false
				}
				if tx.DataType == Type.NEW_CHAN && GenerateAddressFromPubkey(tx.PublicKey) == creatoraddr {
					refalse = true
					return false
				}

			}

		}
		return true
	})
	if refalse {
		return false
	}
	if retrue {
		return true
	}
	return false
}

//  判断用户是否是管道的创建者
func UserIsChanCreator(channame, useraddress string) bool {
	flag := true
	_localblockchain.Traverse(func(block *Block, err error) bool {
		for i := len(block.TxInfos) - 1; i >= 0; i-- {
			tx := block.TxInfos[i]
			//
			if tx.Key == channame {
				if useraddress == GenerateAddressFromPubkey(tx.PublicKey) {
					if tx.DataType == Type.DEL_CHAN {
						return false
					}
					if tx.DataType == Type.NEW_CHAN {
						flag = false
						return false
					}
				}

			}

		}
		return true
	})
	if !flag {
		return true
	}
	return false
}

// 判断是否存在管道
func IsExsistChan(name string, address string) bool {
	flag := true
	_localblockchain.Traverse(func(block *Block, err error) bool {
		for i := len(block.TxInfos) - 1; i >= 0; i-- {
			tx := block.TxInfos[i]
			//
			if tx.Key == name && GenerateAddressFromPubkey(tx.PublicKey) == address {
				if tx.DataType == Type.NEW_CHAN {
					flag = false
					return false
				} else if tx.DataType == Type.DEL_CHAN {
					return false
				}
			}

		}
		return true
	})
	if !flag {
		// 找到
		return true
	}
	return false
}

// 对交易进行校验
func (tx *Transaction) VerifySimple() bool {
	user_address := GenerateAddressFromPubkey(tx.PublicKey)
	rw := LocalWallets.GetBlockChainRootWallet()
	switch tx.DataType {
	case Type.NEW_CHAN:
		if IsExsistChan(tx.Key, GenerateAddressFromPubkey(tx.PublicKey)) {
			return false
		}
	case Type.DEL_CHAN:

		if !UserIsChanCreator(tx.Key, GenerateAddressFromPubkey(tx.PublicKey)) {
			return false
		}
	case Type.JOIN_CHAN:
		arr := strings.Split(string(tx.Value), " ")
		if len(arr) <= 1 {
			fmt.Println("del_chan verify error")
			return false
		}
		joinkey := arr[1]
		arr = strings.Split(tx.Key, ".")
		if len(arr) < 2 {
			return false
		}
		// caddr, err := GetAddressFromUsername(arr[0])
		// if err != nil {
		// 	fmt.Println(err)
		// 	return false
		// }
		ok := VerifyKeyChan([]byte(joinkey), tx.Key)
		if !ok {
			fmt.Println("!ok error")
			return false
		}
	case Type.EXIT_CHAN:
		if UserIsChanCreator(tx.Key, user_address) {
			fmt.Println("用户不能退出自己的chan")
			return false
		}
	case Type.NEW_USER:
		_, err := GetAddressFromUsername(tx.Key)
		if err == nil {
			fmt.Println("new_user error")
			return false
		}
		if !bytes.Equal(rw.PubKey, tx.PublicKey) {
			return false
		}

	}
	return true
}

// 对交易进行校验
func (tx *Transaction) Verify() bool {
	signature := tx.Signature
	tx.Signature = []byte{}
	user_address := GenerateAddressFromPubkey(tx.PublicKey)
	// pre := base64.RawStdEncoding.EncodeToString(tx.PreBlockHash)
	// rw := LocalWallets.GetBlockChainRootWallet()
	if !tx.VerifySimple() {
		return false
	}
	// switch tx.DataType {
	// case Type.NEW_CHAN:
	// 	if IsExsistChan(tx.Key, GenerateAddressFromPubkey(tx.PublicKey)) {
	// 		return false
	// 	}
	// case Type.DEL_CHAN:

	// 	if !UserIsChanCreator(tx.Key, GenerateAddressFromPubkey(tx.PublicKey)) {
	// 		return false
	// 	}
	// case Type.JOIN_CHAN:
	// 	arr := strings.Split(string(tx.Value), " ")
	// 	if len(arr) <= 1 {
	// 		fmt.Println("del_chan verify error")
	// 		return false
	// 	}
	// 	joinkey := arr[1]
	// 	arr = strings.Split(tx.Key, ".")
	// 	if len(arr) < 2 {
	// 		return false
	// 	}
	// 	// caddr, err := GetAddressFromUsername(arr[0])
	// 	// if err != nil {
	// 	// 	fmt.Println(err)
	// 	// 	return false
	// 	// }
	// 	ok := VerifyKeyChan([]byte(joinkey), tx.Key)
	// 	if !ok {
	// 		fmt.Println("!ok error")
	// 		return false
	// 	}
	// case Type.EXIT_CHAN:
	// 	if UserIsChanCreator(tx.Key, user_address) {
	// 		fmt.Println("用户不能退出自己的chan")
	// 		return false
	// 	}
	// case Type.NEW_USER:
	// 	_, err := GetAddressFromUsername(tx.Key)
	// 	if err == nil {
	// 		fmt.Println("new_user error")
	// 		return false
	// 	}
	// 	if !bytes.Equal(rw.PubKey, tx.PublicKey) {
	// 		return false
	// 	}

	// }
	// fmt.Println(user_address)
	if tx.Share {
		//tx.ShareChan   creatorusername.channame
		arr := strings.Split(tx.ShareChan, ".")
		if len(arr) < 2 {

			return false
		}

		if !UserIsInChan(GenerateAddressFromPubkey(tx.PublicKey), arr[0], arr[1]) {
			return false
		}
		if !LocalWallets.HasShareChan(tx.ShareChan) {
			_GetShareChan(tx.ShareChan)
		}
		schn, ok := LocalWallets.ShareChanMap[tx.ShareChan]
		if !ok {
			return false
		}
		tx.PreBlockHash = schn.TailBlockHash
	} else {
		tx.PreBlockHash, _ = LocalWallets.GetUserTailBlockHash(user_address)
	}
	// cur := base64.RawStdEncoding.EncodeToString(tx.PreBlockHash)
	preh := base64.RawStdEncoding.EncodeToString(tx.Hash)
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
		// fmt.Println(tx.Key, "校验失败")
		// fmt.Println(pre, cur)
		// fmt.Println(preh, base64.RawStdEncoding.EncodeToString(tx.Hash))
		if preh == base64.RawStdEncoding.EncodeToString(tx.Hash) {
			return true
		}
		return false
	}
	return true
}
