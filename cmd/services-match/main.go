package main

import (
	"flag"

	"services-match/listener"
	"services-match/matcher"
	"services-match/server"
	fakeServer "services-match/test/mock/fake_server"
)

var (
	port = flag.Int("port", 30053, "The server port")
)

func main() {
	flag.Parse()
	//	匹配轮询
	go matcher.MatchPolling()

	//	Sandbox事件订阅
	go listener.SandboxEventListener()

	//	调用模拟的Sandbox服务
	go fakeServer.Sandbox_Services()

	//	启动匹配服务
	server.LaunchServiecesMatchServer(port)
}
