package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"todocord/domain"
	"todocord/repository"
	"todocord/service"
)

type CommandHandler struct {
	repo    *repository.TaskRepository
	service *service.TaskService
}

func NewCommandHandler(repo *repository.TaskRepository, svc *service.TaskService) *CommandHandler {
	return &CommandHandler{
		repo:    repo,
		service: svc,
	}
}

func (h *CommandHandler) Commands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "add",
			Description: "新しいタスクを登録します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "title",
					Description: "タスク名",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "description",
					Description: "タスクの詳細説明",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "deadline",
					Description: "期限 (例: 2026-05-20 または 2026-05-20 15:00)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "priority",
					Description: "優先度",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "🔴 High (至急)", Value: string(domain.PriorityHigh)},
						{Name: "🟡 Medium (通常)", Value: string(domain.PriorityMedium)},
						{Name: "🟢 Low (余裕あり)", Value: string(domain.PriorityLow)},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "category",
					Description: "カテゴリ",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "部費", Value: string(domain.CategoryBuhi)},
						{Name: "物販関連", Value: string(domain.CategoryMerchandise)},
						{Name: "物品管理", Value: string(domain.CategoryInventory)},
						{Name: "口座計画", Value: string(domain.CategoryAccount)},
						{Name: "楽曲品評会", Value: string(domain.CategoryMusicReview)},
						{Name: "デザイン", Value: string(domain.CategoryDesign)},
						{Name: "学祭", Value: string(domain.CategoryFestival)},
						{Name: "新人歓迎会", Value: string(domain.CategoryWelcome)},
						{Name: "クリスマス会", Value: string(domain.CategoryChristmas)},
						{Name: "大事な話", Value: string(domain.CategoryImportant)},
						{Name: "なし", Value: string(domain.CategoryNone)},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "assignee",
					Description: "担当メンバー",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "demo_url",
					Description: "デモ音源のURL（Dropbox, Drive, SoundCloud等）",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "bpm",
					Description: "楽曲のBPM",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "key",
					Description: "楽曲のKey (例: C Major, Am等)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "shared_link",
					Description: "プロジェクトファイル等の共有フォルダURL",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "create_thread",
					Description: "タスク専用のスレッドを自動作成するかどうか",
					Required:    false,
				},
			},
		},
		{
			Name:        "list",
			Description: "現在残っているタスク一覧を表示・管理します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "category",
					Description: "カテゴリで絞り込む",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "部費", Value: string(domain.CategoryBuhi)},
						{Name: "物販関連", Value: string(domain.CategoryMerchandise)},
						{Name: "物品管理", Value: string(domain.CategoryInventory)},
						{Name: "口座計画", Value: string(domain.CategoryAccount)},
						{Name: "楽曲品評会", Value: string(domain.CategoryMusicReview)},
						{Name: "デザイン", Value: string(domain.CategoryDesign)},
						{Name: "学祭", Value: string(domain.CategoryFestival)},
						{Name: "新人歓迎会", Value: string(domain.CategoryWelcome)},
						{Name: "クリスマス会", Value: string(domain.CategoryChristmas)},
						{Name: "大事な話", Value: string(domain.CategoryImportant)},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "priority",
					Description: "優先度で絞り込む",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "🔴 High", Value: string(domain.PriorityHigh)},
						{Name: "🟡 Medium", Value: string(domain.PriorityMedium)},
						{Name: "🟢 Low", Value: string(domain.PriorityLow)},
					},
				},
			},
		},
		{
			Name:        "assign",
			Description: "タスクに担当者を割り当てます",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "task_id",
					Description: "タスクID",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "member",
					Description: "割り当てるメンバー",
					Required:    true,
				},
			},
		},
		{
			Name:        "status",
			Description: "タスクのステータスまたは制作フェーズを変更します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "task_id",
					Description: "タスクID",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "status",
					Description: "新しいステータス",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "未着手", Value: string(domain.StatusTodo)},
						{Name: "進行中", Value: string(domain.StatusInProgress)},
						{Name: "完了", Value: string(domain.StatusDone)},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "category",
					Description: "新しいカテゴリ",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "部費", Value: string(domain.CategoryBuhi)},
						{Name: "物販関連", Value: string(domain.CategoryMerchandise)},
						{Name: "物品管理", Value: string(domain.CategoryInventory)},
						{Name: "口座計画", Value: string(domain.CategoryAccount)},
						{Name: "楽曲品評会", Value: string(domain.CategoryMusicReview)},
						{Name: "デザイン", Value: string(domain.CategoryDesign)},
						{Name: "学祭", Value: string(domain.CategoryFestival)},
						{Name: "新人歓迎会", Value: string(domain.CategoryWelcome)},
						{Name: "クリスマス会", Value: string(domain.CategoryChristmas)},
						{Name: "大事な話", Value: string(domain.CategoryImportant)},
						{Name: "なし", Value: string(domain.CategoryNone)},
					},
				},
			},
		},
		{
			Name:        "edit",
			Description: "登録済みのタスク内容を修正します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "task_id",
					Description: "修正するタスクID",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "title",
					Description: "新しいタスク名",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "description",
					Description: "新しい詳細説明",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "deadline",
					Description: "新しい期限",
					Required:    false,
				},
			},
		},
		{
			Name:        "delete",
			Description: "タスクを削除します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "task_id",
					Description: "削除するタスクID",
					Required:    true,
				},
			},
		},
		{
			Name:        "mtg",
			Description: "定例会用のアジェンダ（未完了タスク優先度順リスト）を作成します",
		},
		{
			Name:        "reminders",
			Description: "設定中のリマインダー一覧を表示・削除します",
		},
		{
			Name:        "remind",
			Description: "指定した日時にメッセージを通知します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "time",
					Description: "通知日時 (例: 2026-05-20 15:00)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "通知するメッセージ",
					Required:    true,
				},
			},
		},
		{
			Name:        "report",
			Description: "月ごとの稼働・完了タスクレポートを出力します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "year",
					Description: "年 (例: 2026)",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "month",
					Description: "月 (1〜12)",
					Required:    false,
				},
			},
		},
	}
}

func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	optMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optMap[opt.Name] = opt
	}
	return optMap
}

func parseDeadline(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	layouts := []string{
		"2006-01-02 15:04",
		"2006/01/02 15:04",
		"2006-01-02",
		"2006/01/02",
	}

	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, s, time.Local)
		if err == nil {
			if len(layout) == 10 {
				t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.Local)
			}
			return &t, nil
		}
	}
	return nil, fmt.Errorf("日時の形式が正しくありません。YYYY-MM-DD または YYYY-MM-DD HH:MM で入力してください")
}

func (h *CommandHandler) HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	switch data.Name {
	case "add":
		h.handleAdd(s, i)
	case "list":
		h.handleList(s, i)
	case "assign":
		h.handleAssign(s, i)
	case "status":
		h.handleStatus(s, i)
	case "edit":
		h.handleEdit(s, i)
	case "delete":
		h.handleDelete(s, i)
	case "mtg":
		h.handleMtg(s, i)
	case "reminders":
		h.handleReminders(s, i)
	case "remind":
		h.handleRemind(s, i)
	case "report":
		h.handleReport(s, i)
	}
}

func (h *CommandHandler) handleAdd(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := parseOptions(i.ApplicationCommandData().Options)

	title := opts["title"].StringValue()
	desc := "記載なし"
	if opt, ok := opts["description"]; ok {
		desc = opt.StringValue()
	}

	priority := domain.PriorityMedium
	if opt, ok := opts["priority"]; ok {
		priority = domain.TaskPriority(opt.StringValue())
	}

	category := domain.CategoryNone
	if opt, ok := opts["category"]; ok {
		category = domain.TaskCategory(opt.StringValue())
	}

	var deadline *time.Time
	if opt, ok := opts["deadline"]; ok {
		parsed, err := parseDeadline(opt.StringValue())
		if err != nil {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("⚠️ %s", err.Error()),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
		deadline = parsed
	}

	var assigneeID *string
	if opt, ok := opts["assignee"]; ok {
		user := opt.UserValue(s)
		if user != nil {
			assigneeID = &user.ID
		}
	}

	var demoURL, keyInfo, sharedLink *string
	if opt, ok := opts["demo_url"]; ok {
		v := opt.StringValue()
		demoURL = &v
	}
	if opt, ok := opts["key"]; ok {
		v := opt.StringValue()
		keyInfo = &v
	}
	if opt, ok := opts["shared_link"]; ok {
		v := opt.StringValue()
		sharedLink = &v
	}

	var bpm *float64
	if opt, ok := opts["bpm"]; ok {
		v := opt.FloatValue()
		bpm = &v
	}

	createThread := false
	if opt, ok := opts["create_thread"]; ok {
		createThread = opt.BoolValue()
	}

	task := &domain.Task{
		GuildID:     i.GuildID,
		ChannelID:   i.ChannelID,
		Title:       title,
		Description: desc,
		Priority:    priority,
		Status:      domain.StatusTodo,
		Category:    category,
		AssigneeID:  assigneeID,
		Deadline:    deadline,
		DemoURL:     demoURL,
		BPM:         bpm,
		KeyInfo:     keyInfo,
		SharedLink:  sharedLink,
	}

	_, err := h.repo.CreateTask(task)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ タスクの保存中にエラーが発生しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	embed := h.service.BuildTaskEmbed(task)
	components := h.service.BuildTaskDetailComponents(task)

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
	if err != nil {
		return
	}

	if createThread {
		msg, err := s.InteractionResponse(i.Interaction)
		if err == nil && msg != nil {
			threadName := fmt.Sprintf("💬 %s", title)
			if len(threadName) > 100 {
				threadName = threadName[:97] + "..."
			}
			thread, err := s.MessageThreadStartComplex(i.ChannelID, msg.ID, &discordgo.ThreadStart{
				Name:                threadName,
				AutoArchiveDuration: 4320,
			})
			if err == nil && thread != nil {
				task.ThreadID = &thread.ID
				_ = h.repo.UpdateTask(task)

				_, _ = s.ChannelMessageSend(thread.ID, fmt.Sprintf("タスク「**%s**」の専用スレッドが作成されました！制作の進捗報告や議論はこちらで行いましょう。✨", title))
			}
		}
	}
}

func (h *CommandHandler) handleList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := parseOptions(i.ApplicationCommandData().Options)

	var filter domain.TaskFilter
	titleSuffix := ""
	if opt, ok := opts["category"]; ok {
		c := domain.TaskCategory(opt.StringValue())
		filter.Category = &c
		titleSuffix += fmt.Sprintf(" [%s]", c)
	}
	if opt, ok := opts["priority"]; ok {
		p := domain.TaskPriority(opt.StringValue())
		filter.Priority = &p
		titleSuffix += fmt.Sprintf(" [%s]", p)
	}

	tasks, err := h.repo.ListTasks(i.GuildID, &filter)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ 一覧の取得中にエラーが発生しました",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	embed, components := h.service.BuildTaskListSummary(tasks, titleSuffix)
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

func (h *CommandHandler) handleReminders(s *discordgo.Session, i *discordgo.InteractionCreate) {
	reminders, err := h.repo.GetRemindersByGuild(i.GuildID)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "❌ リマインダーの取得に失敗しました", Flags: discordgo.MessageFlagsEphemeral},
		})
		return
	}

	if len(reminders) == 0 {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "設定中のリマインダーはありません。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	var desc string
	var rows []discordgo.MessageComponent
	for idx, rem := range reminders {
		if idx >= 5 {
			desc += fmt.Sprintf("\n...他 %d 件", len(reminders)-5)
			break
		}
		desc += fmt.Sprintf("**#%d** <t:%d:F>\n> %s\n", rem.ID, rem.ScheduledAt.Unix(), rem.Message)
		rows = append(rows, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: fmt.Sprintf("cancel_reminder_%d", rem.ID),
					Label:    fmt.Sprintf("#%d をキャンセル", rem.ID),
					Style:    discordgo.DangerButton,
				},
			},
		})
	}

	embed := &discordgo.MessageEmbed{
		Title:       "⏰ 設定中のリマインダー一覧",
		Description: desc,
		Color:       0x3498DB,
	}
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: rows,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *CommandHandler) handleAssign(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := parseOptions(i.ApplicationCommandData().Options)
	taskID := opts["task_id"].IntValue()
	user := opts["member"].UserValue(s)

	task, err := h.repo.GetTask(taskID)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "⚠️ 指定されたタスクが見つかりません", Flags: discordgo.MessageFlagsEphemeral},
		})
		return
	}

	task.AssigneeID = &user.ID
	if err := h.repo.UpdateTask(task); err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "❌ 更新エラーが発生しました", Flags: discordgo.MessageFlagsEphemeral},
		})
		return
	}

	embed := h.service.BuildTaskEmbed(task)
	components := h.service.BuildTaskDetailComponents(task)
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    fmt.Sprintf("✅ タスクの担当者を <@%s> に設定しました！", user.ID),
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

func (h *CommandHandler) handleStatus(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := parseOptions(i.ApplicationCommandData().Options)
	taskID := opts["task_id"].IntValue()

	task, err := h.repo.GetTask(taskID)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "⚠️ 指定されたタスクが見つかりません", Flags: discordgo.MessageFlagsEphemeral},
		})
		return
	}

	var contentMsg string
	if opt, ok := opts["status"]; ok {
		newStatus := domain.TaskStatus(opt.StringValue())
		if newStatus == domain.StatusDone && task.Status != domain.StatusDone {
			now := time.Now()
			task.CompletedAt = &now
			contentMsg = h.service.GetCelebrationMessage(task, i.Member.User.ID)
		} else if newStatus != domain.StatusDone {
			task.CompletedAt = nil
		}
		task.Status = newStatus
	}

	if opt, ok := opts["category"]; ok {
		task.Category = domain.TaskCategory(opt.StringValue())
	}

	if err := h.repo.UpdateTask(task); err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "❌ 更新エラーが発生しました", Flags: discordgo.MessageFlagsEphemeral},
		})
		return
	}

	if contentMsg == "" {
		contentMsg = "✅ タスクのステータス/フェーズを更新しました。"
	}

	embed := h.service.BuildTaskEmbed(task)
	components := h.service.BuildTaskDetailComponents(task)
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    contentMsg,
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

func (h *CommandHandler) handleEdit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := parseOptions(i.ApplicationCommandData().Options)
	taskID := opts["task_id"].IntValue()

	task, err := h.repo.GetTask(taskID)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "⚠️ 指定されたタスクが見つかりません", Flags: discordgo.MessageFlagsEphemeral},
		})
		return
	}

	if opt, ok := opts["title"]; ok {
		task.Title = opt.StringValue()
	}
	if opt, ok := opts["description"]; ok {
		task.Description = opt.StringValue()
	}
	if opt, ok := opts["deadline"]; ok {
		parsed, err := parseDeadline(opt.StringValue())
		if err != nil {
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{Content: fmt.Sprintf("⚠️ %s", err.Error()), Flags: discordgo.MessageFlagsEphemeral},
			})
			return
		}
		task.Deadline = parsed
	}

	if err := h.repo.UpdateTask(task); err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "❌ 更新エラーが発生しました", Flags: discordgo.MessageFlagsEphemeral},
		})
		return
	}

	embed := h.service.BuildTaskEmbed(task)
	components := h.service.BuildTaskDetailComponents(task)
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    "✅ タスク内容を修正しました。",
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

func (h *CommandHandler) handleDelete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := parseOptions(i.ApplicationCommandData().Options)
	taskID := opts["task_id"].IntValue()

	err := h.repo.DeleteTask(taskID)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "❌ 削除中にエラーが発生しました", Flags: discordgo.MessageFlagsEphemeral},
		})
		return
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🗑️ タスク (ID: %d) を正常に削除しました。", taskID),
		},
	})
}

func (h *CommandHandler) handleMtg(s *discordgo.Session, i *discordgo.InteractionCreate) {
	tasks, err := h.repo.ListTasks(i.GuildID, nil)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "❌ エラーが発生しました", Flags: discordgo.MessageFlagsEphemeral},
		})
		return
	}

	embed, components := h.service.BuildTaskListSummary(tasks, "定例会アジェンダ: 未完了")
	embed.Color = 0x9B59B6
	embed.Footer = &discordgo.MessageEmbedFooter{Text: "優先度順（High -> Medium -> Low）および期限順にソートされています"}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:    "📢 **定例会用アジェンダを出力しました。** 各タスクの進捗状況を確認しましょう！",
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}

func (h *CommandHandler) handleRemind(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := parseOptions(i.ApplicationCommandData().Options)

	timeStr := opts["time"].StringValue()
	message := opts["message"].StringValue()

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

	rem := &domain.Reminder{
		GuildID:     i.GuildID,
		ChannelID:   i.ChannelID,
		UserID:      userID,
		Message:     message,
		ScheduledAt: *scheduledAt,
	}

	if _, err := h.repo.CreateReminder(rem); err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ リマインダーの保存に失敗しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("⏰ リマインダーを設定しました！\n**日時:** <t:%d:F>\n**内容:** %s", scheduledAt.Unix(), message),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *CommandHandler) handleReport(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := parseOptions(i.ApplicationCommandData().Options)

	now := time.Now()
	year := now.Year()
	month := now.Month()

	if opt, ok := opts["year"]; ok {
		year = int(opt.IntValue())
	}
	if opt, ok := opts["month"]; ok {
		month = time.Month(opt.IntValue())
	}

	tasks, err := h.repo.GetMonthlyReport(i.GuildID, year, month)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "❌ レポート取得エラー", Flags: discordgo.MessageFlagsEphemeral},
		})
		return
	}

	desc := fmt.Sprintf("**%d年%d月** に完了したタスク一覧です。皆様の素晴らしい稼働実績です！👏\n\n", year, month)
	if len(tasks) == 0 {
		desc += "該当月に完了したタスクはありません。"
	} else {
		for _, t := range tasks {
			completedStr := ""
			if t.CompletedAt != nil {
				completedStr = t.CompletedAt.Format("01/02 15:04")
			}
			assignee := "未指定"
			if t.AssigneeID != nil && *t.AssigneeID != "" {
				assignee = fmt.Sprintf("<@%s>", *t.AssigneeID)
			}
			desc += fmt.Sprintf("✅ **%s** (担当: %s / 完了日: %s)\n", t.Title, assignee, completedStr)
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📊 %d年%d月 稼働・完了タスクレポート", year, month),
		Description: desc,
		Color:       0xF1C40F,
		Timestamp:   now.Format(time.RFC3339),
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}
