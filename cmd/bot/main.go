package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"todocord/config"
	"todocord/handler"
	"todocord/repository"
	"todocord/scheduler"
	"todocord/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("設定読み込みエラー: %v", err)
	}

	repo, err := repository.NewTaskRepository(cfg.DBPath)
	if err != nil {
		log.Fatalf("DB初期化エラー: %v", err)
	}
	defer repo.Close()

	svc := service.NewTaskService()

	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatalf("Discordセッション作成エラー: %v", err)
	}

	cmdHandler := handler.NewCommandHandler(repo, svc)

	commands := cmdHandler.Commands()
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Botが正常に起動しました！連携ユーザー: %s#%s", r.User.Username, r.User.Discriminator)
log.Println("スラッシュコマンドを登録しています...")
		for _, v := range commands {
			_, err := s.ApplicationCommandCreate(r.User.ID, cfg.GuildID, v)
			if err != nil {
				log.Printf("コマンド '%s' の登録に失敗しました: %v", v.Name, err)
				return
			}
		}
		if cfg.GuildID != "" {
			log.Printf("指定されたサーバー (GuildID: %s) にコマンドを即時登録しました。", cfg.GuildID)
		} else {
			log.Println("グローバルにコマンドを登録しました。（反映まで時間がかかる場合があります）")
		}
	})
	dg.AddHandler(cmdHandler.HandleCommand)
	dg.AddHandler(cmdHandler.HandleComponent)
	dg.AddHandler(cmdHandler.HandleModal)

	if err := dg.Open(); err != nil {
		log.Fatalf("Discordへの接続エラー: %v", err)
	}
	defer dg.Close()

	reminders := scheduler.NewReminderScheduler(repo, svc, dg, cfg)
	reminders.Start()
	defer reminders.Stop()

	log.Println("Botは現在稼働中です。終了するには Ctrl+C を押してください。")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	log.Println("シャットダウン処理を実行してBotを正常に終了しました。")
}
