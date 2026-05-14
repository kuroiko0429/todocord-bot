package domain

import (
	"time"
)

type TaskStatus string

const (
	StatusTodo       TaskStatus = "未着手"
	StatusInProgress TaskStatus = "進行中"
	StatusDone       TaskStatus = "完了"
)

type TaskPriority string

const (
	PriorityHigh   TaskPriority = "High"
	PriorityMedium TaskPriority = "Medium"
	PriorityLow    TaskPriority = "Low"
)

type DTMPhase string

const (
	PhaseNone      DTMPhase = "なし"
	PhaseLyrics    DTMPhase = "作詞"
	PhaseCompose   DTMPhase = "作曲"
	PhaseArrange   DTMPhase = "編曲"
	PhaseMix       DTMPhase = "Mix"
	PhaseMastering DTMPhase = "Mas"
	PhaseRecording DTMPhase = "レコーディング"
)

type Task struct {
	ID          int64
	GuildID     string
	ChannelID   string
	ThreadID    *string
	Title       string
	Description string
	Priority    TaskPriority
	Status      TaskStatus
	Phase       DTMPhase
	AssigneeID  *string
	Deadline    *time.Time
	DemoURL     *string
	BPM         *float64
	KeyInfo     *string
	SharedLink  *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

// 優先度に応じたカラーコード（Discord Embed用）
func (t *Task) PriorityColor() int {
	switch t.Priority {
	case PriorityHigh:
		return 0xE74C3C // 鮮やかな赤
	case PriorityMedium:
		return 0xF1C40F // 鮮やかな黄
	case PriorityLow:
		return 0x2ECC71 // エメラルドグリーン
	default:
		return 0x3498DB // スカイブルー
	}
}

// 優先度の絵文字付きラベル表示
func (t *Task) PriorityLabel() string {
	switch t.Priority {
	case PriorityHigh:
		return "🔴 High (至急)"
	case PriorityMedium:
		return "🟡 Medium (通常)"
	case PriorityLow:
		return "🟢 Low (余裕あり)"
	default:
		return string(t.Priority)
	}
}

// 期限切れかどうかを判定
func (t *Task) IsOverdue(now time.Time) bool {
	if t.Deadline == nil || t.Status == StatusDone {
		return false
	}
	return now.After(*t.Deadline)
}
