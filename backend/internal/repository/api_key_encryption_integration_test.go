//go:build integration

package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// TestAPIKey_EncryptedAtRest_DecryptedOnRead 验证启用加密时 api_key 端到端:
//   - Create 后直查 ent:key_encrypted 为 enc:v1: 密文、key_hash 填充、原文不出现。
//   - GetByID 读出:key 解密还原明文(用户始终可见)。
func TestAPIKey_EncryptedAtRest_DecryptedOnRead(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	client := tx.Client()
	repo := newAPIKeyRepositoryWithSQL(client, tx, newCredEncCipher(t)) // enabled cipher

	user := createEntUser(t, ctx, client, "apikey-enc-roundtrip@example.com")

	key := &service.APIKey{
		UserID: user.ID,
		Key:    "sk-encrypted-roundtrip-xyz",
		Name:   "enc-test",
		Status: service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, key))
	require.NotZero(t, key.ID)

	// 直查 ent:敏感原文不应落库,key_encrypted 密文 + key_hash 填充。
	raw, err := client.APIKey.Get(ctx, key.ID)
	require.NoError(t, err)
	require.NotNil(t, raw.KeyEncrypted, "key_encrypted 应写入")
	require.True(t, strings.HasPrefix(*raw.KeyEncrypted, "enc:v1:"), "key_encrypted 应为密文, got %v", *raw.KeyEncrypted)
	require.NotContains(t, *raw.KeyEncrypted, "sk-encrypted-roundtrip-xyz", "密文不应含明文")
	require.NotNil(t, raw.KeyHash)
	require.NotEmpty(t, *raw.KeyHash, "key_hash 应填充(auth lookup)")

	// GetByID 读出口解密还原明文。
	got, err := repo.GetByID(ctx, key.ID)
	require.NoError(t, err)
	require.Equal(t, "sk-encrypted-roundtrip-xyz", got.Key, "读出口应解密还原明文")
}

// TestAPIKey_DegradeWhenCipherNil 验证 nil cipher(degrade)时写明文 key、key_hash 仍填充、读回退明文。
func TestAPIKey_DegradeWhenCipherNil(t *testing.T) {
	ctx := context.Background()
	tx := testEntTx(t)
	client := tx.Client()
	repo := newAPIKeyRepositoryWithSQL(client, tx, nil) // degrade

	user := createEntUser(t, ctx, client, "apikey-degrade@example.com")

	key := &service.APIKey{
		UserID: user.ID,
		Key:    "sk-degrade-plain",
		Name:   "degrade-test",
		Status: service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, key))

	raw, err := client.APIKey.Get(ctx, key.ID)
	require.NoError(t, err)
	require.Nil(t, raw.KeyEncrypted, "nil cipher 不应写 key_encrypted")
	require.NotNil(t, raw.KeyHash)
	require.NotEmpty(t, *raw.KeyHash, "key_hash 始终填充(用于 auth lookup,与加密无关)")

	got, err := repo.GetByID(ctx, key.ID)
	require.NoError(t, err)
	require.Equal(t, "sk-degrade-plain", got.Key, "degrade 时读回退 key 明文列")
}
