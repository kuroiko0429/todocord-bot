package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"todocord/domain"
)

func (h *CommandHandler) HandleModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionModalSubmit {
		return
	}

	data := i.ModalSubmitData()
	if !strings.HasPrefix(data.CustomID, "remind_modal_") {
		return
	}

	taskIDStr := strings.TrimPrefix(data.CustomID, "remind_modal_")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return
	}

	var timeStr, message string
	for _, row := range data.Components {
		ar, ok := row.(*discordgo.ActionsRow)
		if !ok {
			continue
		}
		for _, comp := range ar.Components {
			ti, ok := comp.(*discordgo.TextInput)
			if !ok {
				continue
			}
			switch ti.CustomID {
			case "remind_time":
				timeStr = ti.Value
			case "remind_message":
				message = ti.Value
			}
		}
	}

	scheduledAt, err := parseDeadline(timeStr)
	if err != nil || scheduledAt == nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "⚠️ 日時の形式が正しくありません。`2026-05-20 15:00` のように入力してください。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if scheduledAt.Before(time.Now()) {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "⚠️ 過去の日時は指定できません。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	userID := ""
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
	}

	task, _ := h.repo.GetTask(taskID)
	taskTitle := fmt.Sprintf("タスクID:%d", taskID)
	if task != nil {
		taskTitle = task.Title
	}

	rem := &domain.Reminder{
		GuildID:     i.GuildID,
		ChannelID:   i.ChannelID,
		UserID:      userID,
		Message:     fmt.Sprintf("[%s] %s", taskTitle, message),
		ScheduledAt: *scheduledAt,
	}

	if _, err := h.repo.CreateReminder(rem); err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ リマインダーの保存に失敗しました。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("⏰ リマインダーを設定しました！\n**タスク:** %s\n**日時:** <t:%d:F>\n**内容:** %s", taskTitle, scheduledAt.Unix(), message),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
