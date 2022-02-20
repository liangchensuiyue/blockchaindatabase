package view

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	BC "go_code/基于区块链的非关系型数据库/blockchain"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/websocket"
	// "bufio"
	// "os"
)

type MyHandler struct {
}
type Message struct {
	Type       string `json:"type"`
	MsgJsonStr string `json:"msg"`
}

var MsgQueue chan Message = make(chan Message, 100)

func Handler(msg Message) {
	MsgQueue <- msg
}

var LBC *BC.BlockChain
var localNodes []string

func ws_handle(conn *websocket.Conn) {
	data, _ := json.Marshal(map[string]interface{}{
		"Nodes": localNodes,
	})
	MsgQueue <- Message{
		Type:       "quorum",
		MsgJsonStr: string(data),
	}
	go func() {
		for {
			data := ""
			e := websocket.Message.Receive(conn, &data)
			blockid, _ := strconv.ParseInt(data, 0, 0)
			// fmt.Println("blll", blockid)
			if e != nil {

				fmt.Println(e.Error())
				conn.Close()
				return
			} else {

				LBC.Traverse(func(block *BC.Block, err error) {

					if block != nil && int64(block.BlockId) > blockid {
						// fmt.Println(block.BlockId)
						infos := []map[string]interface{}{}
						for i, tx := range block.TxInfos {
							addrs := []string{}
							for _, uaddr := range tx.ShareAddress {
								// w, e = BC.LocalWallets.GetUserWallet(uaddr)
								// if e == nil {
								addrs = append(addrs, uaddr)
								// }
							}

							infos = append(infos, map[string]interface{}{
								"Index":            i,
								"UserAddress":      BC.GenerateAddressFromPubkey(tx.PublicKey),
								"Hash":             base64.RawStdEncoding.EncodeToString(tx.Hash),
								"Timestamp":        tx.Timestamp,
								"ShareUserAddress": addrs,
							})
						}
						datastr, _ := json.Marshal(map[string]interface{}{
							"BlockId":       block.BlockId,
							"PrevBlockHash": base64.RawStdEncoding.EncodeToString(block.PreBlockHash),
							"Hash":          base64.RawStdEncoding.EncodeToString(block.Hash),
							"Timestamp":     block.Timestamp,
							"TxInfos":       infos,
						})
						MsgQueue <- Message{
							Type:       "Block",
							MsgJsonStr: string(datastr),
						}

					}

				})
			}
		}
	}()
	for {
		msg := <-MsgQueue
		// fmt.Println(msg)
		time.Sleep(time.Second)
		data, e := json.Marshal(&msg)
		// fmt.Println(e)
		e = websocket.Message.Send(conn, string(data))
		if e != nil {
			// fmt.Println("call", msg)
			MsgQueue <- msg
			break
		}
	}
}

func startWebsocket() {
	http.Handle("/msg", websocket.Handler(ws_handle))

	if err := http.ListenAndServe(":3500", nil); err != nil {
		// mylog.Info("websoket server fail to launch" + err.Error())
		fmt.Println(err.Error())
		// os.Exit(0)
	}
}
func Run(lbc *BC.BlockChain, nodes []string) {
	LBC = lbc
	localNodes = nodes
	// loginhandler := Loginhandler{}
	// registerhandler := Registerhandler{}
	mux := http.NewServeMux()
	go startWebsocket()
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./view"))))
	// http.HandleFunc("/", handler)
	// mux.HandleFunc("/", handler)
	// http.Handle("/user", &myhandler)
	// mux.HandleFunc("/", Loginhandler)

	// http.ListenAndServe(":9090",nil)
	// http.ListenAndServe(":9090",mux)
	http.ListenAndServe(":3400", mux)
}
