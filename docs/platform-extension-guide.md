# 平台扩展指南

本文档指导开发者如何为 UCAS 添加新的云盘平台支持。

## 概述

UCAS 使用驱动工厂模式管理云盘平台。每个平台通过实现 `CloudDrive` 接口来提供统一的操作抽象。

## 步骤

### 1. 创建驱动目录

```bash
mkdir -p internal/core/driver_<platform>
```

### 2. 实现 CloudDrive 接口

创建 `client.go` 文件，实现以下接口：

```go
type CloudDrive interface {
    // GetInfo 获取账号信息
    GetInfo() (*AccountInfo, error)

    // Login 登录验证
    Login() error

    // ListFiles 列出目录文件
    ListFiles(path string) ([]FileInfo, error)

    // CreateFolder 创建文件夹
    CreateFolder(parentPath, name string) error

    // DeleteFile 删除文件
    DeleteFile(path string) error

    // ParseShare 解析分享链接
    ParseShare(shareURL, passCode string) (*ShareInfo, error)

    // SaveLink 保存分享链接
    SaveLink(shareID, fileID, targetPath string) error

    // RenameFile 重命名文件
    RenameFile(path, newName string) error
}
```

### 3. 注册驱动

在 `client.go` 的 `init()` 函数中注册驱动：

```go
func init() {
    core.RegisterDriver("<platform>", func(account *db.Account) (core.CloudDrive, error) {
        return &Client{
            account: account,
        }, nil
    })
}
```

### 4. 导入驱动

在 `internal/api/router.go` 中添加导入：

```go
_ "github.com/zcq/clouddrive-auto-save/internal/core/driver_<platform>"
```

### 5. 编写测试

创建 `client_test.go`，编写单元测试和集成测试。

### 6. 更新文档

更新 `README.md`，添加新平台的说明。

## 参考实现

- `internal/core/quark/` - 夸克网盘驱动
- `internal/core/cloud139/` - 移动云盘驱动

## 注意事项

1. **错误处理**：映射平台特定错误码到统一错误类型
2. **速率限制**：遵守平台 API 调用频率限制
3. **认证方式**：支持 Cookie / Token / OAuth 等认证方式
4. **日志记录**：使用 slog 记录关键操作和错误
5. **测试覆盖**：编写完整的单元测试和 E2E 测试

## 示例代码

以下是一个简化的驱动实现示例：

```go
package driver_example

import (
    "context"
    "fmt"
    "log/slog"

    "github.com/zcq/clouddrive-auto-save/internal/core"
    "github.com/zcq/clouddrive-auto-save/internal/db"
)

// Client 示例驱动客户端
type Client struct {
    account *db.Account
}

func init() {
    core.RegisterDriver("example", func(account *db.Account) (core.CloudDrive, error) {
        return &Client{
            account: account,
        }, nil
    })
}

// GetInfo 获取账号信息
func (c *Client) GetInfo(ctx context.Context) (*core.AccountInfo, error) {
    // TODO: 实现获取账号信息逻辑
    return &core.AccountInfo{
        Nickname:      "示例账号",
        AccountName:   "example@example.com",
        Status:        1,
        CapacityUsed:  1024 * 1024 * 1024, // 1GB
        CapacityTotal: 10 * 1024 * 1024 * 1024, // 10GB
    }, nil
}

// Login 登录验证
func (c *Client) Login(ctx context.Context) error {
    // TODO: 实现登录验证逻辑
    slog.Info("示例驱动登录成功", "account", c.account.Nickname)
    return nil
}

// ListFiles 列出目录文件
func (c *Client) ListFiles(ctx context.Context, path string) ([]core.FileInfo, error) {
    // TODO: 实现列出文件逻辑
    return []core.FileInfo{}, nil
}

// CreateFolder 创建文件夹
func (c *Client) CreateFolder(ctx context.Context, parentPath, name string) error {
    // TODO: 实现创建文件夹逻辑
    return nil
}

// DeleteFile 删除文件
func (c *Client) DeleteFile(ctx context.Context, path string) error {
    // TODO: 实现删除文件逻辑
    return nil
}

// ParseShare 解析分享链接
func (c *Client) ParseShare(ctx context.Context, shareURL, passCode string) (*core.ShareInfo, error) {
    // TODO: 实现解析分享链接逻辑
    return &core.ShareInfo{}, nil
}

// SaveLink 保存分享链接
func (c *Client) SaveLink(ctx context.Context, shareID, fileID, targetPath string) error {
    // TODO: 实现保存分享链接逻辑
    return nil
}

// RenameFile 重命名文件
func (c *Client) RenameFile(ctx context.Context, path, newName string) error {
    // TODO: 实现重命名文件逻辑
    return nil
}
```

## 平台特定注意事项

### 115 网盘
- 支持 Cookie 和 Token 认证
- API 调用频率限制较严格
- 支持秒传功能

### 天翼云盘
- 支持手机号和邮箱登录
- 需要处理验证码
- 支持家庭云功能

### 百度网盘
- 支持 OAuth 2.0 认证
- API 调用需要签名
- 支持超级会员加速

## 测试指南

### 单元测试

```go
func TestClient_GetInfo(t *testing.T) {
    client := &Client{
        account: &db.Account{
            Platform: "example",
            Nickname: "测试账号",
        },
    }

    info, err := client.GetInfo(context.Background())
    if err != nil {
        t.Fatalf("GetInfo failed: %v", err)
    }

    if info.Nickname != "示例账号" {
        t.Errorf("Expected nickname '示例账号', got '%s'", info.Nickname)
    }
}
```

### 集成测试

```go
func TestClient_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // 使用真实凭证进行测试
    client := &Client{
        account: &db.Account{
            Platform:   "example",
            Nickname:   "集成测试账号",
            Credential: "test_credential",
        },
    }

    // 测试登录
    if err := client.Login(context.Background()); err != nil {
        t.Fatalf("Login failed: %v", err)
    }

    // 测试获取信息
    info, err := client.GetInfo(context.Background())
    if err != nil {
        t.Fatalf("GetInfo failed: %v", err)
    }

    t.Logf("Account info: %+v", info)
}
```

## 发布检查清单

- [ ] 实现 CloudDrive 接口的所有方法
- [ ] 编写完整的单元测试
- [ ] 编写集成测试（可选）
- [ ] 更新 README.md 文档
- [ ] 添加平台图标（可选）
- [ ] 提交 Pull Request

## 常见问题

### Q: 如何处理平台 API 变更？
A: 在驱动中实现版本检测和兼容性处理，必要时发布新版本。

### Q: 如何处理网络超时？
A: 在 HTTP 客户端中设置合理的超时时间，并实现重试机制。

### Q: 如何处理并发请求？
A: 使用互斥锁或限流器控制并发请求数量，避免触发平台限制。

## 联系方式

如有疑问，请提交 Issue 或联系维护者。
