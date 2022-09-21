//	TODO 这里是模拟沙盒服务器，用来模拟grpc和发布订阅，将来可能会被完全替代

package fake_server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"

	myRedis "services-match/pkg/db/rdx"
	serverPb "services-match/services-gen-code"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type SandBoxServer struct {
	serverPb.UnimplementedSandboxServiceServer
}

type EventType string
type LaunchSuccessPayload struct {
	MatchId int32
}
type GameStartSuccessPayload struct {
	MatchId int32
}

const (
	EventType_LaunchSuccess    EventType = "LaunchSuccess"
	EventType_GameStartSuccess EventType = "GameStartSuccess"
)

type SandboxEvent struct {
	Type    EventType   // type
	Payload interface{} // 用来放Payload
	Meta    interface{} // 用来放一些辅助识别业务逻辑的其他内容
}

//	生成唯一token
func generateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

//	获取可用端口
func getAvailablePort() (int32, error) {
	address, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", "0.0.0.0"))
	if err != nil {
		return 0, err
	}
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return int32(listener.Addr().(*net.TCPAddr).Port), nil
}

//	返回Sandbox erver info的grpc接口
func (m *SandBoxServer) LaunchByMatch(ctx context.Context, in *serverPb.C2SLaunchByMatch) (*serverPb.S2CLaunchByMatch, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}
	ip := "192.168.0.100"
	port, err := getAvailablePort()
	if err != nil {
		fmt.Println(err)
	}
	token := generateSecureToken(24)
	info := &serverPb.ServerInfo{Ip: ip, Port: port, Token: token}

	if md["match_id"] != nil {
		matchId, _ := strconv.ParseInt(md["match_id"][0], 10, 32)
		go launchSandbox(int32(matchId))
		go startGame(int32(matchId))
	}

	return &serverPb.S2CLaunchByMatch{ServerInfo: info}, nil
}

//	启动Sandbox
func launchSandbox(matchId int32) {
	eventBusRedisClient := myRedis.EventBusRedisClient()
	defer eventBusRedisClient.Close()
	payload := LaunchSuccessPayload{MatchId: matchId}
	meta := "launch sandbox success"

	eventSandboxLaunchSuccessPayload, err := json.Marshal(&SandboxEvent{Type: EventType_LaunchSuccess, Payload: payload, Meta: meta})
	if err != nil {
		log.Fatalf("marshal launch sandbox success event payload failed: %v", err)
	}

	err = eventBusRedisClient.Publish(context.Background(), "Sandbox_LaunchSuccess_m"+fmt.Sprint(matchId), eventSandboxLaunchSuccessPayload).Err()
	if err != nil {
		fmt.Println(err)
	}
}

//	开始游戏
func startGame(matchId int32) {
	eventBusRedisClient := myRedis.EventBusRedisClient()
	defer eventBusRedisClient.Close()
	payload := GameStartSuccessPayload{MatchId: matchId}
	meta := "game start success"

	eventSandboxGameStartSuccessPayload, err := json.Marshal(&SandboxEvent{Type: EventType_GameStartSuccess, Payload: payload, Meta: meta})
	if err != nil {
		log.Fatalf("marshal game start event payload failed: %v", err)
	}

	err = eventBusRedisClient.Publish(context.Background(), "Sandbox_GameStartSuccess_m"+fmt.Sprint(matchId), eventSandboxGameStartSuccessPayload).Err()
	if err != nil {
		fmt.Println(err)
	}
}

//	Sandbox服务
func Sandbox_Services() {
	lis, err := net.Listen("tcp", "localhost:30054")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	serverPb.RegisterSandboxServiceServer(grpcServer, &SandBoxServer{})
	log.Printf("sandbox server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
