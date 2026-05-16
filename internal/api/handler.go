package api

import (
	"database/sql"

	"github.com/Ixecd/blitz/internal/db"
	"github.com/Ixecd/blitz/internal/email"
	"github.com/Ixecd/blitz/internal/lock"
	"github.com/Ixecd/blitz/internal/wallet/btc"
	"github.com/Ixecd/blitz/internal/wallet/eth"
)

type Handler struct {
	btcWallet   *btc.BTCWallet
	ethWallet   *eth.ETHWallet
	queries     *db.Queries
	db          *sql.DB
	mailer      *email.Mailer
	locker      *lock.DistributedLock
	jwtSecret   string
	generalRL   *RateLimiter // 全局限速
	authRL      *RateLimiter // 登录/注册限速（更严）
}

func NewHandler(btcWallet *btc.BTCWallet, ethWallet *eth.ETHWallet, queries *db.Queries, db *sql.DB, locker *lock.DistributedLock, jwtSecret string, mailer *email.Mailer) *Handler {
	return &Handler{
		btcWallet: btcWallet,
		ethWallet: ethWallet,
		queries:   queries,
		db:        db,
		locker:    locker,
		jwtSecret: jwtSecret,
		mailer:    mailer,
		generalRL: NewRateLimiter(2, 100),   // 2 token/s = 120 req/min
		authRL:    NewRateLimiter(0.1, 5),   // 0.1 token/s = 6 req/min
	}
}
