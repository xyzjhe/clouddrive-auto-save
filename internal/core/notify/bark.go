package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/zcq/clouddrive-auto-save/internal/utils"
)

// BarkNotifier Bark 通知渠道，通过统一 Notifier 接口管理
type BarkNotifier struct {
	server       string
	deviceKey    string
	icon         string
	archive      string
	successLevel string
	successSound string
	failureLevel string
	failureSound string
	client       *http.Client
}

// NewBarkNotifier 创建 Bark 通知渠道
func NewBarkNotifier() *BarkNotifier {
	return &BarkNotifier{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name 返回通知渠道名称
func (n *BarkNotifier) Name() string {
	return "bark"
}

// Type 返回通知渠道类型
func (n *BarkNotifier) Type() NotifierType {
	return NotifierTypeBark
}

// Init 初始化 Bark 通知渠道
func (n *BarkNotifier) Init(config map[string]interface{}) error {
	server, _ := config["server"].(string)
	if server == "" {
		server = "https://api.day.app"
	}

	deviceKey, _ := config["device_key"].(string)
	if deviceKey == "" {
		return fmt.Errorf("Bark device_key 不能为空")
	}

	// 可选配置项，带默认值
	icon, _ := config["icon"].(string)
	archive := "true"
	if v, ok := config["archive"].(string); ok && v != "" {
		archive = v
	}

	successLevel := "active"
	if v, ok := config["success_level"].(string); ok && v != "" {
		successLevel = v
	}
	successSound := "birdsong.caf"
	if v, ok := config["success_sound"].(string); ok && v != "" && v != "default" {
		successSound = v
	}
	failureLevel := "timeSensitive"
	if v, ok := config["failure_level"].(string); ok && v != "" {
		failureLevel = v
	}
	failureSound := "alarm.caf"
	if v, ok := config["failure_sound"].(string); ok && v != "" && v != "default" {
		failureSound = v
	}

	n.server = server
	n.deviceKey = deviceKey
	n.icon = icon
	n.archive = archive
	n.successLevel = successLevel
	n.successSound = successSound
	n.failureLevel = failureLevel
	n.failureSound = failureSound
	return nil
}

// Send 发送 Bark 通知
func (n *BarkNotifier) Send(ctx context.Context, message *Message) error {
	// 根据消息级别选择 Bark 级别和铃声
	level := n.successLevel
	sound := n.successSound
	if message.Level == LevelError || message.Level == LevelWarning {
		level = n.failureLevel
		sound = n.failureSound
	}

	return sendBarkDirectWithContext(ctx, n.server, n.deviceKey, message.Title, message.Content, level, sound, n.icon, n.archive)
}

// Test 测试 Bark 通知渠道
func (n *BarkNotifier) Test(ctx context.Context) error {
	return n.Send(ctx, &Message{
		Title:   "UCAS 测试通知",
		Content: "这是一条测试消息，用于验证 Bark 推送配置是否正确。",
		Level:   LevelInfo,
	})
}

// Close 关闭通知渠道
func (n *BarkNotifier) Close() error {
	return nil
}

// BarkPayload Bark 推送请求载荷
type BarkPayload struct {
	DeviceKey string `json:"device_key"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	Level     string `json:"level,omitempty"`
	Badge     int    `json:"badge,omitempty"`
	Sound     string `json:"sound,omitempty"`
	Icon      string `json:"icon,omitempty"`
	Group     string `json:"group,omitempty"`
	URL       string `json:"url,omitempty"`
	IsArchive int    `json:"isArchive"`
}

// sendBarkDirectWithContext 直接通过提供的服务器和 Key 发送推送（支持 context）
func sendBarkDirectWithContext(ctx context.Context, server, key, title, body, level, sound, icon, archive string) error {
	if server == "" {
		server = "https://api.day.app"
	}
	if key == "" {
		return fmt.Errorf("bark device key is empty")
	}

	// 处理默认铃声
	if sound == "default" {
		sound = ""
	}

	isArchive := 1
	if archive == "false" {
		isArchive = 0
	}

	payload := BarkPayload{
		DeviceKey: key,
		Title:     title,
		Body:      body,
		Level:     level,
		Sound:     sound,
		Icon:      icon,
		Group:     "UCAS",
		IsArchive: isArchive,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	slog.Debug("Bark 推送请求", "url", fmt.Sprintf("%s/push", server), "body", string(jsonData))

	// 构造推送 URL
	pushURL := fmt.Sprintf("%s/push", server)
	req, err := http.NewRequestWithContext(ctx, "POST", pushURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bark push failed with status: %d", resp.StatusCode)
	}

	slog.Debug("Bark 推送成功", "title", title)
	return nil
}

// SendBarkDirect 直接发送 Bark 推送（不检查开关，用于测试接口等场景）
func SendBarkDirect(server, key, title, body, level, sound, icon, archive string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return sendBarkDirectWithContext(ctx, server, key, title, body, level, sound, icon, archive)
}

// SendTaskNotification 发送任务完成通知（统一通过 Global Manager 发送到所有已启用渠道）
func SendTaskNotification(taskName string, status string, message string, files []string, duration time.Duration) {
	title := fmt.Sprintf("✅ 转存任务完成: %s", taskName)
	if status == "failed" {
		title = fmt.Sprintf("❌ 转存任务失败: %s", taskName)
	}

	body := fmt.Sprintf("%s\n执行耗时: %s", message, duration.Round(time.Second))
	if len(files) > 0 {
		fileList := ""
		maxFiles := 10
		for i, f := range files {
			if i >= maxFiles {
				fileList += fmt.Sprintf("\n... 等共 %d 个文件", len(files))
				break
			}
			fileList += fmt.Sprintf("\n- %s", f)
		}
		body = fmt.Sprintf("%s\n\n转存文件列表:%s", body, fileList)
	}

	msgLevel := LevelSuccess
	if status == "failed" {
		msgLevel = LevelError
	}

	notifyMsg := &Message{
		Title:   title,
		Content: body,
		Level:   msgLevel,
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := Global.Send(ctx, notifyMsg); err != nil {
			slog.Error("发送全局渠道通知失败", "error", err)
		}

		// 同步统计更新
		utils.BroadcastStatsUpdate()
	}()
}
