package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"todocord/domain"
)

func (h *CommandHandler) HandleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	data := i.MessageComponentData()
	customID := data.CustomID

	if customID == "select_task" {
		if len(data.Values) == 0 {
			return
		}
		taskID, _ := strconv.ParseInt(data.Values[0], 10, 64)
		task, err := h.repo.GetTask(taskID)
		if err != nil {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: "⚠️ タスクの取得に失敗しました", Flags: discordgo.MessageFlagsEphemeral},
			})
			return
		}

		embed := h.service.BuildTaskEmbed(task)
		components := h.service.BuildTaskDetailComponents(task)

		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: components,
			},
		})
		return
	}

	if customID == "back_list" {
		tasks, err := h.repo.ListTasks(i.GuildID, nil)
		if err != nil {
			return
		}
		embed, components := h.service.BuildTaskListSummary(tasks, "")
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: components,
			},
		})
		return
	}

	parts := strings.SplitN(customID, "_", 2)
	if len(parts) < 2 {
		return
	}
	action := parts[0]
	taskID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return
	}

	task, err := h.repo.GetTask(taskID)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "⚠️ タスクが見つかりません", Flags: discordgo.MessageFlagsEphemeral},
		})
		return
	}

	var responseContent string

	switch action {
	case "remind":
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: fmt.Sprintf("remind_modal_%d", taskID),
				Title:    "リマインダーを設定",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "remind_time",
								Label:       "通知日時",
								Style:       discordgo.TextInputShort,
								Placeholder: "2026-05-20 15:00",
								Required:    true,
							},
						},
					},
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "remind_message",
								Label:       "通知メッセージ",
								Style:       discordgo.TextInputParagraph,
								Placeholder: "リマインダーの内容を入力...",
								Required:    true,
							},
						},
					},
				},
			},
		})
		return

	case "complete":
		if task.Status != domain.StatusDone {
			now := time.Now()
			task.CompletedAt = &now
			task.Status = domain.StatusDone
			_ = h.repo.UpdateTask(task)
			responseContent = h.service.GetCelebrationMessage(task, i.Member.User.ID)
		}

	case "thread":
		if task.ThreadID == nil {
			threadName := fmt.Sprintf("💬 %s", task.Title)
			if len(threadName) > 100 {
				threadName = threadName[:97] + "..."
			}
			thread, err := s.MessageThreadStartComplex(i.ChannelID, i.Message.ID, &discordgo.ThreadStart{
				Name:                threadName,
				AutoArchiveDuration: 4320,
			})
			if err == nil && thread != nil {
				task.ThreadID = &thread.ID
				_ = h.repo.UpdateTask(task)
				_, _ = s.ChannelMessageSend(thread.ID, fmt.Sprintf("タスク「**%s**」専用スレッドを作成しました！✨", task.Title))
				responseContent = "✅ 専用スレッドを作成しました。"
			} else {
				responseContent = "❌ スレッド作成に失敗しました。"
			}
		}

	case "status":
		if len(data.Values) > 0 {
			newStatus := domain.TaskStatus(data.Values[0])
			if newStatus == domain.StatusDone && task.Status != domain.StatusDone {
				now := time.Now()
				task.CompletedAt = &now
				responseContent = h.service.GetCelebrationMessage(task, i.Member.User.ID)
			} else if newStatus != domain.StatusDone {
				task.CompletedAt = nil
			}
			task.Status = newStatus
			_ = h.repo.UpdateTask(task)
			if responseContent == "" {
				responseContent = fmt.Sprintf("✅ ステータスを「%s」に変更しました。", newStatus)
			}
		}

	case "phase":
		if len(data.Values) > 0 {
			task.Phase = domain.DTMPhase(data.Values[0])
			_ = h.repo.UpdateTask(task)
			responseContent = fmt.Sprintf("🎶 制作フェーズを「%s」に変更しました。", task.Phase)
		}
	}

	embed := h.service.BuildTaskEmbed(task)
	components := h.service.BuildTaskDetailComponents(task)

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    responseContent,
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}
