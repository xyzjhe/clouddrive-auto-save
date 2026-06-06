package core

import (
	"context"
	"net/http"
	"testing"

	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// mockDriver 是用于测试的 CloudDrive 桩实现
type mockDriver struct {
	account *db.Account
}

func (m *mockDriver) GetInfo(_ context.Context) (*db.Account, error) { return m.account, nil }
func (m *mockDriver) Login(_ context.Context) error                  { return nil }
func (m *mockDriver) ListFiles(_ context.Context, _ string) ([]FileInfo, error) {
	return nil, nil
}
func (m *mockDriver) CreateFolder(_ context.Context, _, _ string) (*FileInfo, error) {
	return nil, nil
}
func (m *mockDriver) DeleteFile(_ context.Context, _ string) error { return nil }
func (m *mockDriver) ParseShare(_ context.Context, _, _, _ string) ([]FileInfo, error) {
	return nil, nil
}
func (m *mockDriver) PrepareTargetPath(_ context.Context, _ string) (string, error) {
	return "", nil
}
func (m *mockDriver) RenameFile(_ context.Context, _, _ string) error { return nil }
func (m *mockDriver) SaveFileTo(_ context.Context, _, _ string) error { return nil }
func (m *mockDriver) SaveLink(_ context.Context, _, _, _ string, _ []string, _ string) error {
	return nil
}

func mockFactory(account *db.Account) CloudDrive {
	return &mockDriver{account: account}
}

// TestRegisterDriverAndGetDriver 测试注册后能通过 GetDriver 获取驱动实例
func TestRegisterDriverAndGetDriver(t *testing.T) {
	const platform = "test_mock"

	RegisterDriver(platform, mockFactory)

	account := &db.Account{Platform: platform, Nickname: "测试账号"}
	d := GetDriver(account)
	if d == nil {
		t.Fatalf("GetDriver(%q) 返回 nil，期望非 nil", platform)
	}

	md, ok := d.(*mockDriver)
	if !ok {
		t.Fatalf("GetDriver 返回的类型不是 *mockDriver")
	}
	if md.account.Nickname != "测试账号" {
		t.Errorf("驱动持有的账号昵称 = %q，期望 %q", md.account.Nickname, "测试账号")
	}
}

// TestGetDriverUnregisteredPlatform 测试未注册平台返回 nil
func TestGetDriverUnregisteredPlatform(t *testing.T) {
	account := &db.Account{Platform: "nonexistent_platform"}
	d := GetDriver(account)
	if d != nil {
		t.Errorf("GetDriver 对未注册平台应返回 nil，实际返回 %v", d)
	}
}

// TestGetDriverByURL 测试各种 URL 格式到平台的映射
// 注意：由于 core_test 无法导入 cloud139/quark 子包（循环依赖），
// 这里使用 mockFactory 注册到 "139" 和 "quark" 键来验证 URL 路由逻辑。
func TestGetDriverByURL(t *testing.T) {
	// 注册 mock 工厂到真实平台键，覆盖可能存在的 init() 注册
	RegisterDriver("139", mockFactory)
	RegisterDriver("quark", mockFactory)

	tests := []struct {
		name    string
		url     string
		wantNil bool
	}{
		{
			name:    "夸克标准分享链接",
			url:     "https://pan.quark.cn/s/abc123",
			wantNil: false,
		},
		{
			name:    "移动云盘 cloud.139.com 域名",
			url:     "https://cloud.139.com/web/xxx",
			wantNil: false,
		},
		{
			name:    "移动云盘 yun.139.com 域名",
			url:     "https://yun.139.com/xxx",
			wantNil: false,
		},
		{
			name:    "移动云盘 caiyun.139.com 域名",
			url:     "https://caiyun.139.com/xxx",
			wantNil: false,
		},
		{
			name:    "无法识别的 URL 返回 nil",
			url:     "https://example.com",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := GetDriverByURL(tt.url)
			if tt.wantNil {
				if d != nil {
					t.Errorf("GetDriverByURL(%q) 期望 nil，实际返回 %T", tt.url, d)
				}
				return
			}
			if d == nil {
				t.Fatalf("GetDriverByURL(%q) 返回 nil，期望非 nil", tt.url)
			}
			if _, ok := d.(*mockDriver); !ok {
				t.Errorf("GetDriverByURL(%q) 返回类型不是 *mockDriver， got %T", tt.url, d)
			}
		})
	}
}

// TestGetDriverByURLPlatforms 测试 GetDriverByURL 返回的驱动实例携带正确的平台标识
func TestGetDriverByURLPlatforms(t *testing.T) {
	// 使用能记录平台信息的工厂来验证 URL -> 平台映射
	platformTracker := make(map[string]string)
	trackFactory := func(platform string) DriveFactory {
		return func(account *db.Account) CloudDrive {
			platformTracker[account.Platform] = "created"
			return &mockDriver{account: account}
		}
	}

	RegisterDriver("139", trackFactory("139"))
	RegisterDriver("quark", trackFactory("quark"))

	quarkURL := "https://pan.quark.cn/s/xyz"
	d := GetDriverByURL(quarkURL)
	if d == nil {
		t.Fatalf("GetDriverByURL(%q) 返回 nil", quarkURL)
	}
	if _, ok := platformTracker["quark"]; !ok {
		t.Error("夸克 URL 未触发 quark 工厂调用")
	}

	yun139URL := "https://yun.139.com/share/abc"
	d = GetDriverByURL(yun139URL)
	if d == nil {
		t.Fatalf("GetDriverByURL(%q) 返回 nil", yun139URL)
	}
	if _, ok := platformTracker["139"]; !ok {
		t.Error("移动云盘 URL 未触发 139 工厂调用")
	}
}

// TestHTTPTransportDefault 测试 HTTPTransport 默认值
func TestHTTPTransportDefault(t *testing.T) {
	if HTTPTransport == nil {
		t.Fatal("HTTPTransport 不应为 nil")
	}
	if HTTPTransport != http.DefaultTransport {
		t.Error("HTTPTransport 默认值不是 http.DefaultTransport")
	}
}
