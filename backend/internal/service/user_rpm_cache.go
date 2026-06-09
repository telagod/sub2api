package service

import "context"

// UserRPMCache provides per-user and per-(user,group) RPM counters.
//
// Unlike the account-level RPMCache that aggregates by upstream AI provider
// account, this interface aggregates by end user (or user+group pair) to
// prevent rate-limit bypass through multiple API keys.
type UserRPMCache interface {
	// IncrementUserGroupRPM atomically increments the per-minute counter for
	// the given (user, group) pair and returns the new value.
	IncrementUserGroupRPM(ctx context.Context, userID, groupID int64) (count int, err error)

	// IncrementUserRPM atomically increments the per-minute counter for the
	// given user and returns the new value.
	IncrementUserRPM(ctx context.Context, userID int64) (count int, err error)

	// GetUserGroupRPM returns the current minute's RPM for a (user, group)
	// pair without incrementing.
	GetUserGroupRPM(ctx context.Context, userID, groupID int64) (count int, err error)

	// GetUserRPM returns the current minute's RPM for a user without
	// incrementing.
	GetUserRPM(ctx context.Context, userID int64) (count int, err error)
}
