package core

import (
	"context"
	"log"
	"time"

	"github.com/Ixecd/web3-blitz/internal/db"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	BTCRequiredConfirms = 6
	ETHRequiredConfirms = 12
)

type ConfirmChecker struct {
	queries *db.Queries
	btcRPC  *rpcclient.Client
	ethRPC  *ethclient.Client
}

func NewConfirmChecker(queries *db.Queries, btcRPC *rpcclient.Client, ethRPC *ethclient.Client) *ConfirmChecker {
	return &ConfirmChecker{queries: queries, btcRPC: btcRPC, ethRPC: ethRPC}
}

func (c *ConfirmChecker) Start(ctx context.Context) {
	log.Println("🔎 ConfirmChecker 已启动")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("⛔ ConfirmChecker 已停止")
			return
		case <-ticker.C:
			c.check(ctx)
		}
	}
}

func (c *ConfirmChecker) check(ctx context.Context) {
	deposits, err := c.queries.ListUnconfirmedDeposits(ctx)
	if err != nil {
		log.Printf("[ERROR] ConfirmChecker 查询未确认充值失败: %v", err)
		return
	}
	if len(deposits) == 0 {
		return
	}

	// 获取当前 BTC 块高
	var btcHeight int64
	if info, err := c.btcRPC.GetBlockChainInfo(); err == nil {
		btcHeight = int64(info.Blocks)
	}

	// 获取当前 ETH 块高
	var ethHeight int64
	if header, err := c.ethRPC.HeaderByNumber(ctx, nil); err == nil {
		ethHeight = header.Number.Int64()
	}

	for _, d := range deposits {
		var required int64
		var currentHeight int64

		switch d.Chain {
		case "btc":
			required = BTCRequiredConfirms
			currentHeight = btcHeight
		case "eth":
			required = ETHRequiredConfirms
			currentHeight = ethHeight
		default:
			continue
		}

		if currentHeight-d.Height >= required {
			if err := c.queries.UpdateDepositConfirmed(ctx, d.ID); err != nil {
				log.Printf("[ERROR] 更新confirmed失败 id=%d: %v", d.ID, err)
				continue
			}
			log.Printf("✅ 充值已确认 id=%d chain=%s txid=%s (块高差=%d)",
				d.ID, d.Chain, d.TxID, currentHeight-d.Height)
		}
	}
}
