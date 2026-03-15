package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tyler-smith/go-bip32"
)

type HDWallet struct {
	MasterKey *bip32.Key   // 大写 M，已导出
}


// NewHDWallet 使用 go-bip32 创建主密钥
func NewHDWallet(seed []byte) (*HDWallet, error) {
	master, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}
	return &HDWallet{MasterKey: master}, nil
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