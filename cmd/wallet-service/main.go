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

	hdWallet, err := core.NewHDWallet()
	if err != nil {
		log.Fatal("HDWallet 初始化失败:", err)
	}

	btcWallet := btc.NewBTCWallet(hdWallet)
	ethWallet := eth.NewETHWallet(hdWallet)

	log.Println("🚀 Wallet Core 服务已启动")

	// === HTTP API（带参数验证 + 日志 + 错误处理）===
	http.HandleFunc("/api/v1/address", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("[ERROR] 不支持的请求方法: %s", r.Method)
			http.Error(w, "只支持 POST 请求", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			UserID string      `json:"user_id"`
			Chain  types.Chain `json:"chain"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[ERROR] 参数解析失败: %v", err)
			http.Error(w, "请求参数格式错误", http.StatusBadRequest)
			return
		}

		// 参数验证
		if req.UserID == "" {
			log.Println("[ERROR] user_id 不能为空")
			http.Error(w, "user_id 不能为空", http.StatusBadRequest)
			return
		}
		if req.Chain != types.ChainBTC && req.Chain != types.ChainETH {
			log.Printf("[ERROR] 不支持的链类型: %s", req.Chain)
			http.Error(w, "不支持的链类型，仅支持 btc/eth", http.StatusBadRequest)
			return
		}

		log.Printf("[INFO] 收到地址生成请求 | user_id=%s | chain=%s", req.UserID, req.Chain)

		var resp types.AddressResponse
		var genErr error

		if req.Chain == types.ChainBTC {
			resp, genErr = btcWallet.GenerateDepositAddress(r.Context(), req.UserID, req.Chain)
		} else {
			resp, genErr = ethWallet.GenerateDepositAddress(r.Context(), req.UserID, req.Chain)
		}

		if genErr != nil {
			log.Printf("[ERROR] 生成地址失败: %v", genErr)
			http.Error(w, "生成地址失败: "+genErr.Error(), http.StatusInternalServerError)
			return
		}

		walletAddressesTotal.Inc()
		log.Printf("[SUCCESS] 地址生成成功 | address=%s | path=%s", resp.Address, resp.Path)

		json.NewEncoder(w).Encode(resp)
	})

	http.Handle("/metrics", promhttp.Handler())

	// === 优雅退出（ctx 被真正消费）===
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{Addr: ":2113"}

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

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	// 等待退出
	<-ctx.Done()
}