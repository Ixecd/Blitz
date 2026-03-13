package core

import (
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"   // ← 必须加这个
)

type HDWallet struct {
	masterKey *hdkeychain.ExtendedKey
}

func NewHDWallet(seed []byte) (*HDWallet, error) {
	// 使用 chaincfg.MainNetParams（regtest 也可以用这个）
	master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	return &HDWallet{masterKey: master}, nil
}