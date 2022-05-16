package quorum

import (
	"context"
	"fmt"
	db "go_code/基于区块链的非关系型数据库/database"
	"go_code/基于区块链的非关系型数据库/linkpool"
	ucgrpc "go_code/基于区块链的非关系型数据库/proto/userclient"
	"go_code/基于区块链的非关系型数据库/util"
	"net"
	"time"

	"google.golang.org/grpc"
)

type Server struct{}

func (this *Server) Newuser(ctx context.Context, in *ucgrpc.UserInfo) (*ucgrpc.VerifyInfo, error) {
	err := db.CreateUser(in.Username, in.Passworld)
	if err != nil {
		return nil, err
	}
	return &ucgrpc.VerifyInfo{Status: true}, nil
}
func (this *Server) Get(in ucgrpc.UserClientService_GetServer) (err error) {
	isclose := false
	req, _ := in.Recv()
	err = db.VeriftUser(req.Username, req.Passworld)
	if err != nil {
		fmt.Println("verifey", err)
		return err
	}
	useraddress, _e := db.GetAddressFromUsername(req.Username)
	if _e != nil {
		fmt.Println("GetAddressFromUsername", _e)
		return _e
	}
	linkpool.Global_Link_pool.AddNode("", func() {
		isclose = true
	})

	if !req.Sharemode {
		req.ShareChan = ""
	}
	block, index := db.Get(req.Key, req.Username, useraddress, req.Sharemode, req.ShareChan)
	if index == -1 {
		in.Send(&ucgrpc.ResQuery{
			Status: false,
		})
	} else {
		in.Send(&ucgrpc.ResQuery{
			Status: true,
			Data:   block.TxInfos[index].Value,
		})
		// info.Status = true
		// info.Data = block.TxInfos[index].Value
	}
	go func() {
		for req, _ = in.Recv(); req != nil; req, _ = in.Recv() {
			if isclose {
				break
			}
			if !req.Sharemode {
				req.ShareChan = ""
			}
			block, index := db.Get(req.Key, req.Username, useraddress, req.Sharemode, req.ShareChan)
			if index == -1 {
				in.Send(&ucgrpc.ResQuery{
					Status: false,
					// Data:   block.TxInfos[index].Value,
				})
			} else {
				in.Send(&ucgrpc.ResQuery{
					Status: true,
					Data:   block.TxInfos[index].Value,
				})
				// info.Status = true
				// info.Data = block.TxInfos[index].Value
			}
		}
	}()
	for {
		time.Sleep(time.Second * 2)
		if isclose {
			break
		}
	}
	return
}
func (this *Server) Put(in ucgrpc.UserClientService_PutServer) (err error) {
	info := &ucgrpc.VerifyInfo{}
	req, err := in.Recv()
	err = db.VeriftUser(req.Username, req.Passworld)
	if err != nil {
		in.SendAndClose(info)
		return err
	}
	useraddress, _e := db.GetAddressFromUsername(req.Username)
	if _e != nil {
		fmt.Println("close ......................")
		in.SendAndClose(info)
		return _e
	}
	linkpool.Global_Link_pool.AddNode("", func() {
		// in.SendAndClose(info)
	})
	N := 1
	if !req.Share {
		v := req.Value

		key := util.Yield16ByteKey([]byte(req.Passworld))
		v = util.AesEncrypt(v, key)
		db.Put(req.Key, v, req.Datatype, useraddress, req.Share, "", req.Strict)

	} else {

		db.Put(req.Key, req.Value, req.Datatype, useraddress, req.Share, req.ShareChan, req.Strict)

	}
	for req, _ = in.Recv(); req != nil; req, _ = in.Recv() {
		N++
		v := req.Value
		if !req.Share {
			key := util.Yield16ByteKey([]byte(req.Passworld))
			v = util.AesEncrypt(v, key)

			db.Put(req.Key, v, req.Datatype, useraddress, req.Share, "", req.Strict)

		} else {

			db.Put(req.Key, req.Value, req.Datatype, useraddress, req.Share, req.ShareChan, req.Strict)

		}
	}
	// fmt.Println("N:", N)
	return
}
func Run() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", 3600))
	if err != nil {
		fmt.Println("网络错误", err)
	}

	//创建grpc的服务
	srv := grpc.NewServer()

	//注册服务
	ucgrpc.RegisterUserClientServiceServer(srv, &Server{})

	//等待网络连接
	err = srv.Serve(ln)
	if err != nil {
		fmt.Println("网络错误", err)
	}
}
