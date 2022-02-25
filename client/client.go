package quorum

import (
	"context"
	"fmt"
	db "go_code/基于区块链的非关系型数据库/database"
	ucgrpc "go_code/基于区块链的非关系型数据库/proto/userclient"
	"go_code/基于区块链的非关系型数据库/util"
	"net"

	"google.golang.org/grpc"
)

type Server struct{}

func (this *Server) Get(ctx context.Context, req *ucgrpc.GetBody) (info *ucgrpc.ResQuery, err error) {
	info = &ucgrpc.ResQuery{}
	info.Status = false
	if !req.Sharemode {
		req.ShareChan = ""
	}
	block, index := db.Get(req.Key, req.Username, req.UserAddress, req.Sharemode, req.ShareChan)
	if index == -1 {
		return
	} else {

		info.Status = true
		info.Data = block.TxInfos[index].Value
	}
	return
}
func (this *Server) Put(ctx context.Context, req *ucgrpc.PutBody) (info *ucgrpc.VerifyInfo, err error) {
	info = &ucgrpc.VerifyInfo{}
	v := req.Value
	if !req.Share {
		key := util.Yield16ByteKey([]byte(req.Passworld))
		v = util.AesEncrypt(v, key)
		db.Put(req.Key, req.Value, req.Datatype, req.UserAddress, req.Share, "", req.Strict)

	} else {
		db.Put(req.Key, req.Value, req.Datatype, req.UserAddress, req.Share, req.ShareChan, req.Strict)

	}
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
