package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DiscordToken  string
	GuildID       string // 開発・テスト用のGuild ID（指定すると即座にコマンドが登録されます）
	DBPath        string
	NotifyChannel string // デフォルトの通知・定例アジェンダ投稿先チャンネルID（任意）
}

func Load() (*Config, error) {
	// .envファイルが存在する場合は読み込む
	_ = godotenv.Load()

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("環境変数 DISCORD_TOKEN が設定されていません")
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "todocord.db" // デフォルトのSQLiteファイルパス
	}

	return &Config{
		DiscordToken:  token,
		GuildID:       os.Getenv("GUILD_ID"),
		DBPath:        dbPath,
		NotifyChannel: os.Getenv("NOTIFY_CHANNEL_ID"),
	}, nil
}
