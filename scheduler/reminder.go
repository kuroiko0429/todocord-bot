package scheduler

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"todocord/config"
	"todocord/domain"
	"todocord/repository"
	"todocord/service"
)

type ReminderScheduler struct {
	repo                *repository.TaskRepository
	service             *service.TaskService
	session             *discordgo.Session
	config              *config.Config
	stopCh              chan struct{}
	lastOverdueCheckDay int
}

func NewReminderScheduler(repo *repository.TaskRepository, svc *service.TaskService, s *discordgo.Session, cfg *config.Config) *ReminderScheduler {
	return &ReminderScheduler{
		repo:    repo,
		service: svc,
		session: s,
		config:  cfg,
		stopCh:  make(chan struct{}),
	}
}

func (s *ReminderScheduler) Start() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.processReminders()
			case <-s.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
	log.Println("スケジューラー（自動リマインダー）を起動しました")
}

func (s *ReminderScheduler) Stop() {
	close(s.stopCh)
	log.Println("スケジューラーを停止しました")
}

func (s *ReminderScheduler) processScheduledReminders(now time.Time) {
	reminders, err := s.repo.GetPendingReminders(now)
	if err != nil {
		return
	}
	for _, rem := range reminders {
		mention := ""
		if rem.UserID != "" {
			mention = fmt.Sprintf("<@%s> ", rem.UserID)
		}
		_, _ = s.session.ChannelMessageSend(rem.ChannelID, fmt.Sprintf("⏰ %s**リマインダー**\n%s", mention, rem.Message))
		_ = s.repo.DeleteReminder(rem.ID)
	}
}

func (s *ReminderScheduler) processReminders() {
	now := time.Now()

	s.processScheduledReminders(now)

	dayTasks, err := s.repo.GetUnremindedDayTasks(now)
	if err == nil {
		for _, t := range dayTasks {
			chID := t.ChannelID
			if s.config.NotifyChannel != "" {
				chID = s.config.NotifyChannel
			}
			assignee := "担当者様"
			if t.AssigneeID != nil && *t.AssigneeID != "" {
				assignee = fmt.Sprintf("<@%s>", *t.AssigneeID)
			}
			msg := fmt.Sprintf("⏰ **【期限1日前リマインド】**\n%s、タスク「**%s**」の期限まで残り24時間を切りました！\n期限: <t:%d:F> (<t:%d:R>)",
				assignee, t.Title, t.Deadline.Unix(), t.Deadline.Unix())

			embed := s.service.BuildTaskEmbed(t)
			_, _ = s.session.ChannelMessageSendComplex(chID, &discordgo.MessageSend{
				Content: msg,
				Embeds:  []*discordgo.MessageEmbed{embed},
			})
			_ = s.repo.MarkRemindedDay(t.ID)
		}
	}

	hourTasks, err := s.repo.GetUnremindedHourTasks(now)
	if err == nil {
		for _, t := range hourTasks {
			chID := t.ChannelID
			if s.config.NotifyChannel != "" {
				chID = s.config.NotifyChannel
			}
			assignee := "担当者様"
			if t.AssigneeID != nil && *t.AssigneeID != "" {
				assignee = fmt.Sprintf("<@%s>", *t.AssigneeID)
			}
			msg := fmt.Sprintf("⚠️ **【期限直前リマインド (残り1時間)】**\n%s、タスク「**%s**」の期限まで残りわずかです！ラストスパート頑張りましょう！🔥\n期限: <t:%d:R>",
				assignee, t.Title, t.Deadline.Unix())

			embed := s.service.BuildTaskEmbed(t)
			_, _ = s.session.ChannelMessageSendComplex(chID, &discordgo.MessageSend{
				Content: msg,
				Embeds:  []*discordgo.MessageEmbed{embed},
			})
			_ = s.repo.MarkRemindedHour(t.ID)
		}
	}

	if now.Hour() == 9 && s.lastOverdueCheckDay != now.Day() {
		s.lastOverdueCheckDay = now.Day()
		overdueTasks, err := s.repo.ListAllOverdueTasks(now)
		if err == nil && len(overdueTasks) > 0 {
			guildTasks := make(map[string][]*domain.Task)
			for _, t := range overdueTasks {
				guildTasks[t.GuildID] = append(guildTasks[t.GuildID], t)
			}

			for _, tasks := range guildTasks {
				if len(tasks) == 0 {
					continue
				}
				chID := s.config.NotifyChannel
				if chID == "" {
					chID = tasks[0].ChannelID
				}

				embed, components := s.service.BuildTaskListSummary(tasks, "⚠️ 期限切れ")
				embed.Color = 0xE74C3C

				_, _ = s.session.ChannelMessageSendComplex(chID, &discordgo.MessageSend{
					Content:    fmt.Sprintf("📢 **【毎日の期限切れタスク通知】**\n現在、期限を過ぎている未完了タスクが **%d件** あります。早めの対応・リスケジュールを行いましょう！", len(tasks)),
					Embeds:     []*discordgo.MessageEmbed{embed},
					Components: components,
				})
			}
		}
	}
}
