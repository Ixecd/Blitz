package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/crypto/argon2"
	"golang.org/x/term"
)

// EncryptSeed 用 passphrase 加密 seed，输出 AES-256-GCM 密文 [salt(32)+nonce(12)+cipher]
func EncryptSeed(seedHex string, passphrase string) ([]byte, error) {
	seed, err := hex.DecodeString(seedHex)
	if err != nil {
		return nil, fmt.Errorf("seed 必须是 hex 字符串: %w", err)
	}

	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	key := argon2.IDKey([]byte(passphrase), salt, 1, 64*1024, 4, 32)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, seed, nil)

	// [salt(32) | nonce(12) | ciphertext]
	out := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	out = append(out, salt...)
	out = append(out, nonce...)
	out = append(out, ciphertext...)
	return out, nil
}

// DecryptSeedFile 读取加密种子文件，用 passphrase 解密，返回 hex seed
func DecryptSeedFile(path string, passphrase string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取加密种子文件 %s 失败: %w", path, err)
	}
	if len(data) < 32+12+1 {
		return "", fmt.Errorf("加密种子文件格式无效")
	}

	salt := data[0:32]
	nonce := data[32:44]
	ciphertext := data[44:]

	key := argon2.IDKey([]byte(passphrase), salt, 1, 64*1024, 4, 32)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("解密种子失败，passphrase 错误或文件损坏")
	}

	return hex.EncodeToString(plain), nil
}

// ReadPassphrase 优先读环境变量，否则交互式输入（隐藏回显）
func ReadPassphrase(envVar string) (string, error) {
	if p := os.Getenv(envVar); p != "" {
		return p, nil
	}
	fmt.Print("🔑 输入 HD 种子解密密码: ")
	bytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	if len(bytes) == 0 {
		return "", fmt.Errorf("密码不能为空")
	}
	return string(bytes), nil
}

// DerivePassphraseHash 从 passphrase 派生一个稳定的哈希用做运行时校验
func DerivePassphraseHash(passphrase string) string {
	h := sha256.Sum256([]byte(passphrase))
	return hex.EncodeToString(h[:])
}
