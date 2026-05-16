package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRotationPlan_initialState(t *testing.T) {
	seedHex := "001bf04eb9b83614176551a93c1d405733869d0c1eb9ec9429049c1da5cec5db"
	wallet, err := NewHDWallet(seedHex)
	require.NoError(t, err)

	rp := NewRotationPlan(wallet.MasterKey)
	assert.Equal(t, 0, rp.generation)
	assert.Equal(t, "m/44'/60'/0'/0", rp.ActivePath())
	assert.Len(t, rp.branches, 1)
}

func TestRotationPlan_deriveAndActivate(t *testing.T) {
	seedHex := "001bf04eb9b83614176551a93c1d405733869d0c1eb9ec9429049c1da5cec5db"
	wallet, err := NewHDWallet(seedHex)
	require.NoError(t, err)

	rp := NewRotationPlan(wallet.MasterKey)

	// 派生下一代
	nextPath, err := rp.DeriveNext()
	require.NoError(t, err)
	assert.Equal(t, "m/44'/60'/0'/1", nextPath)
	assert.Len(t, rp.branches, 2)
	assert.Equal(t, "m/44'/60'/0'/0", rp.ActivePath()) // 尚未切换

	// 激活新分支
	err = rp.ActivateBranch(1)
	require.NoError(t, err)
	assert.Equal(t, "m/44'/60'/0'/1", rp.ActivePath())
	assert.Equal(t, 1, rp.generation)

	// 废弃旧分支
	err = rp.DeprecateBranch(0)
	require.NoError(t, err)
}

func TestRotationPlan_cannotDeprecateActive(t *testing.T) {
	seedHex := "001bf04eb9b83614176551a93c1d405733869d0c1eb9ec9429049c1da5cec5db"
	wallet, err := NewHDWallet(seedHex)
	require.NoError(t, err)

	rp := NewRotationPlan(wallet.MasterKey)
	err = rp.DeprecateBranch(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不能废弃当前活跃分支")
}

func TestRotationPlan_status(t *testing.T) {
	seedHex := "001bf04eb9b83614176551a93c1d405733869d0c1eb9ec9429049c1da5cec5db"
	wallet, err := NewHDWallet(seedHex)
	require.NoError(t, err)

	rp := NewRotationPlan(wallet.MasterKey)
	_, _ = rp.DeriveNext()
	_ = rp.ActivateBranch(1)
	_ = rp.DeprecateBranch(0)

	status := rp.Status()
	assert.Equal(t, 1, status["generation"])
	assert.Equal(t, "m/44'/60'/0'/1", status["active_branch"])
	branches := status["branches"].([]map[string]interface{})
	assert.Len(t, branches, 2)
}

func TestGenerateRandomSeed(t *testing.T) {
	s1, err := GenerateRandomSeed()
	require.NoError(t, err)
	assert.Len(t, s1, 64) // 32 bytes = 64 hex chars

	s2, _ := GenerateRandomSeed()
	assert.NotEqual(t, s1, s2)
}

func TestSeedHash_verify(t *testing.T) {
	seed := "001bf04eb9b83614176551a93c1d405733869d0c1eb9ec9429049c1da5cec5db"
	pubHash := PublishSeedHash(seed)
	assert.True(t, VerifySeedHash(seed, pubHash))
	assert.False(t, VerifySeedHash(seed, "deadbeef"))
	assert.Len(t, pubHash, 64)
}
