package btc

import (
	"context"
	"fmt"

	"github.com/Ixecd/web3-blitz/internal/wallet/core"
	"github.com/Ixecd/web3-blitz/internal/wallet/types"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
)

type BTCWallet struct {
	hdWallet *core.HDWallet
}

func NewBTCWallet(hd *core.HDWallet) *BTCWallet {
	return &BTCWallet{hdWallet: hd}
}

// GenerateDepositAddress 生成真实 BTC 地址（BIP44）
func (w *BTCWallet) GenerateDepositAddress(ctx context.Context, userID string, chain types.Chain) (types.AddressResponse, error) {
	// 使用 userID 生成 index（简单 hash 转 uint32）
	index := uint32(0)
	for _, c := range userID {
		index = index*31 + uint32(c)
	}

	path := fmt.Sprintf("m/44'/0'/0'/0/%d", index)

	// 真实派生
	childKey, err := w.hdWallet.DerivePath(path)
	if err != nil {
		return types.AddressResponse{}, err
	}

	// 生成 bech32 地址（P2WPKH）
	addr, err := btcutil.NewAddressWitnessPubKeyHash(
		btcutil.Hash160(childKey.PublicKey().Key),
		&chaincfg.MainNetParams,
	)
	if err != nil {
		return types.AddressResponse{}, err
	}

	return types.AddressResponse{
		Address: addr.EncodeAddress(),
		Path:    path,
		UserID:  userID,
	}, nil
}

func (w *BTCWallet) GetBalance(ctx context.Context, address string, chain types.Chain) (types.BalanceResponse, error) {
	// TODO: 后面接真实 RPC 查询
	return types.BalanceResponse{
		Address: address,
		Balance: 0.0,
		Chain:   chain,
	}, nil
}