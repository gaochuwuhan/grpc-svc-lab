package main

import (
	"context"
	"grpc-svc-lab/src/pb"
)

type UserSvc struct{

}

func(u *UserSvc) GetUserScore(ctx context.Context,in *pb.UserScoreReq) (*pb.UserScoreRes,error){
	users:=make([]*pb.UserInfo,0)
	//从客户端中拿出请求的in,mock score
	for k,u:=range in.Users{
		users = append(users,&pb.UserInfo{
			UserId: u.UserId,
			UserScore: int32(k+100),
		})
	}

	return &pb.UserScoreRes{
		Users: users,
	}, nil
}
