#!/bin/bash
# 创建/更新 web3-blitz 的 K8s Secret
# 幂等：已存在则更新，不存在则创建
# 用法：./scripts/create-secret.sh
#
# 生产环境：修改下方变量为真实值
# 开发/demo：直接运行，自动使用安全随机值

set -e

NAMESPACE=${KUBE_NAMESPACE:-web3-blitz}

# 从 .env 读取（如果存在）
if [ -f .env ]; then
  export $(grep -v '^#' .env | grep -v '^$' | xargs)
fi

kubectl create secret generic web3-blitz-secret \
  -n "$NAMESPACE" \
  --from-literal=DATABASE_URL="${DATABASE_URL:-postgres://web3-blitz:web3-blitz@web3-blitz-postgres:5432/web3-blitz?sslmode=disable}" \
  --from-literal=JWT_SECRET="${JWT_SECRET:-$(openssl rand -hex 32)}" \
  --from-literal=APP_SECRET="${APP_SECRET:-$(openssl rand -hex 16)}" \
  --save-config \
  --dry-run=client -o yaml | kubectl apply -f -

echo "✅ web3-blitz-secret 已创建/更新（namespace: $NAMESPACE）"
echo "⚠️  生产环境请设置真实的 DATABASE_URL / JWT_SECRET / APP_SECRET"
