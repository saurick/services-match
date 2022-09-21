package listener

import (
	"context"
	"encoding/json"

	"services-match/api"
	myRedis "services-match/pkg/db/rdx"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

type EventType string

const (
	EventType_LaunchSuccess    EventType = "LaunchSuccess"
	EventType_GameStartSuccess EventType = "GameStartSuccess"
)

type SandboxEvent struct {
	Type    EventType   // type
	Payload interface{} // 用来放Payload
	Meta    interface{} // 用来放一些辅助识别业务逻辑的其他内容
}

//	监听Sandbox事件
func SandboxEventListener() {
	//	订阅Sandbox事件
	eventBusRedisClient := myRedis.EventBusRedisClient()
	pubsub := eventBusRedisClient.PSubscribe(context.Background(), "Sandbox_LaunchSuccess_*?", "Sandbox_GameStartSuccess_*?")
	defer func(pubsub *redis.PubSub) {
		err := pubsub.Close()
		if err != nil {
			log.WithError(err).Warn("defer close pubsub failed")
		}
	}(pubsub)

	for {
		ch := pubsub.Channel()
		channel := <-ch
		var (
			sandboxEvent SandboxEvent
			matchId      int32
		)

		if err := json.Unmarshal([]byte(channel.Payload), &sandboxEvent); err != nil {
			log.Fatalf("unmarshal event sandbox launch success payload failed: %v", err)
		}

		switch sandboxEvent.Type {
		case EventType_LaunchSuccess:
			//	解析payload
			if sandboxEvent.Meta == "launch sandbox success" {
				matchId = int32(sandboxEvent.Payload.(map[string]interface{})["MatchId"].(float64))
			}

			if err := api.SandboxLaunchSuccess(matchId); err != nil {
				log.Errorf("sandbox launch failed: %v", err)
			}
		case EventType_GameStartSuccess:
			//	解析payload
			if sandboxEvent.Meta == "game start success" {
				matchId = int32(sandboxEvent.Payload.(map[string]interface{})["MatchId"].(float64))
			}

			if err := api.SandboxGameStartSuccess(matchId); err != nil {
				log.Errorf("sandbox game start failed: %v", err)
			}
		}
	}
}
