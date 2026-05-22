// internal/testutil/testutil.go
package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zcq/clouddrive-auto-save/internal/db"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB 创建测试用内存数据库
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(gormDB)
	require.NoError(t, err)

	return gormDB
}

// AssertJSONEqual 断言 JSON 相等
func AssertJSONEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	assert.Equal(t, expected, actual)
}

// RequireNoError 要求无错误
func RequireNoError(t *testing.T, err error) {
	t.Helper()
	require.NoError(t, err)
}

// AssertError 断言有错误
func AssertError(t *testing.T, err error) {
	t.Helper()
	assert.Error(t, err)
}
