package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	log.Println("🚀 Wallet Core 服务启动中...")

	// HTTP 服务
	http.HandleFunc("/api/v1/address", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "只支持 POST", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			UserID string     `json:"user_id"`
			Chain  types.Chain `json:"chain"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		var resp types.AddressResponse
		var err error

		if req.Chain == types.ChainBTC {
			resp, err = btcWallet.GenerateDepositAddress(r.Context(), req.UserID, req.Chain)
		} else {
			resp, err = ethWallet.GenerateDepositAddress(r.Context(), req.UserID, req.Chain)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		walletAddressesTotal.Inc()
		json.NewEncoder(w).Encode(resp)
	})

	http.Handle("/metrics", promhttp.Handler())

	log.Println("📡 API 服务已启动: http://localhost:2113")
	log.Println("   测试地址生成: curl -X POST http://localhost:2113/api/v1/address -d '{\"user_id\":\"user123\",\"chain\":\"btc\"}'")

	// 优雅退出
	_, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sig; cancel() }()

	http.ListenAndServe(":2113", nil)
}