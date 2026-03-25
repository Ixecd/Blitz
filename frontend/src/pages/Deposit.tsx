import { useState, useEffect } from 'react'
import { useAuth } from '@/contexts/AuthContext'
import { api, ApiError } from '@/api/client'
import { QRCodeSVG } from 'qrcode.react'
import { Copy, Check, ArrowDownToLine, RefreshCw } from 'lucide-react'

type Chain = 'btc' | 'eth'

interface AddressResp {
  address: string
  chain:   string
  user_id: string
}

interface DepositRecord {
  id:         number
  tx_id:      string
  address:    string
  amount:     string
  height:     number
  confirmed:  number
  chain:      string
  created_at: string | null
}

export default function Deposit() {
  const { userID } = useAuth()

  // 地址生成
  const [chain,   setChain]   = useState<Chain>('eth')
  const [address, setAddress] = useState('')
  const [error,   setError]   = useState('')
  const [loading, setLoading] = useState(false)
  const [copied,  setCopied]  = useState(false)

  // 充值历史
  const [deposits,       setDeposits]       = useState<DepositRecord[]>([])
  const [loadingHistory, setLoadingHistory] = useState(false)

  const loadHistory = async () => {
    setLoadingHistory(true)
    try {
      const res = await api.get<DepositRecord[]>('/api/v1/deposits')
      setDeposits(res ?? [])
    } catch {
      // 静默失败，不影响主流程
    } finally {
      setLoadingHistory(false)
    }
  }

  useEffect(() => { loadHistory() }, [])

  const generate = async () => {
    if (!userID) return
    setError('')
    setLoading(true)
    try {
      const res = await api.post<AddressResp>('/api/v1/address', {
        user_id: String(userID),
        chain,
      })
      setAddress(res.address)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : '获取地址失败')
    } finally {
      setLoading(false)
    }
  }

  const copy = async () => {
    if (!address) return
    await navigator.clipboard.writeText(address)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const confirmedLabel = (confirmed: number, chain: string) => {
    const required = chain === 'btc' ? 6 : 12
    if (confirmed >= required) return <span className="text-success">已确认</span>
    return <span className="text-warning">{confirmed}/{required} 确认中</span>
  }

  return (
    <div className="space-y-8">
      <div>
        <div className="flex items-center gap-2 mb-1">
          <ArrowDownToLine size={16} className="text-success" />
          <h2 className="font-display font-bold text-2xl">充值</h2>
        </div>
        <p className="text-text-muted text-sm">生成专属充值地址，向该地址转账即可完成充值</p>
      </div>

      {/* 地址生成卡片 */}
      <div className="card p-6 max-w-md space-y-6">
        <div className="space-y-2">
          <label className="text-xs font-mono text-text-muted uppercase tracking-widest">
            选择链
          </label>
          <div className="flex gap-2">
            {(['eth', 'btc'] as Chain[]).map(c => (
              <button
                key={c}
                onClick={() => { setChain(c); setAddress('') }}
                className={
                  'chain-pill ' +
                  (chain === c ? 'chain-pill-active' : 'chain-pill-inactive')
                }
              >
                {c.toUpperCase()}
              </button>
            ))}
          </div>
        </div>

        <button onClick={generate} disabled={loading} className="btn-primary w-full">
          {loading ? '生成中…' : '获取充值地址'}
        </button>

        {error && (
          <p className="text-danger text-xs font-mono bg-danger/5 border border-danger/20 rounded-lg px-3 py-2">
            {error}
          </p>
        )}

        {address && (
          <div className="space-y-4 animate-fade-up">
            <div className="flex justify-center">
              <div className="p-3 bg-white rounded-xl">
                <QRCodeSVG
                  value={address}
                  size={160}
                  bgColor="#ffffff"
                  fgColor="#070711"
                  level="M"
                />
              </div>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-mono text-text-muted uppercase tracking-widest">
                {chain.toUpperCase()} 充值地址
              </label>
              <div className="flex items-center gap-2">
                <div className="flex-1 bg-raised border border-border rounded-lg px-3 py-2.5 font-mono text-xs text-text-primary break-all leading-relaxed">
                  {address}
                </div>
                <button
                  onClick={copy}
                  className={
                    'shrink-0 w-9 h-9 rounded-lg border flex items-center justify-center transition-all duration-150 ' +
                    (copied
                      ? 'border-success/40 bg-success/10 text-success'
                      : 'border-border bg-raised text-text-muted hover:border-gold hover:text-gold')
                  }
                >
                  {copied ? <Check size={14} /> : <Copy size={14} />}
                </button>
              </div>
            </div>

            <p className="text-xs text-text-faint font-mono text-center">
              {chain === 'btc' ? '等待 6 个区块确认' : '等待 12 个区块确认'}
            </p>
          </div>
        )}
      </div>

      {/* 充值历史 */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h3 className="font-display font-semibold text-lg">充值记录</h3>
          <button
            onClick={loadHistory}
            disabled={loadingHistory}
            className="flex items-center gap-1.5 text-xs text-text-muted hover:text-text-primary transition-colors"
          >
            <RefreshCw size={12} className={loadingHistory ? 'animate-spin' : ''} />
            刷新
          </button>
        </div>

        {deposits.length === 0 ? (
          <div className="card p-8 text-center text-text-faint text-sm font-mono">
            {loadingHistory ? '加载中…' : '暂无充值记录'}
          </div>
        ) : (
          <div className="card divide-y divide-border">
            {deposits.map(d => (
              <div key={d.id} className="px-4 py-3 flex items-center justify-between gap-4">
                <div className="min-w-0 space-y-0.5">
                  <div className="flex items-center gap-2">
                    <span className="chain-pill chain-pill-inactive text-xs py-0.5">
                      {d.chain.toUpperCase()}
                    </span>
                    <span className="font-mono text-sm font-semibold text-text-primary">
                      +{d.amount}
                    </span>
                    <span className="text-xs font-mono">
                      {confirmedLabel(d.confirmed, d.chain)}
                    </span>
                  </div>
                  <p className="font-mono text-xs text-text-faint truncate">
                    {d.tx_id}
                  </p>
                </div>
                <div className="shrink-0 text-xs text-text-faint font-mono text-right">
                  {d.created_at
                    ? new Date(d.created_at).toLocaleDateString('zh-CN')
                    : '—'}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
