// internal/core/plugin/interface.go
package plugin

import "context"

// Plugin 插件接口
type Plugin interface {
	// Name 返回插件名称
	Name() string

	// Version 返回插件版本
	Version() string

	// Description 返回插件描述
	Description() string

	// Init 初始化插件
	Init(config map[string]interface{}) error

	// Hooks 返回插件支持的生命周期钩子
	Hooks() []HookType

	// Execute 执行钩子
	Execute(ctx context.Context, hook HookType, data *HookData) error

	// Close 关闭插件，释放资源
	Close() error
}

// HookType 生命周期钩子类型
type HookType string

const (
	// HookTaskBefore 任务执行前
	HookTaskBefore HookType = "task_before"

	// HookTaskAfter 任务执行后
	HookTaskAfter HookType = "task_after"

	// HookRun 执行转存
	HookRun HookType = "run"
)

// HookData 钩子数据
type HookData struct {
	TaskID    uint
	TaskName  string
	Platform  string
	ShareURL  string
	SavePath  string
	Files     []FileInfo
	Error     error
	Result    *TaskResult
}

// FileInfo 文件信息
type FileInfo struct {
	Name string
	Size int64
	Path string
}

// TaskResult 任务结果
type TaskResult struct {
	Success   bool
	FileCount int
	TotalSize int64
	Duration  int64
	ErrorMsg  string
}

// PluginConfig 插件配置
type PluginConfig struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Hooks       []HookType             `json:"hooks"`
	Config      map[string]interface{} `json:"config"`
}
