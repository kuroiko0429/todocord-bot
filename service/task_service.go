package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"todocord/domain"
)

type TaskService struct{}

func NewTaskService() *TaskService {
	return &TaskService{}
}

// BuildTaskEmbed は単一タスクの美しい詳細Embedを構築します
func (s *TaskService) BuildTaskEmbed(t *domain.Task) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📌 %s", t.Title),
		Description: t.Description,
		Color:       t.PriorityColor(),
		Timestamp:   t.CreatedAt.Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Task ID: %d | 最終更新", t.ID),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "優先度",
				Value:  t.PriorityLabel(),
				Inline: true,
			},
			{
				Name:   "ステータス",
				Value:  string(t.Status),
				Inline: true,
			},
			{
				Name:   "制作フェーズ",
				Value:  string(t.Phase),
				Inline: true,
			},
		},
	}

	assigneeVal := "未割り当て"
	if t.AssigneeID != nil && *t.AssigneeID != "" {
		assigneeVal = fmt.Sprintf("<@%s>", *t.AssigneeID)
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "担当者",
		Value:  assigneeVal,
		Inline: true,
	})

	deadlineVal := "設定なし"
	if t.Deadline != nil {
		// Discordのタイムスタンプ表記を利用（ユーザーのローカルタイムゾーンで表示・相対時間表示）
		unix := t.Deadline.Unix()
		deadlineVal = fmt.Sprintf("<t:%d:F>\n(<t:%d:R>)", unix, unix)
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "期限",
		Value:  deadlineVal,
		Inline: true,
	})

	// DTM特化情報の追加
	if t.BPM != nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🎵 BPM",
			Value:  fmt.Sprintf("%.1f", *t.BPM),
			Inline: true,
		})
	}
	if t.KeyInfo != nil && *t.KeyInfo != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🔑 Key",
			Value:  *t.KeyInfo,
			Inline: true,
		})
	}
	if t.DemoURL != nil && *t.DemoURL != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "🎧 デモ音源",
			Value: fmt.Sprintf("[試聴・ダウンロードを開く](%s)", *t.DemoURL),
		})
	}
	if t.SharedLink != nil && *t.SharedLink != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  "📁 共有リンク (Drive/Dropbox等)",
			Value: fmt.Sprintf("[共有フォルダを開く](%s)", *t.SharedLink),
		})
	}

	return embed
}

// BuildTaskDetailComponents はタスク詳細画面のアクションボタンおよびステータス変更メニューを構築します
func (s *TaskService) BuildTaskDetailComponents(t *domain.Task) []discordgo.MessageComponent {
	// 1行目: クイックアクションボタン
	row1 := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				CustomID: fmt.Sprintf("complete_%d", t.ID),
				Label:    "完了にする 🎉",
				Style:    discordgo.SuccessButton,
				Disabled: t.Status == domain.StatusDone,
			},
			discordgo.Button{
				CustomID: fmt.Sprintf("thread_%d", t.ID),
				Label:    "専用スレッド作成 💬",
				Style:    discordgo.SecondaryButton,
				Disabled: t.ThreadID != nil, // 既にスレッドがある場合は無効化
			},
			discordgo.Button{
				CustomID: fmt.Sprintf("remind_%d", t.ID),
				Label:    "リマインダー設定 ⏰",
				Style:    discordgo.SecondaryButton,
			},
			discordgo.Button{
				CustomID: "back_list",
				Label:    "◀ タスク一覧に戻る",
				Style:    discordgo.PrimaryButton,
			},
		},
	}

	// 2行目: ステータス変更セレクトメニュー
	row2 := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    fmt.Sprintf("status_%d", t.ID),
				Placeholder: "🔄 ステータスを変更...",
				Options: []discordgo.SelectMenuOption{
					{Label: "未着手", Value: string(domain.StatusTodo), Default: t.Status == domain.StatusTodo},
					{Label: "進行中", Value: string(domain.StatusInProgress), Default: t.Status == domain.StatusInProgress},
					{Label: "完了", Value: string(domain.StatusDone), Default: t.Status == domain.StatusDone},
				},
			},
		},
	}

	// 3行目: 制作フェーズ変更セレクトメニュー
	row3 := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    fmt.Sprintf("phase_%d", t.ID),
				Placeholder: "🎶 制作フェーズを変更...",
				Options: []discordgo.SelectMenuOption{
					{Label: "なし", Value: string(domain.PhaseNone), Default: t.Phase == domain.PhaseNone},
					{Label: "作詞", Value: string(domain.PhaseLyrics), Default: t.Phase == domain.PhaseLyrics},
					{Label: "作曲", Value: string(domain.PhaseCompose), Default: t.Phase == domain.PhaseCompose},
					{Label: "編曲", Value: string(domain.PhaseArrange), Default: t.Phase == domain.PhaseArrange},
					{Label: "Mix", Value: string(domain.PhaseMix), Default: t.Phase == domain.PhaseMix},
					{Label: "Mas", Value: string(domain.PhaseMastering), Default: t.Phase == domain.PhaseMastering},
					{Label: "レコーディング", Value: string(domain.PhaseRecording), Default: t.Phase == domain.PhaseRecording},
				},
			},
		},
	}

	return []discordgo.MessageComponent{row1, row2, row3}
}

// BuildTaskListSummary はタスク一覧のサマリーEmbedと選択メニューを構築します
func (s *TaskService) BuildTaskListSummary(tasks []*domain.Task, titlePrefix string) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	if len(tasks) == 0 {
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("📋 %sタスク一覧", titlePrefix),
			Description: "現在残っているタスクはありません。素晴らしいです！✨",
			Color:       0x2ECC71,
		}
		return embed, nil
	}

	embed := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("📋 %sタスク一覧 (計 %d 件)", titlePrefix, len(tasks)),
		Color: 0x3498DB,
	}

	var desc string
	var options []discordgo.SelectMenuOption

	for i, t := range tasks {
		if i < 15 { // サマリー表示は最大15件程度に制限して視認性を確保
			assignee := "未割り当て"
			if t.AssigneeID != nil && *t.AssigneeID != "" {
				assignee = fmt.Sprintf("<@%s>", *t.AssigneeID)
			}
			deadline := "期限なし"
			if t.Deadline != nil {
				deadline = fmt.Sprintf("<t:%d:d>", t.Deadline.Unix())
			}
			desc += fmt.Sprintf("• **ID:%d** [%s] **%s** (担当: %s / 期限: %s)\n", t.ID, t.PriorityLabel(), t.Title, assignee, deadline)
		}

		// セレクトメニューのオプション構築（最大25件）
		if i < 25 {
			label := t.Title
			if len(label) > 50 {
				label = label[:47] + "..."
			}
			descOpt := fmt.Sprintf("優先度:%s | フェーズ:%s", t.Priority, t.Phase)
			options = append(options, discordgo.SelectMenuOption{
				Label:       fmt.Sprintf("ID:%d | %s", t.ID, label),
				Description: descOpt,
				Value:       fmt.Sprintf("%d", t.ID),
			})
		}
	}

	if len(tasks) > 15 {
		desc += fmt.Sprintf("\n...他 %d 件のタスクがあります。", len(tasks)-15)
	}
	embed.Description = desc

	menuRow := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    "select_task",
				Placeholder: "👉 詳細を確認・管理するタスクを選択...",
				Options:     options,
			},
		},
	}

	return embed, []discordgo.MessageComponent{menuRow}
}

// GetCelebrationMessage は完了報告用のランダムなお祝いメッセージを返します
func (s *TaskService) GetCelebrationMessage(t *domain.Task, completedByUserID string) string {
	messages := []string{
		"🎉 お疲れ様です！タスク「**%s**」が完了しました！最高です！✨",
		"🚀 素晴らしいペースですね！タスク「**%s**」完了、ありがとうございます！👏",
		"🎶 制作がまた一歩前進しました！タスク「**%s**」完了！乾杯！🥂",
		"✨ タスク「**%s**」をクリア！素晴らしい貢献に感謝します！💐",
		"🎧 名曲への階段をまた一つ登りました！タスク「**%s**」完了です！🔥",
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	msg := messages[r.Intn(len(messages))]
	base := fmt.Sprintf(msg, t.Title)

	if completedByUserID != "" {
		base += fmt.Sprintf(" (Completed by <@%s>)", completedByUserID)
	}
	return base
}
