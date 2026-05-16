package core

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/tyler-smith/go-bip32"
)

type HDWallet struct {
	MasterKey *bip32.Key
}

// NewHDWallet 从加密的种子文件解密后初始化 HD 钱包
// seedFile 解密用 DecryptSeedFile，passphrase 从 ReadPassphrase 获取
func NewHDWallet(seedHex string) (*HDWallet, error) {
	seed, err := hex.DecodeString(seedHex)
	if err != nil {
		return nil, fmt.Errorf("seed 必须是有效的 hex 字符串: %w", err)
	}

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}

	return &HDWallet{MasterKey: masterKey}, nil
}

// NewHDWalletFromEncrypted 从加密种子文件和 passphrase 初始化
func NewHDWalletFromEncrypted(seedFile, passphrase string) (*HDWallet, error) {
	if _, err := os.Stat(seedFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("加密种子文件 %s 不存在，请先用 encrypt-seed 工具创建", seedFile)
	}

	seedHex, err := DecryptSeedFile(seedFile, passphrase)
	if err != nil {
		return nil, err
	}

	return NewHDWallet(seedHex)
}

// DerivePath 支持 BIP44 路径字符串（如 m/44'/0'/0'/0/5）
func (w *HDWallet) DerivePath(path string) (*bip32.Key, error) {
	return derivePathFromMaster(w.MasterKey, path)
}

func derivePathFromMaster(master *bip32.Key, path string) (*bip32.Key, error) {
	segments := parsePath(path)
	key := master
	for _, seg := range segments {
		child, err := key.NewChildKey(seg)
		if err != nil {
			return nil, err
		}
		key = child
	}
	return key, nil
}

func parsePath(path string) []uint32 {
	var segments []uint32
	path = path
	// 简单解析 m/44'/0'/0'/0/5 格式
	// 去掉前缀 "m/" 后按 "/" 分割
	rest := path
	if len(rest) > 1 && rest[0:2] == "m/" {
		rest = rest[2:]
	}
	if rest == "" {
		return segments
	}

	var current uint32
	for i := 0; i < len(rest); i++ {
		c := rest[i]
		if c >= '0' && c <= '9' {
			current = current*10 + uint32(c-'0')
		} else if c == '\'' {
			current |= 0x80000000 // hardened
		} else if c == '/' {
			segments = append(segments, current)
			current = 0
		}
	}
	segments = append(segments, current)
	return segments
}
