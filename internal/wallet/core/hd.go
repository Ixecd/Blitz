package core

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/tyler-smith/go-bip32"
)

type HDWallet struct {
	MasterKey *bip32.Key
}

// NewHDWallet 从环境变量读取种子（推荐 hex 格式）
func NewHDWallet() (*HDWallet, error) {
	seedHex := os.Getenv("WALLET_HD_SEED")
	if seedHex == "" {
		// 合法的 hex 测试种子（32字节）
		seedHex = "746573742d736565642d666f722d6465762d6f6e6c792d3132333435363738393061626364656631323334353637383930616263646566"
		log.Println("⚠️  使用测试 seed（生产环境请务必设置 WALLET_HD_SEED 环境变量）")
	}

	seed, err := hex.DecodeString(seedHex)
	if err != nil {
		return nil, fmt.Errorf("WALLET_HD_SEED 必须是有效的 hex 字符串: %w", err)
	}

	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}

	return &HDWallet{MasterKey: masterKey}, nil
}

// DerivePath 支持 BIP44 路径字符串（如 m/44'/0'/0'/0/5）
func (w *HDWallet) DerivePath(path string) (*bip32.Key, error) {
	segments := strings.Split(strings.TrimPrefix(path, "m/"), "/")
	key := w.MasterKey

	for _, seg := range segments {
		indexStr := strings.TrimSuffix(seg, "'")
		index, err := strconv.ParseUint(indexStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid path segment: %s", seg)
		}

		child, err := key.NewChildKey(uint32(index))
		if err != nil {
			return nil, err
		}
		key = child
	}

	return key, nil
}