# web3-blitz 监控告警

## 目录结构

```
monitoring/
├── prometheus/
│   ├── prometheus.yml
│   └── rules/web3-blitz.yml
├── alertmanager/
│   └── alertmanager.yml
└── grafana/
    ├── dashboards/web3-blitz.json
    └── provisioning/
        ├── dashboards/dashboard.yml
        └── datasources/prometheus.yml
```

## 告警规则

| 告警名 | 级别 | 触发条件 |
|--------|------|---------|
| ServiceDown | critical | 服务不可达超过1分钟 |
| HighErrorRate | warning | 5分钟内5xx超过10次 |
| HighLatency | warning | P99延迟超过1秒 |
