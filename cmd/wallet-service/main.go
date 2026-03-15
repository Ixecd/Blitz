package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ixecd/web3-blitz/internal/wallet/btc"
	"github.com/Ixecd/web3-blitz/internal/wallet/core"
	"github.com/Ixecd/web3-blitz/internal/wallet/eth"
	"github.com/Ixecd/web3-blitz/internal/wallet/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	walletAddressesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "wallet_addresses_generated_total",
		Help: "Total addresses generated",
	})
)

func main() {
	prometheus.MustRegister(walletAddressesTotal)

	hdWallet, _ := core.NewHDWallet([]byte("test-seed-for-dev-only-1234567890"))
	btcWallet := btc.NewBTCWallet(hdWallet)
	ethWallet := eth.NewETHWallet(hdWallet)

	log.Println("🚀 Wallet Core 服务已启动")

	// 测试地址
	btcResp, _ := btcWallet.GenerateDepositAddress(context.Background(), "test001", types.ChainBTC)
	ethResp, _ := ethWallet.GenerateDepositAddress(context.Background(), "test001", types.ChainETH)
	log.Printf("✅ 测试 BTC 地址: %s", btcResp.Address)
	log.Printf("✅ 测试 ETH 地址: %s", ethResp.Address)

	// HTTP Server
	srv := &http.Server{Addr: ":2113"}

	// API 路由
	http.HandleFunc("/api/v1/address", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "只支持 POST", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			UserID string      `json:"user_id"`
			Chain  types.Chain `json:"chain"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "参数错误", http.StatusBadRequest)
			return
		}

		var resp types.AddressResponse
		var err error

		if req.Chain == types.ChainBTC {
			resp, err = btcWallet.GenerateDepositAddress(r.Context(), req.UserID, req.Chain)
		} else if req.Chain == types.ChainETH {
			resp, err = ethWallet.GenerateDepositAddress(r.Context(), req.UserID, req.Chain)
		} else {
			http.Error(w, "不支持的链", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		walletAddressesTotal.Inc()
		json.NewEncoder(w).Encode(resp)
	})

	http.Handle("/metrics", promhttp.Handler())

	// === 优雅退出（彻底解决 ctx unused）===
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// 启动 HTTP 服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// 真正消费 ctx（关键！）
	go func() {
		<-sig
		log.Println("⛔ 收到关闭信号，正在优雅关闭...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("服务器关闭错误: %v", err)
		}
		cancel()
	}()

	log.Println("📡 API 服务已启动: http://localhost:2113")
	log.Println(`测试 BTC: curl -X POST http://localhost:2113/api/v1/address -H "Content-Type: application/json" -d '{"user_id":"test001","chain":"btc"}'`)
	log.Println(`测试 ETH: curl -X POST http://localhost:2113/api/v1/address -H "Content-Type: application/json" -d '{"user_id":"test001","chain":"eth"}'`)

	// 等待退出
	<-ctx.Done()
}