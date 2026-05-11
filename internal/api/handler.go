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
	btcWallet *btc.BTCWallet
	ethWallet *eth.ETHWallet
	queries   *db.Queries
	db        *sql.DB
	mailer    *email.Mailer
	locker    *lock.DistributedLock
	jwtSecret string
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
	}
}
