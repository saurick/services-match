package gormx

import (
	"time"

	"gorm.io/gorm"
)

const (
	Join         Behavior = "Join"
	Quit         Behavior = "Quit"
	ForceQuit    Behavior = "ForceQuit"
	MatchSuccess Behavior = "MatchSuccess"
)

type Behavior string

//	匹配行为记录
type MatchRecord struct {
	gorm.Model
	UserId   int32    `gorm:"index;comment:用户id"`
	Behavior Behavior `gorm:"index;comment:服务器行为"`
	Remark   string   `gorm:"comment:记录其它信息"`
}

//	对局状态记录
type Match struct {
	gorm.Model
	UserIds     string        `gorm:"comment:参与者id集合"`
	Duration    time.Duration `gorm:"comment:匹配时长"`
	SandboxInfo string        `gorm:"comment:Sandbox相关信息"`
	ReadyAt     *time.Time    `gorm:"comment:Sandbox准备好的时间"`
	StartedAt   *time.Time    `gorm:"comment:游戏开始的时间"`
	Remark      string        `gorm:"comment:记录其它信息"`
}
