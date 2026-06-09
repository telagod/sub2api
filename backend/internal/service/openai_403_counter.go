package service

import "context"

// OpenAI403CounterCache tracks consecutive 403 failures per OpenAI account.
type OpenAI403CounterCache interface {
	// IncrementOpenAI403Count atomically increments the 403 counter and returns the updated count.
	IncrementOpenAI403Count(ctx context.Context, accountID int64, windowMinutes int) (int64, error)
	// ResetOpenAI403Count clears the counter after a successful request.
	ResetOpenAI403Count(ctx context.Context, accountID int64) error
}
