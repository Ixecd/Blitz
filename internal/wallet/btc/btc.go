package btc

import (
	"context"
	"fmt"

	"github.com/Ixecd/web3-blitz/internal/wallet/core"
	"github.com/Ixecd/web3-blitz/internal/wallet/types"
)

type BTCWallet struct {
	hdWallet *core.HDWallet
}

func NewBTCWallet(hd *core.HDWallet) *BTCWallet {
	return &BTCWallet{hdWallet: hd}
}

// GenerateDepositAddress 生成 BTC 充值地址（BIP44 路径）
func (w *BTCWallet) GenerateDepositAddress(ctx context.Context, userID string, chain types.Chain) (types.AddressResponse, error) {
	// 安全截取（防止 userID 太短）
	shortID := userID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}

	address := fmt.Sprintf("bc1q%s-test-address-%s", shortID, chain)
	path := fmt.Sprintf("m/44'/0'/0'/0/%s", userID)

	return types.AddressResponse{
		Address: address,
		Path:    path,
		UserID:  userID,
	}, nil
}

// GetBalance 查询余额（后面接真实 RPC）
func (w *BTCWallet) GetBalance(ctx context.Context, address string, chain types.Chain) (types.BalanceResponse, error) {
	return types.BalanceResponse{
		Address: address,
		Balance: 0.0, // 后面会改成真实查询
		Chain:   chain,
	}, nil
}
