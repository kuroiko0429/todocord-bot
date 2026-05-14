package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
	"todocord/domain"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(dbPath string) (*TaskRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("データベース接続エラー: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("データベース疎通エラー: %w", err)
	}

	repo := &TaskRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("テーブル初期化エラー: %w", err)
	}

	return repo, nil
}

func (r *TaskRepository) Close() error {
	return r.db.Close()
}

func (r *TaskRepository) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			thread_id TEXT,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			priority TEXT NOT NULL,
			status TEXT NOT NULL,
			phase TEXT NOT NULL,
			assignee_id TEXT,
			deadline DATETIME,
			demo_url TEXT,
			bpm REAL,
			key_info TEXT,
			shared_link TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			completed_at DATETIME,
			reminded_day INTEGER DEFAULT 0,
			reminded_hour INTEGER DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS reminders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			guild_id TEXT NOT NULL,
			channel_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			message TEXT NOT NULL,
			scheduled_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL
		);`,
	}
	for _, q := range queries {
		if _, err := r.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (r *TaskRepository) CreateReminder(rem *domain.Reminder) (int64, error) {
	query := `INSERT INTO reminders (guild_id, channel_id, user_id, message, scheduled_at, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	rem.CreatedAt = time.Now()
	res, err := r.db.Exec(query, rem.GuildID, rem.ChannelID, rem.UserID, rem.Message, rem.ScheduledAt, rem.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	rem.ID = id
	return id, nil
}

func (r *TaskRepository) GetPendingReminders(now time.Time) ([]*domain.Reminder, error) {
	query := `SELECT id, guild_id, channel_id, user_id, message, scheduled_at, created_at FROM reminders WHERE scheduled_at <= ?`
	rows, err := r.db.Query(query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []*domain.Reminder
	for rows.Next() {
		var rem domain.Reminder
		if err := rows.Scan(&rem.ID, &rem.GuildID, &rem.ChannelID, &rem.UserID, &rem.Message, &rem.ScheduledAt, &rem.CreatedAt); err != nil {
			return nil, err
		}
		reminders = append(reminders, &rem)
	}
	return reminders, nil
}

func (r *TaskRepository) DeleteReminder(id int64) error {
	_, err := r.db.Exec("DELETE FROM reminders WHERE id = ?", id)
	return err
}

func scanTask(scanner interface{ Scan(dest ...any) error }) (*domain.Task, error) {
	var t domain.Task
	var threadID, assigneeID, demoURL, keyInfo, sharedLink sql.NullString
	var deadline, completedAt sql.NullTime
	var bpm sql.NullFloat64
	var priority, status, phase string

	err := scanner.Scan(
		&t.ID, &t.GuildID, &t.ChannelID, &threadID, &t.Title, &t.Description,
		&priority, &status, &phase, &assigneeID, &deadline, &demoURL, &bpm,
		&keyInfo, &sharedLink, &t.CreatedAt, &t.UpdatedAt, &completedAt,
	)
	if err != nil {
		return nil, err
	}

	t.Priority = domain.TaskPriority(priority)
	t.Status = domain.TaskStatus(status)
	t.Phase = domain.DTMPhase(phase)

	if threadID.Valid {
		t.ThreadID = &threadID.String
	}
	if assigneeID.Valid {
		t.AssigneeID = &assigneeID.String
	}
	if demoURL.Valid {
		t.DemoURL = &demoURL.String
	}
	if keyInfo.Valid {
		t.KeyInfo = &keyInfo.String
	}
	if sharedLink.Valid {
		t.SharedLink = &sharedLink.String
	}
	if deadline.Valid {
		t.Deadline = &deadline.Time
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	if bpm.Valid {
		t.BPM = &bpm.Float64
	}

	return &t, nil
}

func (r *TaskRepository) CreateTask(t *domain.Task) (int64, error) {
	query := `
	INSERT INTO tasks (
		guild_id, channel_id, thread_id, title, description, priority, status, phase,
		assignee_id, deadline, demo_url, bpm, key_info, shared_link, created_at, updated_at, completed_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now

	var threadID, assigneeID, demoURL, keyInfo, sharedLink sql.NullString
	if t.ThreadID != nil {
		threadID = sql.NullString{String: *t.ThreadID, Valid: true}
	}
	if t.AssigneeID != nil {
		assigneeID = sql.NullString{String: *t.AssigneeID, Valid: true}
	}
	if t.DemoURL != nil {
		demoURL = sql.NullString{String: *t.DemoURL, Valid: true}
	}
	if t.KeyInfo != nil {
		keyInfo = sql.NullString{String: *t.KeyInfo, Valid: true}
	}
	if t.SharedLink != nil {
		sharedLink = sql.NullString{String: *t.SharedLink, Valid: true}
	}

	var deadline, completedAt sql.NullTime
	if t.Deadline != nil {
		deadline = sql.NullTime{Time: *t.Deadline, Valid: true}
	}
	if t.CompletedAt != nil {
		completedAt = sql.NullTime{Time: *t.CompletedAt, Valid: true}
	}

	var bpm sql.NullFloat64
	if t.BPM != nil {
		bpm = sql.NullFloat64{Float64: *t.BPM, Valid: true}
	}

	res, err := r.db.Exec(query,
		t.GuildID, t.ChannelID, threadID, t.Title, t.Description, string(t.Priority), string(t.Status), string(t.Phase),
		assigneeID, deadline, demoURL, bpm, keyInfo, sharedLink, t.CreatedAt, t.UpdatedAt, completedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	t.ID = id
	return id, nil
}

func (r *TaskRepository) GetTask(id int64) (*domain.Task, error) {
	query := `
	SELECT id, guild_id, channel_id, thread_id, title, description, priority, status, phase,
	       assignee_id, deadline, demo_url, bpm, key_info, shared_link, created_at, updated_at, completed_at
	FROM tasks WHERE id = ?
	`
	row := r.db.QueryRow(query, id)
	return scanTask(row)
}

func (r *TaskRepository) UpdateTask(t *domain.Task) error {
	query := `
	UPDATE tasks SET
		channel_id = ?, thread_id = ?, title = ?, description = ?, priority = ?, status = ?, phase = ?,
		assignee_id = ?, deadline = ?, demo_url = ?, bpm = ?, key_info = ?, shared_link = ?, updated_at = ?, completed_at = ?
	WHERE id = ?
	`
	t.UpdatedAt = time.Now()

	var threadID, assigneeID, demoURL, keyInfo, sharedLink sql.NullString
	if t.ThreadID != nil {
		threadID = sql.NullString{String: *t.ThreadID, Valid: true}
	}
	if t.AssigneeID != nil {
		assigneeID = sql.NullString{String: *t.AssigneeID, Valid: true}
	}
	if t.DemoURL != nil {
		demoURL = sql.NullString{String: *t.DemoURL, Valid: true}
	}
	if t.KeyInfo != nil {
		keyInfo = sql.NullString{String: *t.KeyInfo, Valid: true}
	}
	if t.SharedLink != nil {
		sharedLink = sql.NullString{String: *t.SharedLink, Valid: true}
	}

	var deadline, completedAt sql.NullTime
	if t.Deadline != nil {
		deadline = sql.NullTime{Time: *t.Deadline, Valid: true}
	}
	if t.CompletedAt != nil {
		completedAt = sql.NullTime{Time: *t.CompletedAt, Valid: true}
	}

	var bpm sql.NullFloat64
	if t.BPM != nil {
		bpm = sql.NullFloat64{Float64: *t.BPM, Valid: true}
	}

	_, err := r.db.Exec(query,
		t.ChannelID, threadID, t.Title, t.Description, string(t.Priority), string(t.Status), string(t.Phase),
		assigneeID, deadline, demoURL, bpm, keyInfo, sharedLink, t.UpdatedAt, completedAt, t.ID,
	)
	return err
}

func (r *TaskRepository) DeleteTask(id int64) error {
	_, err := r.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

// ListTasks は未完了タスク（または指定ステータスのタスク）を優先度・期限順で取得します
func (r *TaskRepository) ListTasks(guildID string, status *domain.TaskStatus) ([]*domain.Task, error) {
	baseQuery := `
	SELECT id, guild_id, channel_id, thread_id, title, description, priority, status, phase,
	       assignee_id, deadline, demo_url, bpm, key_info, shared_link, created_at, updated_at, completed_at
	FROM tasks
	WHERE guild_id = ?
	`
	var rows *sql.Rows
	var err error

	orderClause := `
	ORDER BY 
		CASE priority 
			WHEN 'High' THEN 1 
			WHEN 'Medium' THEN 2 
			WHEN 'Low' THEN 3 
			ELSE 4 
		END,
		deadline ASC, created_at ASC
	`

	if status != nil {
		rows, err = r.db.Query(baseQuery+" AND status = ? "+orderClause, guildID, string(*status))
	} else {
		rows, err = r.db.Query(baseQuery+" AND status != ? "+orderClause, guildID, string(domain.StatusDone))
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// GetUnremindedDayTasks は期限が24時間以内で、まだ1日前リマインドを送っていない未完了タスクを取得します
func (r *TaskRepository) GetUnremindedDayTasks(now time.Time) ([]*domain.Task, error) {
	target := now.Add(24 * time.Hour)
	query := `
	SELECT id, guild_id, channel_id, thread_id, title, description, priority, status, phase,
	       assignee_id, deadline, demo_url, bpm, key_info, shared_link, created_at, updated_at, completed_at
	FROM tasks
	WHERE status != ? AND deadline IS NOT NULL AND reminded_day = 0 AND deadline <= ? AND deadline > ?
	`
	rows, err := r.db.Query(query, string(domain.StatusDone), target, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepository) MarkRemindedDay(id int64) error {
	_, err := r.db.Exec("UPDATE tasks SET reminded_day = 1 WHERE id = ?", id)
	return err
}

// GetUnremindedHourTasks は期限が1時間以内で、まだ1時間前リマインドを送っていない未完了タスクを取得します
func (r *TaskRepository) GetUnremindedHourTasks(now time.Time) ([]*domain.Task, error) {
	target := now.Add(1 * time.Hour)
	query := `
	SELECT id, guild_id, channel_id, thread_id, title, description, priority, status, phase,
	       assignee_id, deadline, demo_url, bpm, key_info, shared_link, created_at, updated_at, completed_at
	FROM tasks
	WHERE status != ? AND deadline IS NOT NULL AND reminded_hour = 0 AND deadline <= ? AND deadline > ?
	`
	rows, err := r.db.Query(query, string(domain.StatusDone), target, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepository) MarkRemindedHour(id int64) error {
	_, err := r.db.Exec("UPDATE tasks SET reminded_hour = 1 WHERE id = ?", id)
	return err
}

// ListAllOverdueTasks は期限切れの未完了タスク一覧を取得します
func (r *TaskRepository) ListAllOverdueTasks(now time.Time) ([]*domain.Task, error) {
	query := `
	SELECT id, guild_id, channel_id, thread_id, title, description, priority, status, phase,
	       assignee_id, deadline, demo_url, bpm, key_info, shared_link, created_at, updated_at, completed_at
	FROM tasks
	WHERE status != ? AND deadline IS NOT NULL AND deadline < ?
	ORDER BY guild_id ASC, deadline ASC
	`
	rows, err := r.db.Query(query, string(domain.StatusDone), now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// GetMonthlyReport は指定年月の完了タスク一覧を取得します（レポート出力用）
func (r *TaskRepository) GetMonthlyReport(guildID string, year int, month time.Month) ([]*domain.Task, error) {
	startDest := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	endDest := startDest.AddDate(0, 1, 0)

	query := `
	SELECT id, guild_id, channel_id, thread_id, title, description, priority, status, phase,
	       assignee_id, deadline, demo_url, bpm, key_info, shared_link, created_at, updated_at, completed_at
	FROM tasks
	WHERE guild_id = ? AND status = ? AND completed_at >= ? AND completed_at < ?
	ORDER BY completed_at ASC
	`
	rows, err := r.db.Query(query, guildID, string(domain.StatusDone), startDest, endDest)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
