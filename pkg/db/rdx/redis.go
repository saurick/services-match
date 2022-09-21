package rdx

import "github.com/go-redis/redis/v8"

//	用来做pubsub
func EventBusRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:63791",
	})
}

//	存match#id:userInfo，表示已经加入匹配池
func MatchPoolRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:63792",
	})
}
