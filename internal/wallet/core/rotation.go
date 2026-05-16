package core

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/tyler-smith/go-bip32"
)

// RotationPlan 种子轮换预案：静默派生新路径分流资产，主动规避长期持钥风险
type RotationPlan struct {
	mu            sync.RWMutex
	masterKey     *bip32.Key
	generation    int           // 当前代数（换一次种子 +1）
	activeBranch  string        // 当前活跃分支路径
	branches      []BranchInfo  // 所有分支（含历史）
	rotatedAt     time.Time     // 最近一次轮换时间
}

type BranchInfo struct {
	Generation int
	Path       string // m/44'/60'/0'/<gen>
	CreatedAt  time.Time
	Deprecated bool // 旧分支仍可收款，但新提币走新分支
}

// NewRotationPlan 从 HD 种子创建轮换计划
func NewRotationPlan(masterKey *bip32.Key) *RotationPlan {
	return &RotationPlan{
		masterKey:   masterKey,
		generation:  0,
		activeBranch: "m/44'/60'/0'/0",
		branches: []BranchInfo{
			{Generation: 0, Path: "m/44'/60'/0'/0", CreatedAt: time.Now()},
		},
		rotatedAt: time.Now(),
	}
}

// ActivePath 返回当前活跃分支的 HD 派生路径
func (rp *RotationPlan) ActivePath() string {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.activeBranch
}

// DeriveNext 静默派生出下一个代的分支，不立即切换
// 新分支地址生成后，可逐步将资产从旧分支转入
func (rp *RotationPlan) DeriveNext() (string, error) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	nextGen := rp.generation + 1
	nextPath := fmt.Sprintf("m/44'/60'/0'/%d", nextGen)

	// 验证路径可派生
	_, err := derivePathFromMaster(rp.masterKey, nextPath)
	if err != nil {
		return "", fmt.Errorf("新分支 %s 派生失败: %w", nextPath, err)
	}

	rp.branches = append(rp.branches, BranchInfo{
		Generation: nextGen,
		Path:       nextPath,
		CreatedAt:  time.Now(),
	})

	return nextPath, nil
}

// ActivateBranch 激活指定代的分支（需确保该分支已通过 DeriveNext 创建）
func (rp *RotationPlan) ActivateBranch(generation int) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	for i := range rp.branches {
		if rp.branches[i].Generation == generation {
			rp.activeBranch = rp.branches[i].Path
			rp.generation = generation
			rp.rotatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("分支 generation=%d 不存在，请先调用 DeriveNext", generation)
}

// DeprecateBranch 标记旧分支为已废弃（不再主动使用，但保留收款能力）
func (rp *RotationPlan) DeprecateBranch(generation int) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if generation == rp.generation {
		return fmt.Errorf("不能废弃当前活跃分支")
	}
	for i := range rp.branches {
		if rp.branches[i].Generation == generation {
			rp.branches[i].Deprecated = true
			return nil
		}
	}
	return fmt.Errorf("分支 generation=%d 不存在", generation)
}

// Status 返回当前轮换状态
func (rp *RotationPlan) Status() map[string]interface{} {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	branches := make([]map[string]interface{}, 0, len(rp.branches))
	for _, b := range rp.branches {
		branches = append(branches, map[string]interface{}{
			"generation": b.Generation,
			"path":       b.Path,
			"created_at": b.CreatedAt.Format(time.RFC3339),
			"deprecated": b.Deprecated,
			"active":     b.Path == rp.activeBranch,
		})
	}

	return map[string]interface{}{
		"generation":    rp.generation,
		"active_branch": rp.activeBranch,
		"rotated_at":    rp.rotatedAt.Format(time.RFC3339),
		"branches":      branches,
	}
}

// GenerateRandomSeed 生成 256-bit 随机种子（用于换种子）
func GenerateRandomSeed() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// VerifySeedHash 用 SHA-256 校验种子完整性，与链上公布的哈希对比
func VerifySeedHash(seedHex, publishedHash string) bool {
	h := sha256.Sum256([]byte(seedHex))
	return hex.EncodeToString(h[:]) == publishedHash
}

// PublishSeedHash 生成种子的公开哈希（可上链/公告用于校验）
func PublishSeedHash(seedHex string) string {
	h := sha256.Sum256([]byte(seedHex))
	return hex.EncodeToString(h[:])
}
