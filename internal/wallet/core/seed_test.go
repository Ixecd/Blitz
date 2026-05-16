package core

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt_roundtrip(t *testing.T) {
	seedHex := "001bf04eb9b83614176551a93c1d405733869d0c1eb9ec9429049c1da5cec5db"
	passphrase := "correct-horse-battery-staple"

	encrypted, err := EncryptSeed(seedHex, passphrase)
	require.NoError(t, err)
	require.Greater(t, len(encrypted), 44) // salt(32)+nonce(12)+至少1字节密文

	f, err := os.CreateTemp("", "seed-enc-*.enc")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	err = os.WriteFile(f.Name(), encrypted, 0600)
	require.NoError(t, err)

	decrypted, err := DecryptSeedFile(f.Name(), passphrase)
	require.NoError(t, err)
	assert.Equal(t, seedHex, decrypted)
}

func TestDecryptSeedFile_wrongPassphrase(t *testing.T) {
	seedHex := "001bf04eb9b83614176551a93c1d405733869d0c1eb9ec9429049c1da5cec5db"

	encrypted, err := EncryptSeed(seedHex, "correct-password")
	require.NoError(t, err)

	f, err := os.CreateTemp("", "seed-wrong-*.enc")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	os.WriteFile(f.Name(), encrypted, 0600)

	_, err = DecryptSeedFile(f.Name(), "wrong-password")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "解密种子失败")
}

func TestDecryptSeedFile_missingFile(t *testing.T) {
	_, err := DecryptSeedFile("/nonexistent/path/hd-seed.enc", "test")
	assert.Error(t, err)
}

func TestNewHDWalletFromEncrypted(t *testing.T) {
	seedHex := "001bf04eb9b83614176551a93c1d405733869d0c1eb9ec9429049c1da5cec5db"
	passphrase := "wallet-unlock-test"

	encrypted, err := EncryptSeed(seedHex, passphrase)
	require.NoError(t, err)

	f, err := os.CreateTemp("", "hd-seed-*.enc")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	os.WriteFile(f.Name(), encrypted, 0600)

	wallet, err := NewHDWalletFromEncrypted(f.Name(), passphrase)
	require.NoError(t, err)
	require.NotNil(t, wallet)
	require.NotNil(t, wallet.MasterKey)

	// 验证能正常派生地址
	key, err := wallet.DerivePath("m/44'/60'/0'/0/0")
	require.NoError(t, err)
	require.NotNil(t, key)
}

func TestDerivePassphraseHash(t *testing.T) {
	h1 := DerivePassphraseHash("my-password")
	h2 := DerivePassphraseHash("my-password")
	h3 := DerivePassphraseHash("different")

	assert.Equal(t, h1, h2)
	assert.NotEqual(t, h1, h3)
	assert.Len(t, h1, 64) // SHA-256 = 64 hex chars
}
