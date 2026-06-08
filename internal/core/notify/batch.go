package notify

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// BatchResult 批量任务结果
type BatchResult struct {
	TaskName string
	Status   string
	Message  string
	Files    []string
	Duration time.Duration
}

// buildBatchTitle 构造批量通知标题
func buildBatchTitle(results []BatchResult) string {
	total := len(results)
	successCount := 0
	for _, r := range results {
		if r.Status == "success" {
			successCount++
		}
	}
	failCount := total - successCount

	if failCount == 0 {
		return fmt.Sprintf("📊 批量转存完成: 全部 %d 个任务成功", total)
	}
	return fmt.Sprintf("📊 批量转存完成: %d成功 / %d失败", successCount, failCount)
}

// buildBatchBody 构造批量通知正文
func buildBatchBody(results []BatchResult, totalDuration time.Duration) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("总耗时: %s\n", formatDuration(totalDuration)))

	for _, r := range results {
		icon := "✅"
		if r.Status == "failed" {
			icon = "❌"
		}
		sb.WriteString(fmt.Sprintf("\n%s %s - %s - 耗时 %s", icon, r.TaskName, r.Message, formatDuration(r.Duration)))
	}

	// 收集所有文件
	var allFiles []string
	for _, r := range results {
		allFiles = append(allFiles, r.Files...)
	}

	if len(allFiles) > 0 {
		sb.WriteString("\n\n转存文件列表:")
		maxFiles := 20
		for i, f := range allFiles {
			if i >= maxFiles {
				sb.WriteString(fmt.Sprintf("\n... 等共 %d 个文件", len(allFiles)))
				break
			}
			sb.WriteString(fmt.Sprintf("\n- %s", f))
		}
	}

	return sb.String()
}

// formatDuration 格式化耗时，秒级精度
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
}

// SendBatchNotification 发送批量执行汇总通知（统一通过 Global Manager 发送）
func SendBatchNotification(results []BatchResult, totalDuration time.Duration) {
	// 判断是否有失败任务
	hasFailure := false
	for _, r := range results {
		if r.Status == "failed" {
			hasFailure = true
			break
		}
	}

	msgLevel := LevelSuccess
	if hasFailure {
		msgLevel = LevelError
	}

	title := buildBatchTitle(results)
	body := buildBatchBody(results, totalDuration)

	notifyMsg := &Message{
		Title:   title,
		Content: body,
		Level:   msgLevel,
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := Global.Send(ctx, notifyMsg); err != nil {
			slog.Error("发送批量通知失败", "error", err)
		}
	}()
}
