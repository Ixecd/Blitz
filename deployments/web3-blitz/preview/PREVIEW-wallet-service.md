# Preview 路由指引 — wallet-service (blue slot)

生成时间：2026-04-04 07:14:14

## 当前环境

未检测到 Istio 或 Nginx Ingress，无法自动生成 Header 路由模板。

## 手动配置选项

### 方案 A：安装 Nginx Ingress Controller

	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.0/deploy/static/provider/cloud/deploy.yaml
	kp deploy --preview   # 重新运行，将自动生成 Ingress canary 模板

### 方案 B：安装 Istio

	curl -L https://istio.io/downloadIstio | sh -
	istioctl install --set profile=minimal
	kp deploy --preview   # 重新运行，将自动生成 VirtualService 模板

### 方案 C：直接 port-forward 验证（本地开发用）

	kubectl port-forward -n <namespace> \
	  deployment/<web3-blitz>-wallet-service-blue <local-port>:2113
	curl http://localhost:<local-port>/healthz

## 确认后切换流量

	kp promote --service wallet-service
