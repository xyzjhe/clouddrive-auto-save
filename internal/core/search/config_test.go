// internal/core/search/config_test.go
package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&Setting{})
	require.NoError(t, err)
	return db
}

type Setting struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"type:text"`
}

func (Setting) TableName() string { return "settings" }

func TestLoadConfig_Empty(t *testing.T) {
	db := setupTestDB(t)
	config, err := LoadConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "", config.CloudSaver.Server)
	assert.Equal(t, "", config.PanSou.Server)
}

func TestSaveAndLoadConfig(t *testing.T) {
	db := setupTestDB(t)
	config := &SearchConfig{
		CloudSaver: CloudSaverConfig{
			Server:   "http://localhost:8080",
			Username: "admin",
			Password: "pass123",
		},
		PanSou: PanSouConfig{
			Server: "https://so.252035.xyz",
		},
	}
	err := SaveConfig(db, config)
	require.NoError(t, err)

	loaded, err := LoadConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", loaded.CloudSaver.Server)
	assert.Equal(t, "admin", loaded.CloudSaver.Username)
	assert.Equal(t, "pass123", loaded.CloudSaver.Password)
	assert.Equal(t, "https://so.252035.xyz", loaded.PanSou.Server)
}

func TestSaveConfig_TokenUpdate(t *testing.T) {
	db := setupTestDB(t)
	config := &SearchConfig{
		CloudSaver: CloudSaverConfig{
			Server: "http://localhost:8080",
			Token:  "old-token",
		},
	}
	err := SaveConfig(db, config)
	require.NoError(t, err)

	// 更新 token
	config.CloudSaver.Token = "new-token"
	err = SaveConfig(db, config)
	require.NoError(t, err)

	loaded, err := LoadConfig(db)
	require.NoError(t, err)
	assert.Equal(t, "new-token", loaded.CloudSaver.Token)
}
