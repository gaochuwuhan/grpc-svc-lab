package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"grpc-svc-lab/src/pb"
	"io"
	"log"
	"net"
	"os"
)

type server struct {
	pb.UnimplementedUserServiceServer
}

//服务端普通调用方法
func (s *server) GetUserScore(c context.Context, req *pb.UserScoreReq) (*pb.UserScoreRes, error) {
	res,err:=new(UserSvc).GetUserScore(c,req)
	return res,err
}

//服务端对客户端流式传入处理
func(s *server) GetUserScoreByClientStream(server pb.UserService_GetUserScoreByClientStreamServer) error {
	//由于客户端不断流式传入，所以服务端要不断读取
	users:=make([]*pb.UserInfo,0)
	for {
		req,err:=server.Recv() //服务端接收 判断是否eof
		fmt.Println("[客户端流]服务端当前接收的req:",req)
		if err == io.EOF{ //接收完毕
			server.SendAndClose(&pb.UserScoreRes{Users: users}) //返回根据req，server端逻辑处理后的users
			break
		}
		if err != nil{
			log.Fatal(err)
			return err
		}
		for _,user :=range req.Users{
			user.UserScore=user.UserId+110 //模拟处理
			users = append(users,user)
		}
	}
	return nil
}


func main() {
	l,err:=net.Listen("tcp", ":2001")
	if err!=nil{
		log.Fatal(err)
		os.Exit(1)
	}
	s:=grpc.NewServer()
	pb.RegisterUserServiceServer(s,&server{})
	s.Serve(l)
}