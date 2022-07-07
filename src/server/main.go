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
	"time"
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

//服务端流式返回
func(s *server) GetUserScoreByServerStream(req *pb.UserScoreReq,server pb.UserService_GetUserScoreByServerStreamServer) error {
	//client端传什么req，server端返回什么
	i:=0
	for{
		if i>5{ //服务端会返回6次给客户端
			break
		}
		//mock res
		res:=new(pb.UserScoreRes)
		res.Users = make([]*pb.UserInfo,0)
		for _,user := range req.Users {
			user.UserScore = 200+user.UserId
			res.Users = append(res.Users,user)
		}
		time.Sleep(1*time.Second)
		server.Send(res)
		i++
	}
	return nil

}

//双向流
func(s *server) GetUserScoreBothStream(server pb.UserService_GetUserScoreBothStreamServer) error{

	//服务端开goroutine 读取req，
	readUsers:=make(chan []*pb.UserInfo)
	l:=0
	go func() {
		for {
			req,_:=server.Recv()
			//if err!=nil{
			//	readUsers <- nil //防止channel死锁
			//	break
			//}
			if l>5{
				readUsers <- nil //防止channel死锁
				break
			}
			readUsers<-req.GetUsers()
			l++


		}
	}()
	//发送回客户端处理后的数据
	for {
		if <- readUsers == nil{
			break
		}
		//将req的score赋值
		res:=new(pb.UserScoreRes)
		res.Users = <-readUsers //只要当前channel有req则读取
		//mock res
		for k,requser:=range res.Users {
			res.Users[k].UserScore = 200+requser.UserId
		}
		fmt.Println("[双向流]服务端响应的users:",res.Users)
		//发送res给客户端
		server.Send(res)
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