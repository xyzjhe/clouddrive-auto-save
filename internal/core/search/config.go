// internal/core/search/config.go
package search

import (
	"fmt"

	"gorm.io/gorm"
)

// SearchConfig 搜索源配置
type SearchConfig struct {
	CloudSaver CloudSaverConfig `json:"cloudsaver"`
	PanSou     PanSouConfig     `json:"pansou"`
}

// CloudSaverConfig CloudSaver 配置
type CloudSaverConfig struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// PanSouConfig PanSou 配置
type PanSouConfig struct {
	Server string `json:"server"`
}

// configKeys 配置项在 Setting 表中的 key 列表
var configKeys = map[string]func(*SearchConfig) *string{
	"search.cloudsaver.server":   func(c *SearchConfig) *string { return &c.CloudSaver.Server },
	"search.cloudsaver.username": func(c *SearchConfig) *string { return &c.CloudSaver.Username },
	"search.cloudsaver.password": func(c *SearchConfig) *string { return &c.CloudSaver.Password },
	"search.cloudsaver.token":    func(c *SearchConfig) *string { return &c.CloudSaver.Token },
	"search.pansou.server":       func(c *SearchConfig) *string { return &c.PanSou.Server },
}

// searchSetting 模型引用（与 db 包中的 Setting 表名一致）
type searchSetting struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"type:text"`
}

func (searchSetting) TableName() string { return "settings" }

// LoadConfig 从 Setting 表加载搜索配置
func LoadConfig(db *gorm.DB) (*SearchConfig, error) {
	config := &SearchConfig{}
	for key, ptrFunc := range configKeys {
		var setting searchSetting
		result := db.Where("key = ?", key).First(&setting)
		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("读取配置 %s 失败: %w", key, result.Error)
		}
		if result.Error == nil {
			*ptrFunc(config) = setting.Value
		}
	}
	return config, nil
}

// SaveConfig 保存搜索配置到 Setting 表
func SaveConfig(db *gorm.DB, config *SearchConfig) error {
	for key, ptrFunc := range configKeys {
		value := *ptrFunc(config)
		setting := searchSetting{Key: key, Value: value}
		result := db.Where("key = ?", key).Assign(searchSetting{Value: value}).FirstOrCreate(&setting)
		if result.Error != nil {
			return fmt.Errorf("保存配置 %s 失败: %w", key, result.Error)
		}
	}
	return nil
}
