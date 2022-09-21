package api

import (
	"context"
	"fmt"
	rdx "services-match/pkg/db/rdx"
	serverPb "services-match/services-gen-code"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type EventQuitSuccessPayload struct {
	IsForce bool
}

type EventMatchSuccessPayload struct {
	PlayerId     int32
	MatchId      int32
	SandboxIp    string
	SandboxPort  int32
	SandboxToken string
}

//	检测玩家是否在匹配中
func CheckIdExist(userId int32) bool {
	matchPoolRedisClient := rdx.MatchPoolRedisClient()
	defer matchPoolRedisClient.Close()
	_, err := matchPoolRedisClient.Get(context.Background(), "match#"+fmt.Sprint(userId)).Result()
	if err == redis.Nil {
		return false
	}
	if err != redis.Nil && err != nil {
		log.Errorf("check exist failed: %v", err)
		return false
	}
	return true
}

//	rpc连接到Sandbox
func ConnectToSandboxServiceClient() (*grpc.ClientConn, serverPb.SandboxServiceClient) {
	port := strconv.Itoa(30054)
	conn, err := grpc.Dial("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	client := serverPb.NewSandboxServiceClient(conn)
	return conn, client
}
