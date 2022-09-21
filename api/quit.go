package api

import (
	"context"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	"services-match/pkg/db/gormx"
	rdx "services-match/pkg/db/rdx"
)

//	玩家退出匹配
func Quit(userId int32) error {
	//	写入退出匹配行为记录
	dbProxy := gormx.OpenMysql()
	if err := dbProxy.Create(&gormx.MatchRecord{UserId: userId, Behavior: gormx.Quit}).Error; err != nil {
		return fmt.Errorf("match record create failed: %v", err)
	}

	//	删除玩家信息
	matchPoolRedisClient := rdx.MatchPoolRedisClient()
	defer matchPoolRedisClient.Close()
	err := matchPoolRedisClient.Del(context.Background(), "match#"+fmt.Sprint(userId)).Err()
	if err != nil {
		return fmt.Errorf("the player quit match failed: %v", err)
	}

	//	发布退出匹配事件
	eventBusRedisClient := rdx.EventBusRedisClient()
	defer eventBusRedisClient.Close()
	eventQuitSuccessPayload, err := json.Marshal(&EventQuitSuccessPayload{IsForce: false})
	if err != nil {
		log.Fatalf("marshal quit success event payload failed: %v", err)
	}
	err = eventBusRedisClient.Publish(context.Background(), "Match_QuitSuccess_u"+fmt.Sprint(userId), eventQuitSuccessPayload).Err()
	if err != nil {
		log.Errorf("publish quit success event failed: %v", err)
	}

	return nil
}
