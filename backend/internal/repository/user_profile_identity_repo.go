package repository

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
	"unsafe"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/telagod/subme/ent"
	"github.com/telagod/subme/ent/authidentity"
	"github.com/telagod/subme/ent/authidentitychannel"
	"github.com/telagod/subme/ent/identityadoptiondecision"
	dbpredicate "github.com/telagod/subme/ent/predicate"
	infraerrors "github.com/telagod/subme/internal/pkg/errors"
	"github.com/telagod/subme/internal/service"
)

var (
	ErrAuthIdentityOwnershipConflict = infraerrors.Conflict(
		"AUTH_IDENTITY_OWNERSHIP_CONFLICT",
		"auth identity already belongs to another user",
	)
	ErrAuthIdentityChannelOwnershipConflict = infraerrors.Conflict(
		"AUTH_IDENTITY_CHANNEL_OWNERSHIP_CONFLICT",
		"auth identity channel already belongs to another user",
	)
	ErrAuthIdentityChannelProviderMismatch = infraerrors.BadRequest(
		"AUTH_IDENTITY_CHANNEL_PROVIDER_MISMATCH",
		"auth identity channel provider must match canonical identity",
	)
)

type ProviderGrantReason string

const (
	ProviderGrantReasonSignup    ProviderGrantReason = "signup"
	ProviderGrantReasonFirstBind ProviderGrantReason = "first_bind"
)

type AuthIdentityKey struct {
	ProviderType    string
	ProviderKey     string
	ProviderSubject string
}

type AuthIdentityChannelKey struct {
	ProviderType   string
	ProviderKey    string
	Channel        string
	ChannelAppID   string
	ChannelSubject string
}

type CreateAuthIdentityInput struct {
	UserID          int64
	Canonical       AuthIdentityKey
	Channel         *AuthIdentityChannelKey
	Issuer          *string
	VerifiedAt      *time.Time
	Metadata        map[string]any
	ChannelMetadata map[string]any
}

type BindAuthIdentityInput = CreateAuthIdentityInput

type CreateAuthIdentityResult struct {
	Identity *dbent.AuthIdentity
	Channel  *dbent.AuthIdentityChannel
}

func (r *CreateAuthIdentityResult) IdentityRef() AuthIdentityKey {
	if r == nil || r.Identity == nil {
		return AuthIdentityKey{}
	}
	ident := r.Identity
	return AuthIdentityKey{
		ProviderType:    ident.ProviderType,
		ProviderKey:     ident.ProviderKey,
		ProviderSubject: ident.ProviderSubject,
	}
}

func (r *CreateAuthIdentityResult) ChannelRef() *AuthIdentityChannelKey {
	if r == nil || r.Channel == nil {
		return nil
	}
	ch := r.Channel
	return &AuthIdentityChannelKey{
		ProviderType:   ch.ProviderType,
		ProviderKey:    ch.ProviderKey,
		Channel:        ch.Channel,
		ChannelAppID:   ch.ChannelAppID,
		ChannelSubject: ch.ChannelSubject,
	}
}

type UserAuthIdentityLookup struct {
	User     *dbent.User
	Identity *dbent.AuthIdentity
	Channel  *dbent.AuthIdentityChannel
}

type ProviderGrantRecordInput struct {
	UserID       int64
	ProviderType string
	GrantReason  ProviderGrantReason
}

type IdentityAdoptionDecisionInput struct {
	PendingAuthSessionID int64
	IdentityID           *int64
	AdoptDisplayName     bool
	AdoptAvatar          bool
}

type sqlQueryExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// repositoryScopedKeyLocks is the global in-process keyed lock registry
// used to serialize identity operations.
var repositoryScopedKeyLocks = createKeyedLockPool()

type scopedKeyLockRegistry struct {
	mu    sync.Mutex
	locks map[string]*scopedKeyLockEntry
}

type scopedKeyLockEntry struct {
	mu   sync.Mutex
	refs int
}

func createKeyedLockPool() *scopedKeyLockRegistry {
	return &scopedKeyLockRegistry{
		locks: make(map[string]*scopedKeyLockEntry),
	}
}

// lock acquires in-process locks on the given keys in sorted order.
// Returns an unlock function that must be called to release all held locks.
func (reg *scopedKeyLockRegistry) lock(keys ...string) func() {
	sortedKeys := deduplicateAndSortKeys(keys...)
	if len(sortedKeys) == 0 {
		return func() {}
	}

	acquired := make([]*scopedKeyLockEntry, 0, len(sortedKeys))
	reg.mu.Lock()
	for _, k := range sortedKeys {
		ent := reg.locks[k]
		if ent == nil {
			ent = &scopedKeyLockEntry{}
			reg.locks[k] = ent
		}
		ent.refs++
		acquired = append(acquired, ent)
	}
	reg.mu.Unlock()

	for _, ent := range acquired {
		ent.mu.Lock()
	}

	return func() {
		// Unlock in reverse order to avoid potential deadlocks.
		for idx := len(acquired) - 1; idx >= 0; idx-- {
			acquired[idx].mu.Unlock()
		}

		reg.mu.Lock()
		defer reg.mu.Unlock()
		for pos, k := range sortedKeys {
			ent := acquired[pos]
			ent.refs--
			if ent.refs == 0 {
				delete(reg.locks, k)
			}
		}
	}
}

// deduplicateAndSortKeys trims, deduplicates, and sorts the provided lock keys.
func deduplicateAndSortKeys(keys ...string) []string {
	if len(keys) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		stripped := strings.TrimSpace(k)
		if stripped == "" {
			continue
		}
		seen[stripped] = struct{}{}
	}
	if len(seen) == 0 {
		return nil
	}

	sorted := make([]string, 0, len(seen))
	for k := range seen {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)
	return sorted
}

// advisoryLockHash computes a 64-bit FNV-1a hash suitable for Postgres advisory locks.
func advisoryLockHash(key string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	return int64(h.Sum64())
}

// lockRepositoryScopedKeys acquires both in-process and Postgres advisory locks
// for the given keys. Returns a release function and any error.
func lockRepositoryScopedKeys(ctx context.Context, client *dbent.Client, exec sqlQueryExecutor, keys ...string) (func(), error) {
	releaseFn := repositoryScopedKeyLocks.lock(keys...)
	sortedKeys := deduplicateAndSortKeys(keys...)
	if len(sortedKeys) == 0 || client == nil || exec == nil || client.Driver().Dialect() != dialect.Postgres {
		return releaseFn, nil
	}

	for _, k := range sortedKeys {
		pgRows, pgErr := exec.QueryContext(ctx, "SELECT pg_advisory_xact_lock($1)", advisoryLockHash(k))
		if pgErr != nil {
			releaseFn()
			return nil, pgErr
		}
		_ = pgRows.Close()
	}
	return releaseFn, nil
}

func (r *userRepository) WithUserProfileIdentityTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	if dbent.TxFromContext(ctx) != nil {
		return fn(ctx)
	}

	txHandle, txErr := r.client.Tx(ctx)
	if txErr != nil {
		return txErr
	}
	defer func() { _ = txHandle.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, txHandle)
	if execErr := fn(txCtx); execErr != nil {
		return execErr
	}
	return txHandle.Commit()
}

func (r *userRepository) CreateAuthIdentity(ctx context.Context, input CreateAuthIdentityInput) (*CreateAuthIdentityResult, error) {
	if validErr := ensureChannelProviderConsistency(input.Canonical, input.Channel); validErr != nil {
		return nil, validErr
	}

	dbClient := clientFromContext(ctx, r.client)

	builder := dbClient.AuthIdentity.Create().
		SetUserID(input.UserID).
		SetProviderType(strings.TrimSpace(input.Canonical.ProviderType)).
		SetProviderKey(strings.TrimSpace(input.Canonical.ProviderKey)).
		SetProviderSubject(strings.TrimSpace(input.Canonical.ProviderSubject)).
		SetMetadata(duplicateMetadataMap(input.Metadata)).
		SetNillableIssuer(input.Issuer).
		SetNillableVerifiedAt(input.VerifiedAt)

	savedIdentity, saveErr := builder.Save(ctx)
	if saveErr != nil {
		return nil, saveErr
	}

	var savedChannel *dbent.AuthIdentityChannel
	if input.Channel != nil {
		var chErr error
		savedChannel, chErr = dbClient.AuthIdentityChannel.Create().
			SetIdentityID(savedIdentity.ID).
			SetProviderType(strings.TrimSpace(input.Channel.ProviderType)).
			SetProviderKey(strings.TrimSpace(input.Channel.ProviderKey)).
			SetChannel(strings.TrimSpace(input.Channel.Channel)).
			SetChannelAppID(strings.TrimSpace(input.Channel.ChannelAppID)).
			SetChannelSubject(strings.TrimSpace(input.Channel.ChannelSubject)).
			SetMetadata(duplicateMetadataMap(input.ChannelMetadata)).
			Save(ctx)
		if chErr != nil {
			return nil, chErr
		}
	}

	return &CreateAuthIdentityResult{Identity: savedIdentity, Channel: savedChannel}, nil
}

func (r *userRepository) GetUserByCanonicalIdentity(ctx context.Context, key AuthIdentityKey) (*UserAuthIdentityLookup, error) {
	found, findErr := clientFromContext(ctx, r.client).AuthIdentity.Query().
		Where(
			authidentity.ProviderTypeEQ(strings.TrimSpace(key.ProviderType)),
			authidentity.ProviderKeyEQ(strings.TrimSpace(key.ProviderKey)),
			authidentity.ProviderSubjectEQ(strings.TrimSpace(key.ProviderSubject)),
		).
		WithUser().
		Only(ctx)
	if findErr != nil {
		return nil, findErr
	}

	return &UserAuthIdentityLookup{
		User:     found.Edges.User,
		Identity: found,
	}, nil
}

func (r *userRepository) GetUserByChannelIdentity(ctx context.Context, key AuthIdentityChannelKey) (*UserAuthIdentityLookup, error) {
	chResult, chErr := clientFromContext(ctx, r.client).AuthIdentityChannel.Query().
		Where(
			authidentitychannel.ProviderTypeEQ(strings.TrimSpace(key.ProviderType)),
			authidentitychannel.ProviderKeyEQ(strings.TrimSpace(key.ProviderKey)),
			authidentitychannel.ChannelEQ(strings.TrimSpace(key.Channel)),
			authidentitychannel.ChannelAppIDEQ(strings.TrimSpace(key.ChannelAppID)),
			authidentitychannel.ChannelSubjectEQ(strings.TrimSpace(key.ChannelSubject)),
		).
		WithIdentity(func(q *dbent.AuthIdentityQuery) {
			q.WithUser()
		}).
		Only(ctx)
	if chErr != nil {
		return nil, chErr
	}

	return &UserAuthIdentityLookup{
		User:     chResult.Edges.Identity.Edges.User,
		Identity: chResult.Edges.Identity,
		Channel:  chResult,
	}, nil
}

func (r *userRepository) ListUserAuthIdentities(ctx context.Context, userID int64) ([]service.UserAuthIdentityRecord, error) {
	rows, queryErr := clientFromContext(ctx, r.client).AuthIdentity.Query().
		Where(authidentity.UserIDEQ(userID)).
		All(ctx)
	if queryErr != nil {
		return nil, queryErr
	}

	result := make([]service.UserAuthIdentityRecord, 0, len(rows))
	for idx := 0; idx < len(rows); idx++ {
		row := rows[idx]
		if row == nil {
			continue
		}
		result = append(result, service.UserAuthIdentityRecord{
			ProviderType:    strings.TrimSpace(row.ProviderType),
			ProviderKey:     strings.TrimSpace(row.ProviderKey),
			ProviderSubject: strings.TrimSpace(row.ProviderSubject),
			VerifiedAt:      row.VerifiedAt,
			Issuer:          row.Issuer,
			Metadata:        duplicateMetadataMap(row.Metadata),
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		})
	}

	return result, nil
}

func (r *userRepository) UnbindUserAuthProvider(ctx context.Context, userID int64, provider string) error {
	normalizedProvider := strings.ToLower(strings.TrimSpace(provider))
	if normalizedProvider == "" || normalizedProvider == "email" {
		return service.ErrIdentityProviderInvalid
	}

	return r.WithUserProfileIdentityTx(ctx, func(txCtx context.Context) error {
		dbClient := clientFromContext(txCtx, r.client)
		matchedIDs, lookupErr := dbClient.AuthIdentity.Query().
			Where(
				authidentity.UserIDEQ(userID),
				authidentity.ProviderTypeEQ(normalizedProvider),
			).
			IDs(txCtx)
		if lookupErr != nil {
			return lookupErr
		}
		if len(matchedIDs) == 0 {
			return nil
		}

		// Clear adoption decision references before deleting identity rows.
		if _, clearErr := dbClient.IdentityAdoptionDecision.Update().
			Where(identityadoptiondecision.IdentityIDIn(matchedIDs...)).
			ClearIdentityID().
			Save(txCtx); clearErr != nil {
			return clearErr
		}
		// Delete associated channel rows.
		if _, delChErr := dbClient.AuthIdentityChannel.Delete().
			Where(authidentitychannel.IdentityIDIn(matchedIDs...)).
			Exec(txCtx); delChErr != nil {
			return delChErr
		}
		// Finally remove the identity rows themselves.
		_, delErr := dbClient.AuthIdentity.Delete().
			Where(
				authidentity.UserIDEQ(userID),
				authidentity.ProviderTypeEQ(normalizedProvider),
			).
			Exec(txCtx)
		return delErr
	})
}

func (r *userRepository) BindAuthIdentityToUser(ctx context.Context, input BindAuthIdentityInput) (*CreateAuthIdentityResult, error) {
	if validErr := ensureChannelProviderConsistency(input.Canonical, input.Channel); validErr != nil {
		return nil, validErr
	}

	var outcome *CreateAuthIdentityResult
	txErr := r.WithUserProfileIdentityTx(ctx, func(txCtx context.Context) error {
		dbClient := clientFromContext(txCtx, r.client)
		canon := input.Canonical

		// Find existing identities that could match (including compatible provider keys).
		existingRecords, qErr := dbClient.AuthIdentity.Query().
			Where(
				authidentity.ProviderTypeEQ(strings.TrimSpace(canon.ProviderType)),
				authidentity.ProviderKeyIn(expandCompatibleProviderKeys(canon.ProviderType, canon.ProviderKey)...),
				authidentity.ProviderSubjectEQ(strings.TrimSpace(canon.ProviderSubject)),
			).
			All(txCtx)
		if qErr != nil {
			return qErr
		}
		ownedIdent := pickOwnedIdentity(existingRecords, input.UserID)
		if ownedIdent == nil && detectIdentityOwnerConflict(existingRecords, input.UserID) {
			return ErrAuthIdentityOwnershipConflict
		}

		if ownedIdent == nil {
			// No existing identity for this user; create one.
			var createErr error
			ownedIdent, createErr = dbClient.AuthIdentity.Create().
				SetUserID(input.UserID).
				SetProviderType(strings.TrimSpace(canon.ProviderType)).
				SetProviderKey(strings.TrimSpace(canon.ProviderKey)).
				SetProviderSubject(strings.TrimSpace(canon.ProviderSubject)).
				SetMetadata(duplicateMetadataMap(input.Metadata)).
				SetNillableIssuer(input.Issuer).
				SetNillableVerifiedAt(input.VerifiedAt).
				Save(txCtx)
			if createErr != nil {
				return createErr
			}
		} else {
			// Update the existing identity with latest metadata.
			resolvedKey := resolveCanonicalProviderKey(canon.ProviderType, ownedIdent.ProviderKey, canon.ProviderKey)
			updater := dbClient.AuthIdentity.UpdateOneID(ownedIdent.ID)
			if resolvedKey != "" && !strings.EqualFold(resolvedKey, ownedIdent.ProviderKey) {
				updater = updater.SetProviderKey(resolvedKey)
			}
			if input.Metadata != nil {
				updater = updater.SetMetadata(duplicateMetadataMap(input.Metadata))
			}
			if input.Issuer != nil {
				updater = updater.SetIssuer(strings.TrimSpace(*input.Issuer))
			}
			if input.VerifiedAt != nil {
				updater = updater.SetVerifiedAt(*input.VerifiedAt)
			}
			var updErr error
			ownedIdent, updErr = updater.Save(txCtx)
			if updErr != nil {
				return updErr
			}
		}

		var boundChannel *dbent.AuthIdentityChannel
		if input.Channel != nil {
			chRecords, chQueryErr := dbClient.AuthIdentityChannel.Query().
				Where(
					authidentitychannel.ProviderTypeEQ(strings.TrimSpace(input.Channel.ProviderType)),
					authidentitychannel.ProviderKeyIn(expandCompatibleProviderKeys(input.Channel.ProviderType, input.Channel.ProviderKey)...),
					authidentitychannel.ChannelEQ(strings.TrimSpace(input.Channel.Channel)),
					authidentitychannel.ChannelAppIDEQ(strings.TrimSpace(input.Channel.ChannelAppID)),
					authidentitychannel.ChannelSubjectEQ(strings.TrimSpace(input.Channel.ChannelSubject)),
				).
				WithIdentity().
				All(txCtx)
			if chQueryErr != nil {
				return chQueryErr
			}
			boundChannel = pickOwnedChannel(chRecords, input.UserID)
			if boundChannel == nil && detectChannelOwnerConflict(chRecords, input.UserID) {
				return ErrAuthIdentityChannelOwnershipConflict
			}
			if boundChannel == nil {
				var chCreateErr error
				boundChannel, chCreateErr = dbClient.AuthIdentityChannel.Create().
					SetIdentityID(ownedIdent.ID).
					SetProviderType(strings.TrimSpace(input.Channel.ProviderType)).
					SetProviderKey(strings.TrimSpace(input.Channel.ProviderKey)).
					SetChannel(strings.TrimSpace(input.Channel.Channel)).
					SetChannelAppID(strings.TrimSpace(input.Channel.ChannelAppID)).
					SetChannelSubject(strings.TrimSpace(input.Channel.ChannelSubject)).
					SetMetadata(duplicateMetadataMap(input.ChannelMetadata)).
					Save(txCtx)
				if chCreateErr != nil {
					return chCreateErr
				}
			} else {
				resolvedChKey := resolveCanonicalProviderKey(input.Channel.ProviderType, boundChannel.ProviderKey, input.Channel.ProviderKey)
				chUpdater := dbClient.AuthIdentityChannel.UpdateOneID(boundChannel.ID).
					SetIdentityID(ownedIdent.ID)
				if resolvedChKey != "" && !strings.EqualFold(resolvedChKey, boundChannel.ProviderKey) {
					chUpdater = chUpdater.SetProviderKey(resolvedChKey)
				}
				if input.ChannelMetadata != nil {
					chUpdater = chUpdater.SetMetadata(duplicateMetadataMap(input.ChannelMetadata))
				}
				var chUpdErr error
				boundChannel, chUpdErr = chUpdater.Save(txCtx)
				if chUpdErr != nil {
					return chUpdErr
				}
			}
		}

		outcome = &CreateAuthIdentityResult{Identity: ownedIdent, Channel: boundChannel}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	return outcome, nil
}

// expandCompatibleProviderKeys returns the set of provider keys that should
// be considered equivalent when matching identities. For WeChat, the legacy
// "wechat" key and the canonical "wechat-main" key are interchangeable.
func expandCompatibleProviderKeys(providerType, providerKey string) []string {
	pt := strings.TrimSpace(strings.ToLower(providerType))
	pk := strings.TrimSpace(providerKey)
	if pk == "" {
		return []string{pk}
	}
	if pt != "wechat" {
		return []string{pk}
	}
	candidates := []string{pk}
	if !strings.EqualFold(pk, "wechat-main") {
		candidates = append(candidates, "wechat-main")
	}
	if !strings.EqualFold(pk, "wechat") {
		candidates = append(candidates, "wechat")
	}
	return candidates
}

// resolveCanonicalProviderKey decides which provider key should be stored,
// preferring "wechat-main" over "wechat" for the wechat provider type.
func resolveCanonicalProviderKey(providerType, currentKey, incomingKey string) string {
	pt := strings.TrimSpace(strings.ToLower(providerType))
	curr := strings.TrimSpace(currentKey)
	incoming := strings.TrimSpace(incomingKey)
	if pt != "wechat" {
		if incoming != "" {
			return incoming
		}
		return curr
	}
	if strings.EqualFold(curr, "wechat") || strings.EqualFold(curr, "wechat-main") || strings.EqualFold(incoming, "wechat-main") {
		return "wechat-main"
	}
	if incoming != "" {
		return incoming
	}
	return curr
}

// providerKeyPriority returns a sort-order rank for a provider key. Lower is better.
func providerKeyPriority(providerType, providerKey string) int {
	pt := strings.TrimSpace(strings.ToLower(providerType))
	pk := strings.TrimSpace(providerKey)
	if pt != "wechat" {
		return 0
	}
	switch {
	case strings.EqualFold(pk, "wechat-main"):
		return 0
	case strings.EqualFold(pk, "wechat"):
		return 2
	default:
		return 1
	}
}

// pickOwnedIdentity selects the best-ranked identity owned by the given user.
func pickOwnedIdentity(records []*dbent.AuthIdentity, ownerID int64) *dbent.AuthIdentity {
	var best *dbent.AuthIdentity
	for _, rec := range records {
		if rec.UserID != ownerID {
			continue
		}
		if best == nil || providerKeyPriority(rec.ProviderType, rec.ProviderKey) < providerKeyPriority(best.ProviderType, best.ProviderKey) {
			best = rec
		}
	}
	return best
}

// detectIdentityOwnerConflict returns true if any record belongs to a different user.
func detectIdentityOwnerConflict(records []*dbent.AuthIdentity, ownerID int64) bool {
	for _, rec := range records {
		if rec.UserID != ownerID {
			return true
		}
	}
	return false
}

// pickOwnedChannel selects the best-ranked channel whose parent identity
// is owned by the given user.
func pickOwnedChannel(records []*dbent.AuthIdentityChannel, ownerID int64) *dbent.AuthIdentityChannel {
	var best *dbent.AuthIdentityChannel
	for _, rec := range records {
		if rec.Edges.Identity == nil || rec.Edges.Identity.UserID != ownerID {
			continue
		}
		if best == nil || providerKeyPriority(rec.ProviderType, rec.ProviderKey) < providerKeyPriority(best.ProviderType, best.ProviderKey) {
			best = rec
		}
	}
	return best
}

// detectChannelOwnerConflict returns true if any channel record belongs to
// a different user (via its parent identity edge).
func detectChannelOwnerConflict(records []*dbent.AuthIdentityChannel, ownerID int64) bool {
	for _, rec := range records {
		if rec.Edges.Identity != nil && rec.Edges.Identity.UserID != ownerID {
			return true
		}
	}
	return false
}

func (r *userRepository) RecordProviderGrant(ctx context.Context, input ProviderGrantRecordInput) (bool, error) {
	sqlExec := resolveTransactionAwareExecutor(ctx, r.sql, r.client)
	if sqlExec == nil {
		return false, fmt.Errorf("no SQL executor available for provider grant recording")
	}

	res, execErr := sqlExec.ExecContext(ctx, `
INSERT INTO user_provider_default_grants (user_id, provider_type, grant_reason)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, provider_type, grant_reason) DO NOTHING`,
		input.UserID,
		strings.TrimSpace(input.ProviderType),
		string(input.GrantReason),
	)
	if execErr != nil {
		return false, execErr
	}
	rowsChanged, affErr := res.RowsAffected()
	if affErr != nil {
		return false, affErr
	}
	return rowsChanged > 0, nil
}

func (r *userRepository) UpsertIdentityAdoptionDecision(ctx context.Context, input IdentityAdoptionDecisionInput) (*dbent.IdentityAdoptionDecision, error) {
	var decision *dbent.IdentityAdoptionDecision
	txErr := r.WithUserProfileIdentityTx(ctx, func(txCtx context.Context) error {
		dbClient := clientFromContext(txCtx, r.client)
		unlock, lockErr := lockRepositoryScopedKeys(
			txCtx,
			dbClient,
			resolveTransactionAwareExecutor(txCtx, r.sql, r.client),
			buildAdoptionLockKeys(input.PendingAuthSessionID, input.IdentityID)...,
		)
		if lockErr != nil {
			return lockErr
		}
		defer unlock()

		// If an identity ID is provided, detach any prior decisions that reference
		// it from a different pending session (ensures one-to-one mapping).
		if input.IdentityID != nil && *input.IdentityID > 0 {
			if _, detachErr := dbClient.IdentityAdoptionDecision.Update().
				Where(
					identityadoptiondecision.IdentityIDEQ(*input.IdentityID),
					dbpredicate.IdentityAdoptionDecision(func(s *entsql.Selector) {
						col := s.C(identityadoptiondecision.FieldPendingAuthSessionID)
						s.Where(entsql.Or(
							entsql.IsNull(col),
							entsql.NEQ(col, input.PendingAuthSessionID),
						))
					}),
				).
				ClearIdentityID().
				Save(txCtx); detachErr != nil {
				return detachErr
			}
		}

		builder := dbClient.IdentityAdoptionDecision.Create().
			SetPendingAuthSessionID(input.PendingAuthSessionID).
			SetAdoptDisplayName(input.AdoptDisplayName).
			SetAdoptAvatar(input.AdoptAvatar).
			SetDecidedAt(time.Now().UTC())
		if input.IdentityID != nil && *input.IdentityID > 0 {
			builder = builder.SetIdentityID(*input.IdentityID)
		}

		upsertedID, upsertErr := builder.
			OnConflictColumns(identityadoptiondecision.FieldPendingAuthSessionID).
			UpdateNewValues().
			ID(txCtx)
		if upsertErr != nil {
			return upsertErr
		}

		var fetchErr error
		decision, fetchErr = dbClient.IdentityAdoptionDecision.Get(txCtx, upsertedID)
		return fetchErr
	})
	if txErr != nil {
		return nil, txErr
	}
	return decision, nil
}

// buildAdoptionLockKeys constructs the advisory lock keys for an identity
// adoption decision based on the session and optional identity ID.
func buildAdoptionLockKeys(sessionID int64, identityID *int64) []string {
	lockKeys := []string{fmt.Sprintf("identity-adoption:pending:%d", sessionID)}
	if identityID != nil && *identityID > 0 {
		lockKeys = append(lockKeys, fmt.Sprintf("identity-adoption:identity:%d", *identityID))
	}
	return lockKeys
}

func (r *userRepository) GetIdentityAdoptionDecisionByPendingAuthSessionID(ctx context.Context, pendingAuthSessionID int64) (*dbent.IdentityAdoptionDecision, error) {
	return clientFromContext(ctx, r.client).IdentityAdoptionDecision.Query().
		Where(identityadoptiondecision.PendingAuthSessionIDEQ(pendingAuthSessionID)).
		Only(ctx)
}

func (r *userRepository) UpdateUserLastLoginAt(ctx context.Context, userID int64, loginAt time.Time) error {
	_, saveErr := clientFromContext(ctx, r.client).User.UpdateOneID(userID).
		SetLastLoginAt(loginAt).
		Save(ctx)
	return saveErr
}

func (r *userRepository) UpdateUserLastActiveAt(ctx context.Context, userID int64, activeAt time.Time) error {
	_, saveErr := clientFromContext(ctx, r.client).User.UpdateOneID(userID).
		SetLastActiveAt(activeAt).
		Save(ctx)
	return saveErr
}

func (r *userRepository) GetUserAvatar(ctx context.Context, userID int64) (*service.UserAvatar, error) {
	sqlExec, resolveErr := r.obtainProfileIdentitySQL(ctx)
	if resolveErr != nil {
		return nil, resolveErr
	}

	sqlRows, queryErr := sqlExec.QueryContext(ctx, `
SELECT storage_provider, storage_key, url, content_type, byte_size, sha256
FROM user_avatars
WHERE user_id = $1`, userID)
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { _ = sqlRows.Close() }()

	if !sqlRows.Next() {
		return nil, sqlRows.Err()
	}

	var avatarData service.UserAvatar
	if scanErr := sqlRows.Scan(
		&avatarData.StorageProvider,
		&avatarData.StorageKey,
		&avatarData.URL,
		&avatarData.ContentType,
		&avatarData.ByteSize,
		&avatarData.SHA256,
	); scanErr != nil {
		return nil, scanErr
	}
	if rowErr := sqlRows.Err(); rowErr != nil {
		return nil, rowErr
	}
	return &avatarData, nil
}

func (r *userRepository) UpsertUserAvatar(ctx context.Context, userID int64, input service.UpsertUserAvatarInput) (*service.UserAvatar, error) {
	sqlExec, resolveErr := r.obtainProfileIdentitySQL(ctx)
	if resolveErr != nil {
		return nil, resolveErr
	}

	_, execErr := sqlExec.ExecContext(ctx, `
INSERT INTO user_avatars (user_id, storage_provider, storage_key, url, content_type, byte_size, sha256, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
ON CONFLICT (user_id) DO UPDATE SET
	storage_provider = EXCLUDED.storage_provider,
	storage_key = EXCLUDED.storage_key,
	url = EXCLUDED.url,
	content_type = EXCLUDED.content_type,
	byte_size = EXCLUDED.byte_size,
	sha256 = EXCLUDED.sha256,
	updated_at = NOW()`,
		userID,
		strings.TrimSpace(input.StorageProvider),
		strings.TrimSpace(input.StorageKey),
		strings.TrimSpace(input.URL),
		strings.TrimSpace(input.ContentType),
		input.ByteSize,
		strings.TrimSpace(input.SHA256),
	)
	if execErr != nil {
		return nil, execErr
	}

	return &service.UserAvatar{
		StorageProvider: strings.TrimSpace(input.StorageProvider),
		StorageKey:      strings.TrimSpace(input.StorageKey),
		URL:             strings.TrimSpace(input.URL),
		ContentType:     strings.TrimSpace(input.ContentType),
		ByteSize:        input.ByteSize,
		SHA256:          strings.TrimSpace(input.SHA256),
	}, nil
}

func (r *userRepository) DeleteUserAvatar(ctx context.Context, userID int64) error {
	sqlExec, resolveErr := r.obtainProfileIdentitySQL(ctx)
	if resolveErr != nil {
		return resolveErr
	}
	_, execErr := sqlExec.ExecContext(ctx, `DELETE FROM user_avatars WHERE user_id = $1`, userID)
	return execErr
}

// duplicateMetadataMap creates a shallow copy of the provided metadata map.
// Returns an empty map if the input is nil or empty.
func duplicateMetadataMap(src map[string]any) map[string]any {
	if len(src) == 0 {
		return map[string]any{}
	}
	dst := make(map[string]any, len(src))
	for key, val := range src {
		dst[key] = val
	}
	return dst
}

// ensureChannelProviderConsistency validates that, when a channel is present,
// its provider type and key match the canonical identity.
func ensureChannelProviderConsistency(canonical AuthIdentityKey, channel *AuthIdentityChannelKey) error {
	if channel == nil {
		return nil
	}

	canonType := strings.TrimSpace(canonical.ProviderType)
	canonKey := strings.TrimSpace(canonical.ProviderKey)
	chType := strings.TrimSpace(channel.ProviderType)
	chKey := strings.TrimSpace(channel.ProviderKey)

	if canonType != chType || canonKey != chKey {
		return ErrAuthIdentityChannelProviderMismatch
	}

	return nil
}

// resolveTransactionAwareExecutor returns a SQL executor that respects the
// current transaction context, falling back to the provided raw executor
// or extracting one from the ent client.
func resolveTransactionAwareExecutor(ctx context.Context, fallbackExec sqlExecutor, entClient *dbent.Client) sqlQueryExecutor {
	if activeTx := dbent.TxFromContext(ctx); activeTx != nil {
		if extracted := extractSQLFromEntClient(activeTx.Client()); extracted != nil {
			return extracted
		}
	}
	if fallbackExec != nil {
		return fallbackExec
	}
	return extractSQLFromEntClient(entClient)
}

// txAwareSQLExecutor is kept as a package-level alias.
func txAwareSQLExecutor(ctx context.Context, fallback sqlExecutor, client *dbent.Client) sqlQueryExecutor {
	return resolveTransactionAwareExecutor(ctx, fallback, client)
}

func (r *userRepository) obtainProfileIdentitySQL(ctx context.Context) (sqlQueryExecutor, error) {
	executor := resolveTransactionAwareExecutor(ctx, r.sql, r.client)
	if executor == nil {
		return nil, fmt.Errorf("no SQL executor available for profile identity operations")
	}
	return executor, nil
}

// extractSQLFromEntClient uses reflection to obtain the underlying SQL driver
// from an ent client, enabling raw SQL queries within the same connection.
func extractSQLFromEntClient(c *dbent.Client) sqlQueryExecutor {
	if c == nil {
		return nil
	}

	rv := reflect.ValueOf(c).Elem()
	cfgField := rv.FieldByName("config")
	drvField := cfgField.FieldByName("driver")
	if !drvField.IsValid() {
		return nil
	}

	driverIface := reflect.NewAt(drvField.Type(), unsafe.Pointer(drvField.UnsafeAddr())).Elem().Interface()
	executor, ok := driverIface.(sqlQueryExecutor)
	if !ok {
		return nil
	}
	return executor
}
