package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zcq/clouddrive-auto-save/internal/core"
	"github.com/zcq/clouddrive-auto-save/internal/crypto"
	"github.com/zcq/clouddrive-auto-save/internal/db"
)

// accountDTO 账号安全返回对象（排除 Cookie 和 AuthToken）
type accountDTO struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Platform      string    `json:"platform"`
	Nickname      string    `json:"nickname"`
	AccountName   string    `json:"account_name"`
	Status        int       `json:"status"`
	CapacityUsed  int64     `json:"capacity_used"`
	CapacityTotal int64     `json:"capacity_total"`
	VipName       string    `json:"vip_name"`
	LastCheck     time.Time `json:"last_check"`
}

// toAccountDTO 将 Account 实体转换为安全返回对象（排除凭据）
func toAccountDTO(a *db.Account) accountDTO {
	return accountDTO{
		ID: a.ID, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
		Platform: a.Platform, Nickname: a.Nickname, AccountName: a.AccountName,
		Status: a.Status, CapacityUsed: a.CapacityUsed, CapacityTotal: a.CapacityTotal,
		VipName: a.VipName, LastCheck: a.LastCheck,
	}
}

// accountInputDTO 账号输入数据传输对象，限制前端可写入的字段白名单
type accountInputDTO struct {
	Platform    string `json:"platform" binding:"required"`
	AccountName string `json:"account_name"`
	Cookie      string `json:"cookie"`
	AuthToken   string `json:"auth_token"`
}

// sanitizeCredentials 清理凭据中的空白字符和换行符
func sanitizeCredentials(dto *accountInputDTO) {
	dto.Cookie = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(dto.Cookie, "\n", ""), "\r", ""))
	dto.AuthToken = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(dto.AuthToken, "\n", ""), "\r", ""))
}

func listAccounts(c *gin.Context) {
	var accounts []db.Account
	db.DB.Find(&accounts)

	dtos := make([]accountDTO, len(accounts))
	for i := range accounts {
		dtos[i] = toAccountDTO(&accounts[i])
	}
	c.PureJSON(http.StatusOK, dtos)
}

func createAccount(c *gin.Context) {
	var dto accountInputDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sanitizeCredentials(&dto)

	// 加密凭据（如已启用）
	if crypto.Enabled() {
		dto.Cookie = crypto.Encrypt(dto.Cookie)
		dto.AuthToken = crypto.Encrypt(dto.AuthToken)
	}

	account := db.Account{
		Platform:    dto.Platform,
		AccountName: dto.AccountName,
		Cookie:      dto.Cookie,
		AuthToken:   dto.AuthToken,
	}

	slog.Info("添加账号", "name", account.AccountName, "platform", account.Platform)
	if err := db.DB.Create(&account).Error; err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
		return
	}

	// 自动执行一次校验
	if err := performAccountCheck(&account, c.Request.Context()); err != nil {
		slog.Error("添加账号后自动校验失败", "name", account.AccountName, "error", err)
	}

	c.PureJSON(http.StatusOK, toAccountDTO(&account))
}

func updateAccount(c *gin.Context) {
	id := c.Param("id")
	var account db.Account
	if err := db.DB.First(&account, id).Error; err != nil {
		c.PureJSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	var dto accountInputDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.PureJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sanitizeCredentials(&dto)

	// 加密凭据（如已启用）
	if crypto.Enabled() {
		dto.Cookie = crypto.Encrypt(dto.Cookie)
		dto.AuthToken = crypto.Encrypt(dto.AuthToken)
	}

	// 仅更新白名单字段，防止通过 JSON 注入覆盖 ID/Status 等敏感字段
	account.Platform = dto.Platform
	account.AccountName = dto.AccountName
	account.Cookie = dto.Cookie
	account.AuthToken = dto.AuthToken

	slog.Info("更新账号", "name", account.AccountName)
	if err := db.DB.Save(&account).Error; err != nil {
		c.PureJSON(http.StatusInternalServerError, gin.H{"error": "failed to update account"})
		return
	}

	// 自动执行一次校验
	if err := performAccountCheck(&account, c.Request.Context()); err != nil {
		slog.Error("更新账号后自动校验失败", "name", account.AccountName, "error", err)
	}

	c.PureJSON(http.StatusOK, toAccountDTO(&account))
}

func deleteAccount(c *gin.Context) {
	id := c.Param("id")

	// 检查是否有关联任务
	var count int64
	db.DB.Model(&db.Task{}).Where("account_id = ?", id).Count(&count)
	if count > 0 {
		slog.Error("尝试删除账号失败: 存在关联任务", "account_id", id, "task_count", count)
		c.PureJSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("该账号有关联的 %d 个任务，请先删除关联任务", count)})
		return
	}

	slog.Info("删除账号", "account_id", id)
	db.DB.Delete(&db.Account{}, id)
	c.PureJSON(http.StatusOK, gin.H{"message": "deleted"})
}

func checkAccount(c *gin.Context) {
	id := c.Param("id")
	var account db.Account
	if err := db.DB.First(&account, id).Error; err != nil {
		slog.Error("账号校验失败: 未找到", "account_id", id)
		c.PureJSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	if err := performAccountCheck(&account, c.Request.Context()); err != nil {
		slog.Error("账号校验失败", "account_id", id, "error", err)
		c.PureJSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "account": toAccountDTO(&account)})
		return
	}

	c.PureJSON(http.StatusOK, toAccountDTO(&account))
}

func performAccountCheck(account *db.Account, ctx context.Context) error {
	// 解密凭据供驱动使用
	if err := crypto.DecryptAccount(account); err != nil {
		slog.Error("解密账号凭据失败", "account_id", account.ID, "error", err)
		return fmt.Errorf("解密凭据失败: %w", err)
	}

	driver := core.GetDriver(account)
	if driver == nil {
		return fmt.Errorf("driver not found for platform: %s", account.Platform)
	}

	updatedAccount, err := driver.GetInfo(ctx)
	if err != nil {
		now := time.Now()
		account.Status = 0
		account.LastCheck = now
		db.DB.Model(account).Updates(map[string]interface{}{
			"status":     0,
			"last_check": now,
		})
		return err
	}

	err = db.DB.Model(account).Updates(map[string]interface{}{
		"nickname":       updatedAccount.Nickname,
		"account_name":   updatedAccount.AccountName,
		"status":         1,
		"capacity_used":  updatedAccount.CapacityUsed,
		"capacity_total": updatedAccount.CapacityTotal,
		"vip_name":       updatedAccount.VipName,
		"last_check":     time.Now(),
	}).Error
	if err != nil {
		return err
	}

	// 重新加载完整信息
	return db.DB.First(account, account.ID).Error
}
