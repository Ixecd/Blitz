package core

import (
	"context"
	"github.com/Ixecd/blitz/internal/wallet/types"
)

// WalletService 交易所钱包统一核心接口（以后加 TRON、SOL 都走这里）
type WalletService interface {
	// 生成充值地址
	GenerateDepositAddress(ctx context.Context, userID string, chain types.Chain) (types.AddressResponse, error)

	// 查询地址余额
	GetBalance(ctx context.Context, address string, chain types.Chain) (types.BalanceResponse, error)

	// 获取 HD 钱包管理器（用于提币签名等）
	GetHDWallet() HDWallet
}