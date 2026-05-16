package api

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(10, 10) // 10 tokens/s, burst 10

	// 前 10 次都应该通过
	for i := range 10 {
		assert.True(t, rl.Allow("test"), "第 %d 次应该通过", i+1)
	}

	// 第 11 次被限
	assert.False(t, rl.Allow("test"))

	// 等 200ms 恢复 2 个 token
	time.Sleep(200 * time.Millisecond)
	assert.True(t, rl.Allow("test"))
	assert.True(t, rl.Allow("test"))
	assert.False(t, rl.Allow("test"))
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	rl := NewRateLimiter(1, 1)

	assert.True(t, rl.Allow("ip-a"))
	assert.False(t, rl.Allow("ip-a"))
	assert.True(t, rl.Allow("ip-b")) // 不同 IP 不受影响
}

func TestRateLimiter_Concurrency(t *testing.T) {
	rl := NewRateLimiter(100, 100)
	var wg sync.WaitGroup
	results := make([]bool, 50)

	for i := range 50 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = rl.Allow("concurrent")
		}(i)
	}
	wg.Wait()

	passed := 0
	for _, ok := range results {
		if ok {
			passed++
		}
	}
	assert.Greater(t, passed, 0, "并发下至少有些请求应通过")
}

func TestRateLimitMiddleware_Allowed(t *testing.T) {
	rl := NewRateLimiter(10, 10)
	handler := RateLimitMiddleware(rl, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitMiddleware_Blocked(t *testing.T) {
	rl := NewRateLimiter(0, 0) // burst=0，直接拒绝
	handler := RateLimitMiddleware(rl, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("不应该执行到这里")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, "60", w.Header().Get("Retry-After"))
}

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	assert.Equal(t, "192.168.1.1:12345", clientIP(req))

	req.Header.Set("X-Real-IP", "10.0.0.1")
	assert.Equal(t, "10.0.0.1", clientIP(req))

	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	assert.Equal(t, "1.2.3.4", clientIP(req))
}
