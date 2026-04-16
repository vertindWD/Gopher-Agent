package model

import (
	"time"

	"gorm.io/gorm"
)

// 任务状态枚举
const (
	TaskStatusPending   = "pending"   // 已接收请求，推入 Kafka 等待消费
	TaskStatusRunning   = "running"   // Worker 正在调用 Agent 处理
	TaskStatusCompleted = "completed" // 处理成功
	TaskStatusFailed    = "failed"    // 处理失败（大模型超时或报错）
)

// AgentTask 代表一次 AI 任务
type AgentTask struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TaskID    string         `gorm:"type:varchar(64);uniqueIndex;not null" json:"task_id"` // 对外暴露的唯一业务ID
	Prompt    string         `gorm:"type:text;not null" json:"prompt"`                     // 用户的原始指令
	Status    string         `gorm:"type:varchar(20);default:'pending'" json:"status"`     // 任务当前状态
	Result    string         `gorm:"type:longtext" json:"result"`                          // AI 最终给出的结果
	ErrorMsg  string         `gorm:"type:text" json:"error_msg"`                           // 失败时的错误栈日志
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (AgentTask) TableName() string {
	return "agent_tasks"
}
