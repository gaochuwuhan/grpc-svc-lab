package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"grpc-svc-lab/src/pb"
	"log"
	"time"
)

func main() {
	conn,err:=grpc.Dial("localhost:2001",grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()
	client:=pb.NewUserServiceClient(conn)
	//调用普通grpc请求
	//CallUserNormalGetScore(client)
	//调用服务端流grpc
	CallUserScoreFromClientStream(client)


}

func CallUserNormalGetScore(client pb.UserServiceClient){
	users:=make([]*pb.UserInfo,3,3)
	for i:=0;i<3;i++{
		users[i] = &pb.UserInfo{UserId:int32(i+1)}
	}
	req,_:=client.GetUserScore(context.Background(),&pb.UserScoreReq{Users:users})
	fmt.Println("【普通】从服务端返回userinfo：",req.GetUsers())
}

func CallUserScoreFromClientStream(client pb.UserServiceClient){
//客户端发送流
	c,err:=client.GetUserScoreByClientStream(context.Background()) //得到客户端可以发送的行为接口，只有send和 	CloseAndRecv() (*UserScoreRes, error)

	if err != nil{
		log.Fatal(err)
		return
	}

	//客户端每1s发送一个req
	i:=0
	for {
		if i>3{
			res,_:=c.CloseAndRecv() //客户端关闭并从服务端接收res
			fmt.Println("[客户端流] 客户端即将关闭发送，从服务端得到的res users是：",res.Users)
			break
		}
		//模拟发送的数据
		req:=new(pb.UserScoreReq)
		users:=make([]*pb.UserInfo,0)
		//假设一次就发送一条userinfo，也可以发送多个再用个for循环
		users = append(users,&pb.UserInfo{UserId:int32(i+10)})
		req.Users = users
		//发送
		c.Send(req)
		time.Sleep(1 * time.Second)
		i++

	}
}
