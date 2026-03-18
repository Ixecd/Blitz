#!/bin/bash
# 测试余额校验：充值 - 提币 = 可用余额
# 前置：服务已启动，blitz_wallet 有足够余额

BASE="http://localhost:2113"
WALLET_ADDR="bcrt1q3mcvdwr2dmmjh5k0qq576rh6autw23esj5ypnp"  # test001 的 BTC 充值地址
TO_ADDR="bcrt1q573l4xp2tl68q7d9vfcunc854zupsdr252mfzz"       # 提币目标地址
USER="test-$(date +%s)"

echo "=== Step 1: 生成充值地址 ==="
RESP=$(curl -s -X POST $BASE/api/v1/address \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"$USER\",\"chain\":\"btc\"}")
echo $RESP
ADDR=$(echo $RESP | grep -o '"address":"[^"]*"' | cut -d'"' -f4)
echo "充值地址: $ADDR"

echo ""
echo "=== Step 2: 充值 0.5 BTC ==="
TXID=$(bitcoin-cli -regtest -rpcwallet=blitz_wallet sendtoaddress "$ADDR" 0.5)
echo "txid: $TXID"

echo ""
echo "=== Step 3: 挖 1 块（写入DB，confirmed=false）==="
bitcoin-cli -regtest -rpcwallet=blitz_wallet -generate 1 > /dev/null
sleep 1

echo ""
echo "=== Step 4: 确认前提币 0.3，应该被拦截 ==="
curl -s -X POST $BASE/api/v1/withdraw \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"$USER\",\"to_address\":\"$TO_ADDR\",\"amount\":0.3,\"chain\":\"btc\"}"
echo ""

echo ""
echo "=== Step 5: 再挖 5 块，等待 ConfirmChecker 确认（最多30s）==="
bitcoin-cli -regtest -rpcwallet=blitz_wallet -generate 6 > /dev/null
echo "等待 ConfirmChecker 轮询..."
sleep 32

echo ""
echo "=== Step 6: 确认后提币 0.3，应该成功 ==="
curl -s -X POST $BASE/api/v1/withdraw \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"$USER\",\"to_address\":\"$TO_ADDR\",\"amount\":0.3,\"chain\":\"btc\"}"
echo ""

echo ""
echo "=== Step 7: 再提币 0.3，余额只剩 0.2，应该被拦截 ==="
curl -s -X POST $BASE/api/v1/withdraw \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":\"$USER\",\"to_address\":\"$TO_ADDR\",\"amount\":0.3,\"chain\":\"btc\"}"
echo ""

echo ""
echo "=== Step 8: 查询提币历史 ==="
curl -s "$BASE/api/v1/withdrawals?user_id=$USER"
echo ""

echo ""
echo "=== Step 9: 查询累计充值 ==="
curl -s "$BASE/api/v1/balance/total?user_id=$USER&chain=btc"
echo ""