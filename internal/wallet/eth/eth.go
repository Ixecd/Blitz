package eth

import (
	"context"
	"fmt"

	"github.com/Ixecd/web3-blitz/internal/wallet/core"
	"github.com/Ixecd/web3-blitz/internal/wallet/types"
)

type ETHWallet struct {
	hdWallet *core.HDWallet
}

func NewETHWallet(hd *core.HDWallet) *ETHWallet {
	return &ETHWallet{hdWallet: hd}
}

// GenerateDepositAddress 生成 ETH 充值地址
func (w *ETHWallet) GenerateDepositAddress(ctx context.Context, userID string, chain types.Chain) (types.AddressResponse, error) {
	// 安全截取（防止 userID 太短）
	shortID := userID
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}

	address := fmt.Sprintf("0x%s-test-address-%s", shortID, chain)
	path := fmt.Sprintf("m/44'/60'/0'/0/%s", userID)

	return types.AddressResponse{
		Address: address,
		Path:    path,
		UserID:  userID,
	}, nil
}

// GetBalance 查询余额（后面接真实 RPC）
func (w *ETHWallet) GetBalance(ctx context.Context, address string, chain types.Chain) (types.BalanceResponse, error) {
	return types.BalanceResponse{
		Address: address,
		Balance: 0.0,
		Chain:   chain,
	}, nil
}