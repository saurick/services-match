package server

import (
	"context"
	"fmt"
	"net"
	"services-match/api"
	serverPb "services-match/services-gen-code"

	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MatchServer struct {
	serverPb.UnimplementedMatchServer
}

//	玩家加入匹配
func (m *MatchServer) Join(ctx context.Context, in *serverPb.C2SJoin) (*serverPb.S2CJoin, error) {
	userId := in.UserId
	if err := api.Join(userId); err != nil {
		return nil, status.Error(codes.Canceled, "the player join match failed")
	}
	return &serverPb.S2CJoin{}, nil
}

//	玩家退出匹配
func (m *MatchServer) Quit(ctx context.Context, in *serverPb.C2SQuit) (*serverPb.S2CQuit, error) {
	userId := in.UserId
	if err := api.Quit(userId); err != nil {
		return nil, status.Error(codes.Canceled, "the player quit match failed")
	}
	return &serverPb.S2CQuit{}, nil
}

//	强制玩家退出匹配
func (m *MatchServer) ForceQuit(ctx context.Context, in *serverPb.C2SForceQuit) (*serverPb.S2CForceQuit, error) {
	userId := in.UserId
	if err := api.ForceQuit(userId); err != nil {
		return nil, status.Error(codes.Canceled, "failed to force the player to quit the match")
	}
	return &serverPb.S2CForceQuit{}, nil
}

//	启动匹配服务
func LaunchServiecesMatchServer(port *int) {
	//	监听端口
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	//	注册grpc服务器
	grpcServer := grpc.NewServer()
	serverPb.RegisterMatchServer(grpcServer, &MatchServer{})
	log.Printf("server listening at %v", lis.Addr())

	//	启动grpc服务
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
