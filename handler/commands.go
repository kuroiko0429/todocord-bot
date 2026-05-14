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
					Name:        "phase",
					Description: "DTM制作フェーズ",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "作詞", Value: string(domain.PhaseLyrics)},
						{Name: "作曲", Value: string(domain.PhaseCompose)},
						{Name: "編曲", Value: string(domain.PhaseArrange)},
						{Name: "Mix", Value: string(domain.PhaseMix)},
						{Name: "Mas", Value: string(domain.PhaseMastering)},
						{Name: "レコーディング", Value: string(domain.PhaseRecording)},
						{Name: "なし", Value: string(domain.PhaseNone)},
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
					Name:        "phase",
					Description: "新しい制作フェーズ",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "作詞", Value: string(domain.PhaseLyrics)},
						{Name: "作曲", Value: string(domain.PhaseCompose)},
						{Name: "編曲", Value: string(domain.PhaseArrange)},
						{Name: "Mix", Value: string(domain.PhaseMix)},
						{Name: "Mas", Value: string(domain.PhaseMastering)},
						{Name: "レコーディング", Value: string(domain.PhaseRecording)},
						{Name: "なし", Value: string(domain.PhaseNone)},
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

	phase := domain.PhaseNone
	if opt, ok := opts["phase"]; ok {
		phase = domain.DTMPhase(opt.StringValue())
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
		Phase:       phase,
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
	tasks, err := h.repo.ListTasks(i.GuildID, nil)
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

	embed, components := h.service.BuildTaskListSummary(tasks, "")
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
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

	if opt, ok := opts["phase"]; ok {
		task.Phase = domain.DTMPhase(opt.StringValue())
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
