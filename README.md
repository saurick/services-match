# 游戏匹配服务器

## 目标
实现双玩家匹配，先进入匹配池的优先进行匹配，如果有两个玩家加入，马上进行匹配，如果只有一个，等待几秒后和机器人进行匹配，使用redis进行事件发布订阅，使用protobuf进行消息转发

## 启动redis和mysql容器服务
执行命令`docker-compose up -d`启动相关容器服务

## 构建和运行
执行命令`make build-services-match`生成bin目录和可执行文件，直接执行`./bin/./services-match`来运行程序（注意端口是否冲突，默认端口是`30053`）

## mock
mock服务器在路径`test/mock/fake_server`下，目的是模拟Sandbox服务器rpc请求和事件发布

## 测试
需要同时运行[游戏匹配网关](https://git.sysfun.cn/projects/CB/repos/gateways-logic-match/browse)，本地安装redis-cli
1. 先clone[模拟游戏客户端](https://git.sysfun.cn/projects/CB/repos/testclients-csharp-match/browse)
```
$ git clone ssh://git@git.sysfun.cn:7999/cb/testclients-csharp-match.git
$ cd testclients-csharp-match
```
2. 开一个终端模拟1个玩家请求匹配
```
$ redis-cli -p 63790 set playerToken1 1
$ dotnet run Program.cs match playerToken1 30052
```
3. 开一个终端模拟玩家11个玩家请求匹配
```
$ cat > match.sh <<EOF
# bash/bin

max=11

for i in \`seq 1 \$max\`
do
    redis-cli -p 63790 set playerToken\$i \$i
done

for i in \`seq 1 \$max\`
do
    player="playerToken\$i"
    dotnet run Program.cs "match" \$player 30052 &
    sleep 1
done
EOF

$ chmod 700 match.sh

$ ./match.sh
```
4. 模拟1个玩家取消匹配
```
$ redis-cli -p 63790 set playerToken1 1
$ dotnet run Program.cs cancel playerToken1 30052
```

## TODO
* 单元测试
* 性能测试
* 匹配机器人实现
* 其他匹配规则的实现
