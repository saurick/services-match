package api

import (
	"fmt"
	"services-match/pkg/db/gormx"
	"time"
)

//	Sandbox启动成功后
func SandboxLaunchSuccess(matchId int32) error {
	//	写入Sandbox启动成功对局记录
	dbProxy := gormx.OpenMysql()
	currentTime := time.Now()
	if err := dbProxy.Model(&gormx.Match{}).Where("id = ?", matchId).Updates(&gormx.Match{ReadyAt: &currentTime}).Error; err != nil {
		return fmt.Errorf("table match update the field ready_at failed: %v", err)
	}
	return nil
}

//	Sandbox游戏成功开始后
func SandboxGameStartSuccess(matchId int32) error {
	//	写入Sandbox游戏开始对局记录
	dbProxy := gormx.OpenMysql()
	currentTime := time.Now()
	if err := dbProxy.Model(&gormx.Match{}).Where("id = ?", matchId).Updates(&gormx.Match{StartedAt: &currentTime}).Error; err != nil {
		return fmt.Errorf("table match update the field started_at failed: %v", err)
	}
	return nil
}
