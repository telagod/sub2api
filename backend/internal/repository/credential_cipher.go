package repository

import (
	"log/slog"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// credentialEncPrefixV1 标记一个 credentials 值已被字段级加密。
// 值格式: "enc:v1:" + base64(AES-256-GCM(nonce+ciphertext+tag))。
// 前缀既标识"已加密"又标识算法版本，未来换算法/轮换 key 可升 "enc:v2:"。
const credentialEncPrefixV1 = "enc:v1:"

// CredentialCipher 负责 account.credentials 中敏感子键的字段级加解密。
//
// 设计要点（详见 credential-encryption 蓝图）：
//   - 字段级：仅加密 service.IsSensitiveCredentialKey 命中的 string 值，
//     保留 model_mapping 等嵌套 map / base_url 等非敏感字段为明文，
//     使 BulkUpdate 的 JSONB partial merge 与 GetModelMapping 指针缓存不受影响。
//   - 值前缀自描述：无需 schema 迁移，明文/密文可在同一 map 内共存（向后兼容）。
//   - degrade-safe：仅当加密密钥为持久配置（cfg.Totp.EncryptionKeyConfigured）时写入加密，
//     避免自动生成的易失密钥在重启后写出永久不可解的密文。读路径永不 gate。
//   - nil-safe：所有方法对 nil receiver / nil 入参原样返回，便于测试构造点传 nil。
type CredentialCipher struct {
	enc     service.SecretEncryptor
	enabled bool // = cfg.Totp.EncryptionKeyConfigured；仅 gate 写入加密
}

// NewCredentialCipher 构造字段级凭证加解密器（供 wire 注入）。
func NewCredentialCipher(cfg *config.Config, enc service.SecretEncryptor) *CredentialCipher {
	enabled := cfg != nil && cfg.Totp.EncryptionKeyConfigured
	return &CredentialCipher{enc: enc, enabled: enabled}
}

// pkgCredentialCipher 供包级 accountEntityToService 读路径解密使用。
// 由 NewAccountRepository 在启动构造期设置一次，运行期只读。
// 未设置（nil）时 DecryptMap 的 nil-safe 使其退化为明文直通（测试默认 / 向后兼容）。
var pkgCredentialCipher *CredentialCipher

// setPkgCredentialCipher 注册包级解密器。仅应在启动构造期调用。
func setPkgCredentialCipher(c *CredentialCipher) { pkgCredentialCipher = c }

// EncryptMap 加密 credentials 中敏感子键的 string 值，返回新 map（不修改入参）。
//
// degrade-safe：cipher 为 nil / 未启用 / 无加密器时原样返回（写明文）。
// 幂等 + 防伪造（BLINDSPOT-3）：对已带 enc:v1: 前缀的敏感值，先用 Decrypt 验证——
//   - 能解密 → 是本系统写出的密文，跳过（幂等，不双重加密）；
//   - 不能解密 → 是客户端伪造的前缀，剥离后按明文重新加密，
//     防止 {"api_key":"enc:v1:spoofed"} 这类伪造前缀的明文直接落库。
func (c *CredentialCipher) EncryptMap(in map[string]any) map[string]any {
	if c == nil || !c.enabled || c.enc == nil || in == nil {
		return in
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		s, ok := v.(string)
		if !ok || !service.IsSensitiveCredentialKey(k) {
			out[k] = v
			continue
		}
		if strings.HasPrefix(s, credentialEncPrefixV1) {
			body := strings.TrimPrefix(s, credentialEncPrefixV1)
			if _, err := c.enc.Decrypt(body); err == nil {
				// 本系统密文，幂等跳过。
				out[k] = v
				continue
			}
			// 伪造前缀，剥离当明文重新加密。
			s = body
		}
		ct, err := c.enc.Encrypt(s)
		if err != nil {
			// 加密失败：保留明文，避免数据丢失（degrade）。
			slog.Error("credential encrypt failed, storing plaintext", "key", k, "err", err)
			out[k] = v
			continue
		}
		out[k] = credentialEncPrefixV1 + ct
	}
	return out
}

// DecryptMap 解密带 enc:v1: 前缀的 credentials 值，返回新 map（不修改入参）。
//
// 不受 enabled gate：即使写入加密被禁用，仍须能解密历史密文。
// 向后兼容：无前缀的值视为历史明文，原样返回。
// 容错：单值解密失败保留密文 + warn（不 panic、不使整个账号不可用）。
func (c *CredentialCipher) DecryptMap(in map[string]any) map[string]any {
	if c == nil || c.enc == nil || in == nil {
		return in
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		s, ok := v.(string)
		if !ok || !strings.HasPrefix(s, credentialEncPrefixV1) {
			out[k] = v
			continue
		}
		pt, err := c.enc.Decrypt(strings.TrimPrefix(s, credentialEncPrefixV1))
		if err != nil {
			// 解密失败（key 不匹配/密文损坏）：保留密文原样 + warn，留待运维排查。
			slog.Warn("credential decrypt failed, keeping ciphertext", "key", k, "err", err)
			out[k] = v
			continue
		}
		out[k] = pt
	}
	return out
}
