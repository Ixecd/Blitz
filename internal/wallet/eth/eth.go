package eth

import (
	"context"
	"fmt"

	"github.com/Ixecd/web3-blitz/internal/wallet/core"
	"github.com/Ixecd/web3-blitz/internal/wallet/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type ETHWallet struct {
	hdWallet *core.HDWallet
}

func NewETHWallet(hd *core.HDWallet) *ETHWallet {
	return &ETHWallet{hdWallet: hd}
}

// GenerateDepositAddress 生成真实 ETH 充值地址（BIP44）
func (w *ETHWallet) GenerateDepositAddress(ctx context.Context, userID string, chain types.Chain) (types.AddressResponse, error) {
	// 使用 userID 生成 index（简单哈希转 uint32）
	index := uint32(0)
	for _, c := range userID {
		index = index*31 + uint32(c)
	}

	path := fmt.Sprintf("m/44'/60'/0'/0/%d", index)

	// 真实派生
	childKey, err := w.hdWallet.DerivePath(path)
	if err != nil {
		return types.AddressResponse{}, err
	}

	// ETH 地址生成（私钥 → 公钥 → 地址）
	privKeyBytes := childKey.Key
	privKey, err := crypto.ToECDSA(privKeyBytes)
	if err != nil {
		return types.AddressResponse{}, err
	}

	address := crypto.PubkeyToAddress(privKey.PublicKey)

	return types.AddressResponse{
		Address: address.Hex(),
		Path:    path,
		UserID:  userID,
	}, nil
}

// GetBalance 查询余额（后面接真实 RPC）
func (w *ETHWallet) GetBalance(ctx context.Context, address string, chain types.Chain) (types.BalanceResponse, error) {
	// TODO: 后面接真实 RPC
	return types.BalanceResponse{
		Address: address,
		Balance: 0.0,
		Chain:   chain,
	}, nil
}