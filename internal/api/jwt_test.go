package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ixecd/blitz/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestJWTMiddleware_NoHeader(t *testing.T) {
	handler := auth.JWTMiddleware("test-secret", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("不应该执行到这里")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/withdraw", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	handler := auth.JWTMiddleware("test-secret", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("不应该执行到这里")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/withdraw", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	secret := "test-jwt-secret-that-is-long-enough"
	token, err := auth.GenerateToken(42, "testuser", secret)
	assert.NoError(t, err)

	called := false
	handler := auth.JWTMiddleware(secret, func(w http.ResponseWriter, r *http.Request) {
		called = true
		claims := auth.GetClaims(r)
		assert.NotNil(t, claims)
		assert.Equal(t, int64(42), claims.UserID)
		assert.Equal(t, "testuser", claims.Username)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/withdraw", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTMiddleware_WrongSecret(t *testing.T) {
	token, err := auth.GenerateToken(42, "testuser", "right-secret-that-is-long-enough-32chars")
	assert.NoError(t, err)

	handler := auth.JWTMiddleware("wrong-secret-that-is-also-long-enough", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("不应该执行到这里")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/withdraw", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetClaims_Nil(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, auth.GetClaims(req))
}
