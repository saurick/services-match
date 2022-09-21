package matcher

import (
	"context"
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"services-match/api"
	rdx "services-match/pkg/db/rdx"
	"sort"
	"time"
)

//	检测玩家加入匹配的时间戳是否在数组中
func checkInSeconds(seconds []int64, joinAt int64) bool {
	for _, second := range seconds {
		if second == joinAt {
			return true
		}
	}
	return false
}

//	TODO 将来可能更换匹配规则，目前是存在匹配队列最久的两个玩家匹配
func match(matchingQueue []api.UserInfo, seconds []int64) {
	var players []int32 //	已匹配的玩家

	//	如果有两个以上玩家加入，马上匹配
	if len(matchingQueue) >= 2 {
		// 按时间戳从小到大排序
		sort.Slice(matchingQueue, func(i, j int) bool {
			return matchingQueue[i].JoinAt.Unix() < matchingQueue[j].JoinAt.Unix()
		})
		sort.Slice(seconds, func(i, j int) bool {
			return seconds[i] < seconds[j]
		})

		//	玩家进行匹配
		players = []int32{}
		player1 := matchingQueue[0].UserId
		player2 := matchingQueue[1].UserId
		players = append(players, player1)
		players = append(players, player2)

		//	玩家匹配成功后
		api.MatchSuccess(matchingQueue, seconds, players)
	}

	//	如果只有一个玩家，那么超过8秒之后就和机器人匹配
	if len(matchingQueue) == 1 {
		dateStamp := time.Now().Unix()
		sub := dateStamp - seconds[0]
		if sub > 8 {
			//	玩家和机器人进行匹配
			players = []int32{}
			uniquePlayer := matchingQueue[0].UserId
			players = append(players, uniquePlayer)
			var robot int32 = 10086 //	TODO 如何代表机器人
			players = append(players, robot)

			//	玩家匹配成功后
			api.MatchSuccess(matchingQueue, seconds, players)
		}
	}
}

//	匹配轮询
func MatchPolling() {
	matchPoolRedisClient := rdx.MatchPoolRedisClient()
	var (
		keys          []string
		err           error
		cursor        uint64
		matchingQueue []api.UserInfo //	准备进入匹配的玩家
		seconds       []int64        //	准备进入匹配的玩家的时间戳
	)
	for {
		//	初始化准备进入匹配的玩家数组
		matchingQueue = []api.UserInfo{}
		//	初始化准备进入匹配的玩家的时间戳数组
		seconds = []int64{}

		//	获取玩家信息的所有key
		keys, _, err = matchPoolRedisClient.Scan(context.Background(), cursor, "*?", 0).Result()
		if err != nil {
			log.Errorf("scan keys failed err: %v", err)
		}

		//	获取所有玩家的信息以及加入匹配的时间戳，并分别存入数组
		for _, key := range keys {
			val, err := matchPoolRedisClient.Get(context.Background(), key).Result()
			if err != nil {
				log.Errorf("get key values failed err: %v", err)
			}
			var userInfo api.UserInfo
			err = json.Unmarshal([]byte(val), &userInfo)
			if err != nil {
				log.Errorf("unmarshal user info failed: %v", err)
			}
			joinDateStamp := userInfo.JoinAt.Unix()
			if !checkInSeconds(seconds, joinDateStamp) {
				seconds = append(seconds, joinDateStamp)
				matchingQueue = append(matchingQueue, userInfo)
			}
		}

		//	具体的玩家匹配逻辑
		match(matchingQueue, seconds)

		//	每秒轮询一次
		time.Sleep(time.Second)
	}
}
