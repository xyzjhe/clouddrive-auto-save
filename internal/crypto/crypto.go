package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/zcq/clouddrive-auto-save/internal/db"
)

var aesGCM cipher.AEAD

// Init 初始化加密模块，keyHex 为 64 字符的 hex 编码 AES-256 密钥。
// 为空则禁用加密（向后兼容）。
func Init(keyHex string) error {
	if keyHex == "" {
		slog.Info("凭据加密未启用（UCAS_SECRET_KEY 未设置）")
		return nil
	}

	key, err := hexToBytes(keyHex)
	if err != nil || len(key) != 32 {
		return fmt.Errorf("无效的加密密钥：需要 64 字符的 hex 编码 AES-256 密钥")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("创建 AES cipher 失败: %w", err)
	}

	aesGCM, err = cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("创建 GCM 失败: %w", err)
	}

	slog.Info("凭据加密已启用")
	return nil
}

// Enabled 返回加密是否已启用
func Enabled() bool {
	return aesGCM != nil
}

// Encrypt 加密明文，返回 "base64(nonce):base64(ciphertext)" 格式
func Encrypt(plaintext string) string {
	if aesGCM == nil || plaintext == "" {
		return plaintext
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		slog.Error("生成 nonce 失败", "error", err)
		return plaintext
	}

	ciphertext := aesGCM.Seal(nil, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(nonce) + ":" + base64.StdEncoding.EncodeToString(ciphertext)
}

// Decrypt 解密 "base64(nonce):base64(ciphertext)" 格式的密文
func Decrypt(encoded string) (string, error) {
	if aesGCM == nil || encoded == "" {
		return encoded, nil
	}

	// 如果不包含冒号分隔符，视为明文（兼容旧数据）
	if !strings.Contains(encoded, ":") {
		return encoded, nil
	}

	parts := strings.SplitN(encoded, ":", 2)
	nonce, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("解码 nonce 失败: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("解码密文失败: %w", err)
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}

	return string(plaintext), nil
}

// IsEncrypted 检查字符串是否为加密格式（base64(nonce):base64(ciphertext)）
// 仅当冒号前后都是合法 base64 且 nonce 长度为 16 字符（12 字节 AES-GCM nonce）时才判定为加密
func IsEncrypted(s string) bool {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	// AES-GCM nonce 经 base64 编码后固定为 16 字符
	nonce, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil || len(nonce) != 12 {
		return false
	}
	// 验证第二段也是合法 base64
	_, err = base64.StdEncoding.DecodeString(parts[1])
	return err == nil
}

// EncryptAccount 加密 Account 的凭据字段
func EncryptAccount(account *db.Account) {
	if account == nil {
		return
	}
	account.Cookie = Encrypt(account.Cookie)
	account.AuthToken = Encrypt(account.AuthToken)
}

// DecryptAccount 解密 Account 的凭据字段
func DecryptAccount(account *db.Account) error {
	if account == nil {
		return nil
	}

	cookie, err := Decrypt(account.Cookie)
	if err != nil {
		return fmt.Errorf("解密 Cookie 失败: %w", err)
	}
	account.Cookie = cookie

	authToken, err := Decrypt(account.AuthToken)
	if err != nil {
		return fmt.Errorf("解密 AuthToken 失败: %w", err)
	}
	account.AuthToken = authToken

	return nil
}

func hexToBytes(hex string) ([]byte, error) {
	if len(hex) != 64 {
		return nil, fmt.Errorf("hex 字符串长度必须为 64，实际为 %d", len(hex))
	}
	b := make([]byte, 32)
	for i := 0; i < 32; i++ {
		hi := hexCharToByte(hex[i*2])
		lo := hexCharToByte(hex[i*2+1])
		if hi == 0xff || lo == 0xff {
			return nil, fmt.Errorf("无效的 hex 字符")
		}
		b[i] = hi<<4 | lo
	}
	return b, nil
}

func hexCharToByte(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	default:
		return 0xff
	}
}
