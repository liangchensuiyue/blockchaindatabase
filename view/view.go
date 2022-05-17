package view

import (
	"encoding/json"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	"go_code/基于区块链的非关系型数据库/quorum"
	"net/http"
	"strconv"
	// "bufio"
	// "os"
)

type MyHandler struct {
}
type Message struct {
	Type       string `json:"type"`
	MsgJsonStr string `json:"msg"`
}
type Quorum struct {
	Ip string
}
type BlockInfo struct {
	BlockId   uint64
	Hash      string
	PrevHash  string
	Timestamp uint64
	Txnums    uint64
}
type UserInfo struct {
	Username        string
	UserAddress     string
	UserTxNums      uint64
	UserRelateBlock []BlockInfo
}
type PipeInfo struct {
	Pipename        string
	PipeTxNums      uint64
	PipeRelateBlock []BlockInfo
	PipeUser        []UserInfo
}

type NodeInfo struct {
	Ip                 string
	LatestblockId      uint64
	SystxRateInfo      string
	SysblockRateInfo   string
	Isaccountant       bool
	LocalWalletAddress []string
}
type ViewHander struct {
	GetQuorum    func() []Quorum
	GetNodeInfo  func() NodeInfo
	GetPipeInfo  func(string) PipeInfo
	GetUserInfo  func(string) UserInfo
	GetBlockInfo func(uint64) BlockInfo
	GetAllBlock  func() []BlockInfo
}

var GlobalViewHandler ViewHander

var LBC *BC.BlockChain
var LocalNodes *quorum.BlockChainNode

func GetQuorum(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	qus := GlobalViewHandler.GetQuorum()
	data, e := json.Marshal(map[string]interface{}{
		"Quorums": qus,
	})
	if e != nil {
		w.Write([]byte("{Quorums:[]}"))
	} else {
		w.Write(data)
	}
}
func GetNodeInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	info := GlobalViewHandler.GetNodeInfo()
	data, e := json.Marshal(info)
	if e != nil {
		w.Write([]byte("{}"))
	} else {
		w.Write(data)
	}
}
func GetPipeInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	info := GlobalViewHandler.GetPipeInfo(r.FormValue("pname"))
	data, e := json.Marshal(info)
	if e != nil {
		w.Write([]byte("{}"))
	} else {
		w.Write(data)
	}
}
func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	info := GlobalViewHandler.GetUserInfo(r.FormValue("uname"))
	data, e := json.Marshal(info)
	if e != nil {
		w.Write([]byte("{}"))
	} else {
		w.Write(data)
	}
}
func GetBlockInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	id, _ := strconv.Atoi(r.FormValue("id"))
	info := GlobalViewHandler.GetBlockInfo(uint64(id))
	data, e := json.Marshal(info)
	if e != nil {
		w.Write([]byte("{}"))
	} else {
		w.Write(data)
	}
}
func GetAllBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	info := GlobalViewHandler.GetAllBlock()
	data, e := json.Marshal(map[string]interface{}{
		"Blocks": info,
	})
	if e != nil {
		w.Write([]byte("{\"Blocks\":[]}"))
	} else {
		w.Write(data)
	}
}
func Run(lbc *BC.BlockChain, lnd *quorum.BlockChainNode) {
	LBC = lbc
	LocalNodes = lnd
	// loginhandler := Loginhandler{}
	// registerhandler := Registerhandler{}
	mux := http.NewServeMux()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./view"))))
	mux.HandleFunc("/GetQuorum", GetQuorum)
	mux.HandleFunc("/GetNodeInfo", GetNodeInfo)
	mux.HandleFunc("/GetPipeInfo", GetPipeInfo)
	mux.HandleFunc("/GetUserInfo", GetUserInfo)
	mux.HandleFunc("/GetBlockInfo", GetBlockInfo)
	mux.HandleFunc("/GetAllBlock", GetAllBlock)

	// http.HandleFunc("/", handler)
	// mux.HandleFunc("/", handler)
	// http.Handle("/user", &myhandler)
	// mux.HandleFunc("/", Loginhandler)

	// http.ListenAndServe(":9090",nil)
	// http.ListenAndServe(":9090",mux)
	http.ListenAndServe(":3400", mux)
}
