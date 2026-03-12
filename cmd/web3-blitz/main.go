package main

	import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	
	"internal/pkg/code"
	)
	
	func main() {
	port := getenv("PORT", "8080")
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/", authHandler)
	
	addr := ":" + port
	fmt.Printf("service %q listening on %s\n", "web3-blitz", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
	}
	
	func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
	}

	func authHandler(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer demo-token" {
			errResp(w, code.ErrUnauthorized)
			return
		}
		w.Write([]byte("authorized home"))
	}
	
	func errResp(w http.ResponseWriter, code code.ErrorCode) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]interface{}{
			"code": int(code),
			"msg": code.ErrorCodeString(),
		}
		json.NewEncoder(w).Encode(resp)
	}

	func getenv(key, fallback string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return fallback
	}
