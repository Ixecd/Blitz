package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Ixecd/blitz/internal/audit"
	"github.com/Ixecd/blitz/internal/auth"
	"github.com/Ixecd/blitz/internal/code"
	"github.com/Ixecd/blitz/internal/db"
	"github.com/Ixecd/blitz/internal/metrics"
	"github.com/Ixecd/blitz/internal/wallet/types"
	"github.com/ethereum/go-ethereum/common"
)

func (h *Handler) GenerateAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Fail(w, code.ErrInvalidArg)
		return
	}

	claims := auth.GetClaims(r)
	if claims == nil {
		Fail(w, code.ErrUnauthorized)
		return
	}

	var req struct {
		Chain types.Chain `json:"chain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, code.ErrInvalidArg)
		return
	}

	userID := fmt.Sprintf("%d", claims.UserID)

	var resp types.AddressResponse
	var genErr error

	switch req.Chain {
	case types.ChainBTC:
		resp, genErr = h.btcWallet.GenerateDepositAddress(r.Context(), userID, req.Chain)
	case types.ChainETH:
		resp, genErr = h.ethWallet.GenerateDepositAddress(r.Context(), userID, req.Chain)
	default:
		Fail(w, code.ErrWalletChainNotSupported)
		return
	}

	if genErr != nil {
		FailMsg(w, code.ErrInternal, genErr.Error())
		return
	}

	OK(w, resp)
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		Fail(w, code.ErrInvalidArg)
		return
	}

	address := r.URL.Query().Get("address")
	chainStr := r.URL.Query().Get("chain")

	if address == "" || chainStr == "" {
		FailMsg(w, code.ErrInvalidArg, "缺少 address 或 chain 参数")
		return
	}

	var resp types.BalanceResponse
	var err error

	switch chainStr {
	case "btc":
		resp, err = h.btcWallet.GetBalance(r.Context(), address, types.ChainBTC)
	case "eth":
		resp, err = h.ethWallet.GetBalance(r.Context(), address, types.ChainETH)
	default:
		Fail(w, code.ErrWalletChainNotSupported)
		return
	}

	if err != nil {
		FailInternal(w, err)
		return
	}

	OK(w, resp)
}

func (h *Handler) ListDeposits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		Fail(w, code.ErrInvalidArg)
		return
	}

	claims := auth.GetClaims(r)
	if claims == nil {
		Fail(w, code.ErrUnauthorized)
		return
	}
	userID := fmt.Sprintf("%d", claims.UserID)

	deposits, err := h.queries.ListDepositsByUserID(r.Context(), userID)
	if err != nil {
		FailInternal(w, err)
		return
	}

	OK(w, deposits)
}

func (h *Handler) GetTotalBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		Fail(w, code.ErrInvalidArg)
		return
	}

	claims := auth.GetClaims(r)
	if claims == nil {
		Fail(w, code.ErrUnauthorized)
		return
	}

	chainStr := r.URL.Query().Get("chain")
	if chainStr == "" {
		FailMsg(w, code.ErrInvalidArg, "缺少 chain 参数")
		return
	}

	userID := fmt.Sprintf("%d", claims.UserID)

	total, err := h.queries.GetTotalDepositByUserIDAndChain(r.Context(), db.GetTotalDepositByUserIDAndChainParams{
		UserID: userID,
		Chain:  chainStr,
	})
	if err != nil {
		FailInternal(w, err)
		return
	}

	var totalFloat float64
	if t, ok := total.(string); ok {
		fmt.Sscanf(t, "%f", &totalFloat)
	} else {
		totalFloat, _ = total.(float64)
	}

	OK(w, map[string]interface{}{
		"user_id": userID,
		"chain":   chainStr,
		"total":   totalFloat,
	})
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Fail(w, code.ErrInvalidArg)
		return
	}

	var req struct {
		ToAddress string      `json:"to_address"`
		Amount    float64     `json:"amount"`
		Chain     types.Chain `json:"chain"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, code.ErrInvalidArg)
		return
	}
	if req.ToAddress == "" || req.Amount <= 0 {
		FailMsg(w, code.ErrInvalidArg, "to_address / amount 不能为空或非正数")
		return
	}
	// 目标地址合法性校验
	if req.Chain == types.ChainETH && !common.IsHexAddress(req.ToAddress) {
		FailMsg(w, code.ErrInvalidArg, "ETH 地址格式不合法")
		return
	}

	ctx := r.Context()
	claims := auth.GetClaims(r)
	if claims == nil {
		Fail(w, code.ErrUnauthorized)
		return
	}
	userID := fmt.Sprintf("%d", claims.UserID)

	// 单笔提币上限（BTC）
	const maxSingleBTC = 1.0
	const maxSingleETH = 10.0
	switch req.Chain {
	case types.ChainBTC:
		if req.Amount > maxSingleBTC {
			FailMsg(w, code.ErrWalletInsufficientBalance,
				fmt.Sprintf("单笔提币上限 %.1f BTC，本次 %.8f", maxSingleBTC, req.Amount))
			return
		}
	case types.ChainETH:
		if req.Amount > maxSingleETH {
			FailMsg(w, code.ErrWalletInsufficientBalance,
				fmt.Sprintf("单笔提币上限 %.1f ETH，本次 %.8f", maxSingleETH, req.Amount))
			return
		}
	}

	// 分布式锁，防止重复提币
	lockKey := fmt.Sprintf("withdraw:%s:%s", userID, req.Chain)
	l, err := h.locker.Acquire(ctx, lockKey)
	if err != nil {
		metrics.LockAcquireFailTotal.WithLabelValues(lockKey).Inc()
		Fail(w, code.ErrWalletDuplicateWithdraw)
		return
	}
	defer l.Release(context.Background())

	// 类型转换
	toFloat := func(v interface{}) float64 {
		switch val := v.(type) {
		case float64:
			return val
		case int64:
			return float64(val)
		case string:
			var f float64
			fmt.Sscanf(val, "%f", &f)
			return f
		case []byte:
			var f float64
			fmt.Sscanf(string(val), "%f", &f)
			return f
		}
		return 0
	}

	// 余额校验
	rawDeposit, err := h.queries.GetTotalDepositByUserIDAndChain(ctx, db.GetTotalDepositByUserIDAndChainParams{
		UserID: userID,
		Chain:  string(req.Chain),
	})
	if err != nil {
		FailInternal(w, err)
		return
	}

	rawWithdrawal, err := h.queries.GetTotalWithdrawalByUserIDAndChain(ctx, db.GetTotalWithdrawalByUserIDAndChainParams{
		UserID: userID,
		Chain:  string(req.Chain),
	})
	if err != nil {
		FailInternal(w, err)
		return
	}

	available := toFloat(rawDeposit) - toFloat(rawWithdrawal)
	if available < req.Amount {
		FailMsg(w, code.ErrWalletInsufficientBalance,
			fmt.Sprintf("余额不足: 可用 %.8f，请求 %.8f", available, req.Amount))
		return
	}

	// 限额校验
	userLevel, err := h.queries.GetUserLevel(ctx, claims.UserID)
	if err != nil {
		FailInternal(w, err)
		return
	}

	limit, err := h.queries.GetWithdrawalLimit(ctx, int32(userLevel))
	if err != nil {
		FailInternal(w, err)
		return
	}

	rawUsed, err := h.queries.GetLast24hWithdrawalByUserAndChain(ctx, db.GetLast24hWithdrawalByUserAndChainParams{
		UserID: userID,
		Chain:  string(req.Chain),
	})
	if err != nil {
		FailInternal(w, err)
		return
	}

	usedToday := toFloat(rawUsed)
	var dailyLimit float64
	switch req.Chain {
	case types.ChainBTC:
		fmt.Sscanf(limit.BtcDaily, "%f", &dailyLimit)
	case types.ChainETH:
		fmt.Sscanf(limit.EthDaily, "%f", &dailyLimit)
	}

	if usedToday+req.Amount > dailyLimit {
		FailMsg(w, code.ErrWalletDailyLimitExceeded,
			fmt.Sprintf("超出每日提币限额: 已用 %.8f，本次 %.8f，限额 %.8f（%s）",
				usedToday, req.Amount, dailyLimit, limit.LevelName))
		return
	}

	// 写入 pending 记录
	record, err := h.queries.CreateWithdrawal(ctx, db.CreateWithdrawalParams{
		UserID:  userID,
		Address: req.ToAddress,
		Amount:  fmt.Sprintf("%.8f", req.Amount),
		Chain:   string(req.Chain),
	})
	if err != nil {
		FailInternal(w, err)
		return
	}

	audit.WithdrawSubmitted(userID, string(req.Chain),
		fmt.Sprintf("%.8f", req.Amount), req.ToAddress)

	// 广播交易
	var txID string
	var fee float64
	var broadcastErr error

	switch req.Chain {
	case types.ChainBTC:
		res, err := h.btcWallet.Withdraw(ctx, req.ToAddress, req.Amount)
		txID, fee, broadcastErr = res.TxID, res.Fee, err
	case types.ChainETH:
		res, err := h.ethWallet.Withdraw(ctx, req.ToAddress, req.Amount)
		txID, fee, broadcastErr = res.TxID, res.Fee, err
	default:
		Fail(w, code.ErrWalletChainNotSupported)
		return
	}

	// 更新 DB 状态
	status := "completed"
	if broadcastErr != nil {
		status = "failed"
		slog.Error("提币广播失败", "id", record.ID, "err", broadcastErr)
	}

	_ = h.queries.UpdateWithdrawalTx(ctx, db.UpdateWithdrawalTxParams{
		TxID:   sql.NullString{String: txID, Valid: txID != ""},
		Fee:    fmt.Sprintf("%.8f", fee),
		Status: status,
		ID:     record.ID,
	})

	if broadcastErr != nil {
		metrics.WithdrawTotal.WithLabelValues(string(req.Chain), "failed").Inc()
		audit.WithdrawFailed(userID, string(req.Chain),
			fmt.Sprintf("%.8f", req.Amount), broadcastErr.Error())
	} else {
		metrics.WithdrawTotal.WithLabelValues(string(req.Chain), "completed").Inc()
		metrics.WithdrawAmount.WithLabelValues(string(req.Chain)).Add(req.Amount)
		audit.WithdrawCompleted(userID, string(req.Chain),
			fmt.Sprintf("%.8f", req.Amount), txID)
	}

	OK(w, map[string]interface{}{
		"id":         record.ID,
		"tx_id":      txID,
		"user_id":    userID,
		"to_address": req.ToAddress,
		"amount":     req.Amount,
		"fee":        fee,
		"status":     status,
		"chain":      req.Chain,
	})
}

func (h *Handler) ListWithdrawals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		Fail(w, code.ErrInvalidArg)
		return
	}

	claims := auth.GetClaims(r)
	if claims == nil {
		Fail(w, code.ErrUnauthorized)
		return
	}
	userID := fmt.Sprintf("%d", claims.UserID)

	withdrawals, err := h.queries.ListWithdrawalsByUserID(r.Context(), userID)
	if err != nil {
		FailInternal(w, err)
		return
	}

	type WithdrawalResp struct {
		ID        int64   `json:"id"`
		TxID      string  `json:"tx_id"`
		Address   string  `json:"address"`
		UserID    string  `json:"user_id"`
		Amount    float64 `json:"amount"`
		Fee       float64 `json:"fee"`
		Status    string  `json:"status"`
		Chain     string  `json:"chain"`
		CreatedAt string  `json:"created_at"`
	}

	resp := make([]WithdrawalResp, 0, len(withdrawals))
	for _, wl := range withdrawals {
		var amount, fee float64
		fmt.Sscanf(wl.Amount, "%f", &amount)
		fmt.Sscanf(wl.Fee, "%f", &fee)
		resp = append(resp, WithdrawalResp{
			ID:        wl.ID,
			TxID:      wl.TxID.String,
			Address:   wl.Address,
			UserID:    wl.UserID,
			Amount:    amount,
			Fee:       fee,
			Status:    wl.Status,
			Chain:     wl.Chain,
			CreatedAt: wl.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		})
	}

	OK(w, resp)
}
