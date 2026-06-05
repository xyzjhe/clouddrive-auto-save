package scheduler

import (
	"testing"
	"time"
)

func TestScheduler_AddAndRemoveTask(t *testing.T) {
	s := New(nil)
	s.Start()
	defer s.Stop()

	s.UpdateTask(1, "custom", "0 * * * * *")

	if len(s.CustomEntryIDs) != 1 {
		t.Errorf("Expected 1 task in scheduler, got %d", len(s.CustomEntryIDs))
	}

	s.RemoveTask(1)
	if len(s.CustomEntryIDs) != 0 {
		t.Errorf("Expected 0 tasks after removal, got %d", len(s.CustomEntryIDs))
	}
}

// --- isRunDayAllowed 测试 ---

func TestIsRunDayAllowed_EmptyString(t *testing.T) {
	if !isRunDayAllowed("") {
		t.Error("空字符串应返回 true")
	}
}

func TestIsRunDayAllowed_EmptyArray(t *testing.T) {
	if !isRunDayAllowed("[]") {
		t.Error("'[]' 应返回 true")
	}
}

func TestIsRunDayAllowed_Weekdays(t *testing.T) {
	today := int(time.Now().Weekday())
	if today == 0 {
		today = 7 // Go 的 Sunday=0 转换为 7
	}

	allowed := isRunDayAllowed("[1,2,3,4,5]")
	if today >= 1 && today <= 5 {
		// 工作日应允许
		if !allowed {
			t.Errorf("今天是周%d（工作日），[1,2,3,4,5] 应返回 true", today)
		}
	} else {
		// 周末应不允许
		if allowed {
			t.Errorf("今天是周%d（周末），[1,2,3,4,5] 应返回 false", today)
		}
	}
}

func TestIsRunDayAllowed_Weekend(t *testing.T) {
	today := int(time.Now().Weekday())
	if today == 0 {
		today = 7
	}

	allowed := isRunDayAllowed("[6,7]")
	if today == 6 || today == 7 {
		// 周末应允许
		if !allowed {
			t.Errorf("今天是周%d（周末），[6,7] 应返回 true", today)
		}
	} else {
		// 工作日应不允许
		if allowed {
			t.Errorf("今天是周%d（工作日），[6,7] 应返回 false", today)
		}
	}
}

func TestIsRunDayAllowed_InvalidJSON(t *testing.T) {
	if !isRunDayAllowed("not-json") {
		t.Error("无效 JSON 应容错返回 true")
	}
}

func TestIsRunDayAllowed_InvalidJSONPartial(t *testing.T) {
	if !isRunDayAllowed("[1,2,broken") {
		t.Error("不完整的 JSON 应容错返回 true")
	}
}

// --- ValidateCron 测试 ---

func TestValidateCron_ValidExpressions(t *testing.T) {
	valid := []string{
		"0 * * * * *",    // 每分钟
		"*/5 * * * * *",  // 每5秒
		"0 0 * * * *",    // 每小时
		"0 30 9 * * 1-5", // 工作日 9:30
		"0 0 0 1 1 *",    // 每年1月1日
	}
	for _, expr := range valid {
		if err := ValidateCron(expr); err != nil {
			t.Errorf("合法 cron 表达式 '%s' 不应报错: %v", expr, err)
		}
	}
}

func TestValidateCron_InvalidExpressions(t *testing.T) {
	invalid := []string{
		"",          // 空字符串
		"* * * *",   // 字段不足（需要6字段）
		"abc",       // 非法语法
		"0 0 0 0 0", // 字段不足
	}
	for _, expr := range invalid {
		if err := ValidateCron(expr); err == nil {
			t.Errorf("非法 cron 表达式 '%s' 应返回错误", expr)
		}
	}
}

// --- UpdateTask 测试 ---

func TestUpdateTask_OffModeRemovesSchedule(t *testing.T) {
	s := New(nil)
	s.Start()
	defer s.Stop()

	// 先添加一个 custom 调度
	s.UpdateTask(10, "custom", "0 * * * * *")
	if len(s.CustomEntryIDs) != 1 {
		t.Fatalf("添加后应有 1 个调度，实际 %d", len(s.CustomEntryIDs))
	}

	// mode="off" 应移除已有调度
	s.UpdateTask(10, "off", "")
	if _, exists := s.CustomEntryIDs[10]; exists {
		t.Error("mode='off' 后任务 10 不应存在于 CustomEntryIDs")
	}
	if len(s.CustomEntryIDs) != 0 {
		t.Errorf("mode='off' 后 CustomEntryIDs 应为空，实际 %d 个", len(s.CustomEntryIDs))
	}
}

func TestUpdateTask_CustomModeAddsSchedule(t *testing.T) {
	s := New(nil)
	s.Start()
	defer s.Stop()

	s.UpdateTask(20, "custom", "*/10 * * * * *")
	if entryID, exists := s.CustomEntryIDs[20]; !exists {
		t.Error("mode='custom' + 合法 cron 应添加调度到 CustomEntryIDs")
	} else if entryID == 0 {
		t.Error("EntryID 不应为 0")
	}
}

func TestUpdateTask_CustomModeInvalidCron(t *testing.T) {
	s := New(nil)
	s.Start()
	defer s.Stop()

	// 非法 cron 表达式，AddFunc 会失败，不应添加到 map
	s.UpdateTask(30, "custom", "not-a-cron")
	if _, exists := s.CustomEntryIDs[30]; exists {
		t.Error("非法 cron 表达式不应添加调度到 CustomEntryIDs")
	}
}

func TestUpdateTask_OffModeOnNonexistentTask(t *testing.T) {
	s := New(nil)
	s.Start()
	defer s.Stop()

	// 对不存在的任务调用 off，不应 panic
	s.UpdateTask(999, "off", "")
	if len(s.CustomEntryIDs) != 0 {
		t.Errorf("对不存在的任务执行 off 后 CustomEntryIDs 应为空，实际 %d 个", len(s.CustomEntryIDs))
	}
}

func TestUpdateTask_OffThenCustomSameTask(t *testing.T) {
	s := New(nil)
	s.Start()
	defer s.Stop()

	// 添加 -> 移除 -> 再添加 同一个任务
	s.UpdateTask(40, "custom", "0 * * * * *")
	if len(s.CustomEntryIDs) != 1 {
		t.Fatalf("第一次添加后应有 1 个调度，实际 %d", len(s.CustomEntryIDs))
	}

	s.UpdateTask(40, "off", "")
	if len(s.CustomEntryIDs) != 0 {
		t.Fatalf("off 后应无调度，实际 %d 个", len(s.CustomEntryIDs))
	}

	s.UpdateTask(40, "custom", "*/30 * * * * *")
	if len(s.CustomEntryIDs) != 1 {
		t.Errorf("再次添加后应有 1 个调度，实际 %d", len(s.CustomEntryIDs))
	}
	if _, exists := s.CustomEntryIDs[40]; !exists {
		t.Error("任务 40 应存在于 CustomEntryIDs")
	}
}
