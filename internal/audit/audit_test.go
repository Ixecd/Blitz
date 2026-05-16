package audit

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditLogger(t *testing.T) {
	f, err := os.CreateTemp("", "audit-test-*.log")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	defer f.Close()

	l := &Logger{file: f}

	l.Log(Event{
		Action: "withdraw.completed",
		UserID: "42",
		Chain:  "eth",
		Amount: "1.50000000",
		TxID:   "0xabc",
		Result: "completed",
	})

	// 读回验证
	content, err := os.ReadFile(f.Name())
	require.NoError(t, err)

	assert.Contains(t, string(content), `"action":"withdraw.completed"`)
	assert.Contains(t, string(content), `"user_id":"42"`)
	assert.Contains(t, string(content), `"tx_id":"0xabc"`)
	assert.Contains(t, string(content), `"result":"completed"`)
}

func TestAuditLogger_Close(t *testing.T) {
	f, err := os.CreateTemp("", "audit-close-*.log")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	l := &Logger{file: f}
	l.file.Close()
}

func TestConvenienceFuncs_NilLogger(t *testing.T) {
	// 未初始化 defaultLogger 时不应该 panic
	assert.NotPanics(t, func() {
		WithdrawSubmitted("1", "btc", "0.1", "addr")
		WithdrawCompleted("1", "btc", "0.1", "txid")
		WithdrawFailed("1", "btc", "0.1", "oops")
	})
}

func TestInit(t *testing.T) {
	f, err := os.CreateTemp("", "audit-init-*.log")
	require.NoError(t, err)
	path := f.Name()
	f.Close()
	defer os.Remove(path)

	err = Init(path)
	require.NoError(t, err)
	defer Close()

	require.NotNil(t, defaultLogger)
	WithdrawSubmitted("99", "eth", "2.0", "0xdead")

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "withdraw.submitted")
	assert.Contains(t, string(content), "99")
}
