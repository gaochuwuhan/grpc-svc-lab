package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"grpc-svc-lab/src/pb"
	"io"
	"log"
	"sync"
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
	//调用客户端流grpc
	//CallUserScoreFromClientStream(client)

	//调用服务端流grpc
	//CallUserScoreFromServerStream(client)
	//调用双向流grpc
	CallUserScoreBothStream(client)

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
	c,err:=client.GetUserScoreByClientStream(context.Background()) //得到客户端可以发送的行为接口，只有send和 	CloseAndRecv

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

func CallUserScoreFromServerStream(client pb.UserServiceClient){
	//模拟发送的数据
	req:=new(pb.UserScoreReq)
	users:=make([]*pb.UserInfo,0)
	for i:=0;i<5;i++{
		users = append(users,&pb.UserInfo{UserId:int32(i+20)})
		req.Users = users
	}
	fmt.Println("[服务端流]客户端mock的users",users)


	c,_:=client.GetUserScoreByServerStream(context.Background(),req)
	//客户端不断接收
	for {
		res,err:=c.Recv()
		if err == io.EOF {
			break
		}
		if err!=nil{
			log.Fatal(err)
			break
		}
		fmt.Println("[服务端流]客户端拿到从服务端返回的users",res.Users)

	}
}

func CallUserScoreBothStream(client pb.UserServiceClient){
	c,_:=client.GetUserScoreBothStream(context.Background())
	wg:=sync.WaitGroup{}
	wg.Add(2)
	//流式传送+流式接收：两个goroutine
	//1. 发送
	go func() {
		for {
			//time.Sleep(*time.Second)
			//模拟发送的数据
			req := new(pb.UserScoreReq)
			users := make([]*pb.UserInfo, 0)
			for i := 0; i < 3; i++ {
				users = append(users, &pb.UserInfo{UserId: int32(i + 30)})
				req.Users = users
			}
			err := c.Send(req)
			if err != nil {
				wg.Done()
				break
			}
		}

	}()

	//2.获取响应的数据
	go func() {
		for {
			res,err:=c.Recv()
			if err != nil {
				fmt.Println(err)
				wg.Done()
				break
			}
			fmt.Println("[双向流]客户端得到返回的响应users:",res.GetUsers())

		}
	}()
	wg.Wait()


}