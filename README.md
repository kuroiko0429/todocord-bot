# 🎵 Todocord Bot — DTM役員サーバー向け Discord ToDo管理Bot

Go言語製のDiscordスラッシュコマンドBot。DTM（デスクトップミュージック）サークル・役員サーバーの制作管理に特化した、多機能なタスク管理システムです。

---

## ✨ 主な機能

### 📋 基本タスク管理
| コマンド | 説明 |
|---|---|
| `/add` | タスクを登録（タイトル・詳細・期限・担当者・音源URL・BPM・Keyなど） |
| `/list` | 未完了タスクをリッチEmbedと対話型セレクトメニューで表示 |
| `/assign` | 担当メンバーを割り当て |
| `/status` | ステータス（未着手/進行中/完了）や制作フェーズを変更 |
| `/edit` | タスク内容を修正 |
| `/delete` | タスクを削除 |

### 🔔 通知・リマインド
- **期限1日前・1時間前**: 担当者にメンション付きで自動DM通知
- **毎朝9時**: 期限切れの未完了タスク一覧を対象チャンネルへ自動投稿
- **タスク完了時**: ランダムなお祝いメッセージを自動送信

### 🎶 DTM特化機能
- 制作フェーズ管理（作詞 / 作曲 / 編曲 / Mix / Mas / レコーディング）
- BPM・Key情報のメタデータ保存
- デモ音源URL（SoundCloud・Drive等）の紐付け
- Dropbox・Google Drive等の共有リンク管理
- 大規模タスクの**専用スレッド自動作成**

### 🏛️ 役員・運営向け機能
| コマンド | 説明 |
|---|---|
| `/mtg` | 定例会アジェンダを優先度順（High → Medium → Low）で自動生成 |
| `/report` | 指定月の完了タスクレポートを出力 |

---

## 🚀 セットアップ

### 必要なもの
- Go 1.21+
- Discord Bot Token（[Discord Developer Portal](https://discord.com/developers/applications)で取得）

### 手順

**1. 環境変数の設定**
```bash
make setup        # .env.example から .env を生成
```
生成された `.env` を編集します:
```env
DISCORD_TOKEN=your_bot_token_here      # 必須
GUILD_ID=your_test_server_id           # 任意（指定するとコマンドが即時反映）
DB_PATH=todocord.db                    # SQLiteファイルのパス（デフォルトでOK）
NOTIFY_CHANNEL_ID=channel_id           # 任意（リマインドの投稿先チャンネル）
```

**2. ビルドして起動**
```bash
make build        # バイナリをビルド
./build/bot       # 起動

# または開発用に直接実行
make run
```

**3. Discordサーバーへの招待**

Discord Developer Portal で Bot を作成し、以下のスコープと権限を付与してサーバーへ招待してください:
- **スコープ**: `bot`, `applications.commands`
- **権限**: `Send Messages`, `Read Message History`, `Create Public Threads`, `Embed Links`, `Mention Everyone`

---

## 🗂️ プロジェクト構成

```
todocord-bot/
├── cmd/bot/
│   └── main.go             # エントリポイント
├── config/
│   └── config.go           # 環境変数設定
├── domain/
│   └── task.go             # タスクドメインモデル（型・定数定義）
├── handler/
│   ├── commands.go         # スラッシュコマンドハンドラー
│   └── components.go       # ボタン・セレクトメニューハンドラー
├── repository/
│   ├── sqlite.go           # SQLite CRUD / クエリ
│   └── sqlite_test.go      # リポジトリ単体テスト
├── scheduler/
│   └── reminder.go         # 自動リマインドワーカー
├── service/
│   └── task_service.go     # Embed・コンポーネント生成ロジック
├── .env.example            # 環境変数サンプル
├── Makefile
└── go.mod
```

---

## 🧪 テスト

```bash
make test
```

---

## 🛠️ 技術スタック

| 項目 | 技術 |
|---|---|
| 言語 | Go 1.21+ |
| Discord API | `bwmarrin/discordgo` |
| データベース | SQLite (`modernc.org/sqlite` / CGO不要) |
| 環境変数 | `joho/godotenv` |

---

## 📝 ライセンス

MIT
