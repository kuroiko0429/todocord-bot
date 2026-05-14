package repository

import (
	"os"
	"testing"
	"time"

	"todocord/domain"
)

func TestTaskRepository_CRUD(t *testing.T) {
	dbPath := "test_todocord.db"
	defer os.Remove(dbPath)

	repo, err := NewTaskRepository(dbPath)
	if err != nil {
		t.Fatalf("初期化エラー: %v", err)
	}
	defer repo.Close()

	deadline := time.Now().Add(2 * time.Hour)
	bpm := 128.0
	demoURL := "https://example.com/demo.mp3"
	keyInfo := "Am"

	task := &domain.Task{
		GuildID:     "guild1",
		ChannelID:   "channel1",
		Title:       "テスト曲制作",
		Description: "最高のアレンジにする",
		Priority:    domain.PriorityHigh,
		Status:      domain.StatusTodo,
		Phase:       domain.PhaseArrange,
		Deadline:    &deadline,
		BPM:         &bpm,
		DemoURL:     &demoURL,
		KeyInfo:     &keyInfo,
	}

	id, err := repo.CreateTask(task)
	if err != nil || id == 0 {
		t.Fatalf("CreateTaskエラー: %v", err)
	}

	fetched, err := repo.GetTask(id)
	if err != nil {
		t.Fatalf("GetTaskエラー: %v", err)
	}
	if fetched.Title != task.Title || fetched.Phase != task.Phase {
		t.Errorf("取得したタスクの内容が一致しません: %+v", fetched)
	}

	fetched.Status = domain.StatusInProgress
	if err := repo.UpdateTask(fetched); err != nil {
		t.Fatalf("UpdateTaskエラー: %v", err)
	}

	updated, _ := repo.GetTask(id)
	if updated.Status != domain.StatusInProgress {
		t.Errorf("更新後のステータスが一致しません: %s", updated.Status)
	}

	list, err := repo.ListTasks("guild1", nil)
	if err != nil || len(list) != 1 {
		t.Fatalf("ListTasksエラー: %v, 長さ: %d", err, len(list))
	}

	if err := repo.DeleteTask(id); err != nil {
		t.Fatalf("DeleteTaskエラー: %v", err)
	}

	listAfter, _ := repo.ListTasks("guild1", nil)
	if len(listAfter) != 0 {
		t.Errorf("削除後のタスク一覧が空になりません")
	}
}
