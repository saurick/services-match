package api

import (
	"context"
	"encoding/json"
	"fmt"
	"services-match/pkg/db/gormx"
	rdx "services-match/pkg/db/rdx"
	serverPb "services-match/services-gen-code"
	"time"

	log "github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
)

//	匹配成功后
func MatchSuccess(matchingQueue []UserInfo, seconds []int64, players []int32) error {
	var (
		conn    *grpc.ClientConn
		client  serverPb.SandboxServiceClient
		result  *serverPb.S2CLaunchByMatch //	rpc返回结果
		matchId int32                      //	匹配id，即自增id
		err     error
	)

	//	匹配成功事务
	dbProxy := gormx.OpenMysql()
	err = dbProxy.Transaction(func(tx *gorm.DB) error {
		//	写入匹配成功对局记录
		currentTime := time.Now()
		userIds := fmt.Sprint(players[0]) + ";" + fmt.Sprint(players[1])
		duration := currentTime.Sub(time.Unix(seconds[0], 0))
		if err = tx.Create(&gormx.Match{UserIds: userIds, Duration: duration}).Error; err != nil {
			return fmt.Errorf("table match create failed: %v", err)
		}

		//	取出match id，即gorm model的自增id
		var match gormx.Match
		if err = tx.Last(&match).Error; err != nil {
			return fmt.Errorf("query last record of table match failed: %v", err)
		}
		matchId = int32(match.Model.ID)

		//	请求获取Sandbox的ip:port:token信息
		conn, client = ConnectToSandboxServiceClient()
		defer conn.Close()
		md := metadata.New(map[string]string{"match_id": fmt.Sprint(matchId)})
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		result, err = client.LaunchByMatch(ctx, &serverPb.C2SLaunchByMatch{})
		if err != nil {
			return fmt.Errorf("rpc 'launch by match' failed: %v", err)
		}

		//	向匹配成功对局记录更新SandboxInfo字段
		sandboxInfo := result.ServerInfo.Ip + ":" + fmt.Sprint(result.ServerInfo.Port) + ":" + result.ServerInfo.Token
		if err := tx.Model(&gormx.Match{}).Where("id = ?", matchId).Updates(&gormx.Match{SandboxInfo: sandboxInfo}).Error; err != nil {
			return fmt.Errorf("table match update the field sandbox_info failed: %v", err)
		}

		//	写入玩家1和玩家2进行匹配行为记录
		if err = tx.Create(&gormx.MatchRecord{UserId: players[0], Behavior: gormx.MatchSuccess}).Error; err != nil {
			return fmt.Errorf("match record for player0 create failed: %v", err)
		}
		if err = tx.Create(&gormx.MatchRecord{UserId: players[1], Behavior: gormx.MatchSuccess}).Error; err != nil {
			return fmt.Errorf("match record for player1 create failed: %v", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("match success transaction failed: %v", err)
	}

	//	删除玩家信息
	matchPoolRedisClient := rdx.MatchPoolRedisClient()
	defer matchPoolRedisClient.Close()
	for _, s := range matchingQueue {
		err = matchPoolRedisClient.Del(context.Background(), "match#"+fmt.Sprint(s.UserId)).Err()
		if err != nil {
			return fmt.Errorf("remove matched player failed: %v", err)
		}
	}

	player1Id := players[0]
	player2Id := players[1]

	eventBusRedisClient := rdx.EventBusRedisClient()
	defer eventBusRedisClient.Close()
	//	发布玩家1匹配成功事件
	eventMatchSuccessPayload, err := json.Marshal(&EventMatchSuccessPayload{PlayerId: player1Id, MatchId: matchId, SandboxIp: result.ServerInfo.Ip, SandboxPort: result.ServerInfo.Port, SandboxToken: result.ServerInfo.Token})
	if err != nil {
		log.Fatalf("marshal match success event payload failed: %v", err)
	}
	err = eventBusRedisClient.Publish(context.Background(), "Match_MatchSuccess_m"+fmt.Sprint(matchId)+"_u"+fmt.Sprint(player1Id), eventMatchSuccessPayload).Err()
	if err != nil {
		log.Errorf("publish match success event for player1 failed: %v", err)
	}
	//	发布玩家2匹配成功事件
	eventMatchSuccessPayload, err = json.Marshal(&EventMatchSuccessPayload{PlayerId: player2Id, MatchId: matchId, SandboxIp: result.ServerInfo.Ip, SandboxPort: result.ServerInfo.Port, SandboxToken: result.ServerInfo.Token})
	if err != nil {
		log.Fatalf("marshal match success event payload failed: %v", err)
	}
	err = eventBusRedisClient.Publish(context.Background(), "Match_MatchSuccess_m"+fmt.Sprint(matchId)+"_u"+fmt.Sprint(player2Id), eventMatchSuccessPayload).Err()
	if err != nil {
		log.Errorf("publish match success event for player2 failed: %v", err)
	}
	return nil
}
