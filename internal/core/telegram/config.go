// internal/core/telegram/config.go
package telegram

// Config Telegram 配置
type Config struct {
	Enabled         bool    `json:"enabled"`
	BotToken        string  `json:"bot_token"`
	AllowedIDs      []int64 `json:"allowed_ids"`
	NotifyOnSuccess bool    `json:"notify_on_success"`
	NotifyOnFailure bool    `json:"notify_on_failure"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Enabled:         false,
		BotToken:        "",
		AllowedIDs:      []int64{},
		NotifyOnSuccess: true,
		NotifyOnFailure: true,
	}
}
