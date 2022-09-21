package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"services-match/pkg/db/gormx"
	rdx "services-match/pkg/db/rdx"
)

type UserInfo struct {
	UserId int32
	JoinAt time.Time
}

type EventJoinSuccessPayload struct{}

//	玩家加入匹配
func Join(userId int32) error {
	//	检查玩家是否在匹配
	if CheckIdExist(userId) {
		return fmt.Errorf("player %d exist in match", userId)
	}

	//	写入加入匹配行为记录
	dbProxy := gormx.OpenMysql()
	if err := dbProxy.Create(&gormx.MatchRecord{UserId: userId, Behavior: gormx.Join}).Error; err != nil {
		return fmt.Errorf("match record create failed: %v", err)
	}

	//	玩家加入匹配，并保存玩家的信息
	matchPoolRedisClient := rdx.MatchPoolRedisClient()
	defer matchPoolRedisClient.Close()
	currentTime := time.Now()
	userInfoJson, err := json.Marshal(&UserInfo{UserId: userId, JoinAt: currentTime})
	if err != nil {
		log.Fatalf("marshal user info json failed: %v", err)
	}
	err = matchPoolRedisClient.Set(context.Background(), "match#"+fmt.Sprint(userId), string(userInfoJson), 0).Err()
	if err != nil {
		return fmt.Errorf("the player join match failed: %v", err)
	}

	//	发布加入匹配成功事件
	eventBusRedisClient := rdx.EventBusRedisClient()
	defer eventBusRedisClient.Close()
	eventJoinSuccessPayload, err := json.Marshal(&EventJoinSuccessPayload{})
	if err != nil {
		log.Fatalf("marshal join success event payload failed: %v", err)
	}
	err = eventBusRedisClient.Publish(context.Background(), "Match_JoinSuccess_u"+fmt.Sprint(userId), eventJoinSuccessPayload).Err()
	if err != nil {
		log.Errorf("publish join success event failed: %v", err)
	}

	return nil
}
