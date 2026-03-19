package core

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Ixecd/web3-blitz/internal/config"
	"github.com/Ixecd/web3-blitz/internal/db"
	"github.com/Ixecd/web3-blitz/internal/metrics"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	BTCRequiredConfirms = 6
	ETHRequiredConfirms = 12
	leaderKey           = "/blitz/leader/confirm-checker"
	leaseTTL            = 15 // 秒
)

type ConfirmChecker struct {
	queries    *db.Queries
	btcRPC     *config.BTCRPCHolder
	ethRPC     *config.ETHRPCHolder
	etcdClient *clientv3.Client
}

func NewConfirmChecker(queries *db.Queries, btcRPC *config.BTCRPCHolder, ethRPC *config.ETHRPCHolder, etcdClient *clientv3.Client) *ConfirmChecker {
	return &ConfirmChecker{
		queries:    queries,
		btcRPC:     btcRPC,
		ethRPC:     ethRPC,
		etcdClient: etcdClient,
	}
}

func (c *ConfirmChecker) Start(ctx context.Context) {
	log.Println("🔎 ConfirmChecker 已启动，竞选 Leader...")

	for {
		select {
		case <-ctx.Done():
			log.Println("⛔ ConfirmChecker 已停止")
			return
		default:
			c.campaign(ctx)
			// campaign 退出说明失去 Leader，等 3s 再重新竞选
			select {
			case <-ctx.Done():
				return
			case <-time.After(3 * time.Second):
			}
		}
	}
}

// campaign 竞选 Leader，成功后持续持有 lease 并执行 check
// lease 过期或 ctx 取消时退出
func (c *ConfirmChecker) campaign(ctx context.Context) {
	// 1. 创建 lease
	lease, err := c.etcdClient.Grant(ctx, leaseTTL)
	if err != nil {
		log.Printf("[WARN] ConfirmChecker 创建 lease 失败: %v", err)
		return
	}

	// 2. 原子竞选：key 不存在才写入
	txn := c.etcdClient.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(leaderKey), "=", 0)).
		Then(clientv3.OpPut(leaderKey, "1", clientv3.WithLease(lease.ID))).
		Else()

	resp, err := txn.Commit()
	if err != nil || !resp.Succeeded {
		// 竞选失败，撤销 lease，等待重试
		c.etcdClient.Revoke(ctx, lease.ID)
		if err == nil {
			log.Println("🔎 ConfirmChecker 竞选失败，当前非 Leader，等待中...")
		}
		return
	}

	log.Println("👑 ConfirmChecker 成为 Leader")

	// 3. 启动 lease 续期（keepalive），保持 Leader 身份
	keepAlive, err := c.etcdClient.KeepAlive(ctx, lease.ID)
	if err != nil {
		c.etcdClient.Revoke(ctx, lease.ID)
		log.Printf("[WARN] ConfirmChecker keepalive 失败: %v", err)
		return
	}

	// 4. 作为 Leader 跑 check 循环
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	defer c.etcdClient.Revoke(context.Background(), lease.ID)

	for {
		select {
		case <-ctx.Done():
			log.Println("⛔ ConfirmChecker Leader 退出")
			return
		case _, ok := <-keepAlive:
			if !ok {
				// keepalive channel 关闭，lease 过期，失去 Leader
				log.Println("⚠️  ConfirmChecker lease 过期，重新竞选")
				return
			}
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

	var btcHeight int64
	if info, err := c.btcRPC.Get().GetBlockChainInfo(); err == nil {
		btcHeight = int64(info.Blocks)
	}

	var ethHeight int64
	if header, err := c.ethRPC.Get().HeaderByNumber(ctx, nil); err == nil {
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
			metrics.DepositTotal.WithLabelValues(d.Chain, "confirmed").Inc()
			if err := c.queries.UpdateDepositConfirmed(ctx, d.ID); err != nil {
				log.Printf("[ERROR] 更新confirmed失败 id=%d: %v", d.ID, err)
				continue
			}
			// 确认充值后检查是否需要升级用户等级
			go c.checkAndUpgradeLevel(ctx, d.UserID)
			log.Printf("✅ 充值已确认 id=%d chain=%s txid=%s (块高差=%d)",
				d.ID, d.Chain, d.TxID, currentHeight-d.Height)
		}
	}
}

func (c *ConfirmChecker) checkAndUpgradeLevel(ctx context.Context, userID string) {
	// 查询该用户所有链的累计确认充值总额（BTC 换算成 BTC 单位）
	btcTotal, err := c.queries.GetTotalDepositByUserIDAndChain(ctx, db.GetTotalDepositByUserIDAndChainParams{
		UserID: userID,
		Chain:  "btc",
	})
	if err != nil {
		return
	}

	var btcFloat float64
	if v, ok := btcTotal.(string); ok {
		fmt.Sscanf(v, "%f", &btcFloat)
	}

	// 根据累计充值判断等级
	var newLevel int32
	switch {
	case btcFloat >= 50:
		newLevel = 3
	case btcFloat >= 10:
		newLevel = 2
	case btcFloat >= 1:
		newLevel = 1
	default:
		newLevel = 0
	}

	// 查用户 ID（userID 是 string，需要找到对应的 int64）
	// 注意：deposits.user_id 是 TEXT，users.id 是 BIGINT，需要额外查询
	// 暂时跳过，等 user_id 统一后处理
	log.Printf("📊 用户 %s 累计BTC充值 %.8f，建议等级 %d", userID, btcFloat, newLevel)
}
