package domain

import "time"

type Reminder struct {
	ID          int64
	GuildID     string
	ChannelID   string
	UserID      string
	Message     string
	ScheduledAt time.Time
	CreatedAt   time.Time
}
