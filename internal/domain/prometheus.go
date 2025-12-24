package domain

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// 定义一个全局计数器
var (
	IdempotentInterceptTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "app_idempotent_intercept_total", // 指标名字
		Help: "Redis 幂等拦截的总次数",
	})
)
