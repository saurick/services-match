package api

import (
	"context"
	"encoding/json"
	"fmt"
	"services-match/pkg/db/gormx"
	rdx "services-match/pkg/db/rdx"

	log "github.com/sirupsen/logrus"
)

//	强制玩家退出匹配
func ForceQuit(userId int32) error {
	//	写入强制玩家退出匹配行为记录
	dbProxy := gormx.OpenMysql()
	if err := dbProxy.Create(&gormx.MatchRecord{UserId: userId, Behavior: gormx.ForceQuit}).Error; err != nil {
		return fmt.Errorf("match record create failed: %v", err)
	}

	//	删除玩家信息
	matchPoolRedisClient := rdx.MatchPoolRedisClient()
	defer matchPoolRedisClient.Close()
	err := matchPoolRedisClient.Del(context.Background(), "match#"+fmt.Sprint(userId)).Err()
	if err != nil {
		return fmt.Errorf("failed to force the player to quit the match: %v", err)
	}

	//	发布强制玩家退出事件
	eventBusRedisClient := rdx.EventBusRedisClient()
	defer eventBusRedisClient.Close()
	eventForceQuitSuccessPayload, err := json.Marshal(&EventQuitSuccessPayload{IsForce: true})
	if err != nil {
		log.Fatalf("marshal force quit event payload failed: %v", err)
	}
	err = eventBusRedisClient.Publish(context.Background(), "Match_QuitSuccess_u"+fmt.Sprint(userId), eventForceQuitSuccessPayload).Err()
	if err != nil {
		log.Errorf("publish force quit success event failed: %v", err)
	}

	return nil
}
