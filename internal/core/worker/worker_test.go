//go:build e2e

package worker

import (
	"testing"
	"time"

	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	testDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	// 设置全局 DB，因为 notify 包会用到
	db.DB = testDB
	testDB.AutoMigrate(&db.Account{}, &db.Task{}, &db.Setting{})
	return testDB
}

func TestManager_Execute(t *testing.T) {
	testDB := setupTestDB(t)
	m := NewManager(1, 10, testDB)

	// 注册 Mock 驱动
	core.RegisterDriver("mock", func(account *db.Account) core.CloudDrive {
		return &MockDriver{
			Files: []core.FileInfo{
				{ID: "f1", Name: "file1.mp4", UpdateTime: time.Now()},
				{ID: "f2", Name: "file2.mp4", UpdateTime: time.Now()},
			},
		}
	})

	account := db.Account{Platform: "mock", Nickname: "TestUser"}
	testDB.Create(&account)

	task := db.Task{
		AccountID: account.ID,
		Account:   account,
		Name:      "TestTask",
		ShareURL:  "http://share.com/1",
		SavePath:  "/test",
		Status:    "pending",
	}
	testDB.Create(&task)

	// 执行任务
	m.execute(Job{Task: &task})

	// 验证结果
	var updatedTask db.Task
	testDB.First(&updatedTask, task.ID)
	if updatedTask.Status != "success" {
		t.Errorf("expected task status success, got %s", updatedTask.Status)
	}
	if updatedTask.Percent != 100 {
		t.Errorf("expected task percent 100, got %d", updatedTask.Percent)
	}
}

func TestManager_Execute_SkipExisting(t *testing.T) {
	testDB := setupTestDB(t)
	m := NewManager(1, 10, testDB)

	core.RegisterDriver("mock_skip", func(account *db.Account) core.CloudDrive {
		return &MockDriver{
			Files: []core.FileInfo{
				{ID: "f1", Name: "file1.mp4", UpdateTime: time.Now()},
			},
		}
	})

	account := db.Account{Platform: "mock_skip", Nickname: "TestUser"}
	testDB.Create(&account)

	task := db.Task{
		AccountID: account.ID,
		Account:   account,
		Name:      "TestTask",
		ShareURL:  "http://share.com/1",
		SavePath:  "/test",
		Status:    "pending",
	}
	testDB.Create(&task)

	m.execute(Job{Task: &task})

	var updatedTask db.Task
	testDB.First(&updatedTask, task.ID)
	if updatedTask.Status != "success" {
		t.Errorf("expected success, got %s", updatedTask.Status)
	}
}

func TestManager_Execute_StartFileFilter(t *testing.T) {
	testDB := setupTestDB(t)
	m := NewManager(1, 10, testDB)

	now := time.Now()
	files := []core.FileInfo{
		{ID: "f1", Name: "old.mp4", UpdateTime: now.Add(-2 * time.Hour)},
		{ID: "f2", Name: "start.mp4", UpdateTime: now.Add(-1 * time.Hour)},
		{ID: "f3", Name: "new.mp4", UpdateTime: now},
	}

	var spy *MockDriver
	core.RegisterDriver("mock_startfile", func(account *db.Account) core.CloudDrive {
		spy = &MockDriver{ShareFiles: files}
		return spy
	})

	account := db.Account{Platform: "mock_startfile", Nickname: "TestUser"}
	testDB.Create(&account)

	task := db.Task{
		AccountID:    account.ID,
		Account:      account,
		Name:         "StartFileTask",
		ShareURL:     "http://share.com/1",
		SavePath:     "/test",
		StartFileID:  "f2", // 从 f2 开始
		ScheduleMode: "off",
	}
	testDB.Create(&task)

	m.execute(Job{Task: &task})

	if spy.SaveLinkCalls == 0 {
		t.Fatal("expected SaveLink to be called")
	}

	idSet := make(map[string]bool)
	for _, id := range spy.SavedFileIDs {
		idSet[id] = true
	}

	if !idSet["f2"] || !idSet["f3"] {
		t.Errorf("expected f2 and f3 to be saved, got %v", spy.SavedFileIDs)
	}
	if idSet["f1"] {
		t.Errorf("expected f1 to be filtered out, but it was saved")
	}
}

func TestManager_Execute_RegexFilter(t *testing.T) {
	testDB := setupTestDB(t)
	m := NewManager(1, 10, testDB)

	files := []core.FileInfo{
		{ID: "f1", Name: "movie.mp4", UpdateTime: time.Now()},
		{ID: "f2", Name: "info.txt", UpdateTime: time.Now()},
	}

	var spy *MockDriver
	core.RegisterDriver("mock_regex", func(account *db.Account) core.CloudDrive {
		spy = &MockDriver{ShareFiles: files}
		return spy
	})

	account := db.Account{Platform: "mock_regex", Nickname: "TestUser"}
	testDB.Create(&account)

	task := db.Task{
		AccountID: account.ID,
		Account:   account,
		Name:      "RegexTask",
		Pattern:   ".*\\.mp4$", // 仅匹配 mp4
		ShareURL:  "http://share.com/1",
		SavePath:  "/test",
	}
	testDB.Create(&task)

	m.execute(Job{Task: &task})

	if len(spy.SavedFileIDs) != 1 || spy.SavedFileIDs[0] != "f1" {
		t.Errorf("expected only f1 (mp4) to be saved, got %v", spy.SavedFileIDs)
	}
}

func TestManager_Execute_Deduplication_With_Renamer(t *testing.T) {
	testDB := setupTestDB(t)
	m := NewManager(1, 10, testDB)

	var spy *MockDriver
	core.RegisterDriver("mock_dedup", func(account *db.Account) core.CloudDrive {
		spy = &MockDriver{
			ShareFiles: []core.FileInfo{
				{ID: "f1", Name: "original.mp4", UpdateTime: time.Now()},
			},
			Files: []core.FileInfo{
				{ID: "existing_id", Name: "MyTask.mp4"}, // 模拟目标目录已存在重命名后的名字
			},
		}
		return spy
	})

	account := db.Account{Platform: "mock_dedup", Nickname: "TestUser"}
	testDB.Create(&account)

	task := db.Task{
		AccountID:   account.ID,
		Account:     account,
		Name:        "MyTask",
		Pattern:     "original",
		Replacement: "{TASKNAME}.{EXT}", // 预期重命名为 MyTask.mp4
		ShareURL:    "http://share.com/1",
		SavePath:    "/test",
	}
	testDB.Create(&task)

	m.execute(Job{Task: &task})

	if spy.SaveLinkCalls > 0 && len(spy.SavedFileIDs) > 0 {
		t.Errorf("expected file to be skipped due to deduplication, but SaveLink was called with %v", spy.SavedFileIDs)
	}
}

func TestIsFatalError_FatalPatterns(t *testing.T) {
	cases := []string{
		"链接失效",
		"该链接失效，请重新获取",
		"链接过期",
		"分享链接过期了",
		"提取码错误",
		"您输入的提取码错误",
		"提取码无效",
		"提取码无效，请检查",
		"分享已删除",
		"该分享已删除",
		"分享已过期",
		"该分享已过期，无法访问",
		"权限不足",
		"当前账号权限不足",
		"文件涉及违规内容",
		"该分享涉及违规",
		"好友已取消了分享",
		"该分享已取消了分享",
	}
	for _, msg := range cases {
		if !isFatalError(msg) {
			t.Errorf("isFatalError(%q) = false, 期望 true", msg)
		}
	}
}

func TestIsFatalError_ExactChecks(t *testing.T) {
	cases := []string{
		"不存在",
		"目标目录不存在",
		"cookie过期",
		"用户cookie过期请重新登录",
		"Cookie过期",
		"检测到Cookie过期",
		"token无效",
		"返回token无效",
		"token过期",
		"accesstoken过期",
		"Token无效",
		"返回Token无效",
		"Token过期",
		"AccessToken过期",
	}
	for _, msg := range cases {
		if !isFatalError(msg) {
			t.Errorf("isFatalError(%q) = false, 期望 true", msg)
		}
	}
}

func TestIsFatalError_NonFatal(t *testing.T) {
	cases := []string{
		"timeout",
		"网络超时",
		"connection refused",
		"普通错误信息",
		" dial tcp: lookup example.com: no such host",
		"重试后仍然失败",
		"解析分享失败: unexpected EOF",
		"",
	}
	for _, msg := range cases {
		if isFatalError(msg) {
			t.Errorf("isFatalError(%q) = true, 期望 false", msg)
		}
	}
}

func TestIsFatalError_BroadWordsNotFatal(t *testing.T) {
	// 宽泛词不带限定后缀，不应被判定为致命错误
	cases := []string{
		"token",
		"invalid token",
		"Token",
		"cookie",
		"Cookie",
		"expired",
	}
	for _, msg := range cases {
		if isFatalError(msg) {
			t.Errorf("isFatalError(%q) = true, 期望 false (宽泛词不应误判)", msg)
		}
	}
}
