package oauthsync

import "context"

type PostLoginSyncer interface {
	SyncAfterLogin(ctx context.Context, providerType string, userID int64, claims map[string]any)
	SyncAfterRegistration(ctx context.Context, providerType string, userID int64, claims map[string]any)
}
