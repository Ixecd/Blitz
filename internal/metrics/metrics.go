package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// 充值相关
	DepositTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "blitz_deposit_total",
		Help: "Total number of deposits detected",
	}, []string{"chain", "status"}) // status: detected / confirmed

	DepositAmount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "blitz_deposit_amount_total",
		Help: "Total deposit amount",
	}, []string{"chain"})

	// 提币相关
	WithdrawTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "blitz_withdraw_total",
		Help: "Total number of withdrawals",
	}, []string{"chain", "status"}) // status: completed / failed

	WithdrawAmount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "blitz_withdraw_amount_total",
		Help: "Total withdrawal amount",
	}, []string{"chain"})

	// 死信队列
	DeadLetterTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "blitz_dead_letter_total",
		Help: "Total number of dead letters",
	}, []string{"type"})

	// reorg 告警
	ReorgTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "blitz_reorg_total",
		Help: "Total number of chain reorganizations detected",
	}, []string{"chain"})

	// etcd 锁等待
	LockAcquireFailTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "blitz_lock_acquire_fail_total",
		Help: "Total number of distributed lock acquire failures",
	}, []string{"key"})
)

func Init() {
	prometheus.MustRegister(
		DepositTotal,
		DepositAmount,
		WithdrawTotal,
		WithdrawAmount,
		DeadLetterTotal,
		ReorgTotal,
		LockAcquireFailTotal,
	)
}
