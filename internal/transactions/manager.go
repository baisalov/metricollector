package transactions

import "context"

type DiscardManager struct{}

func (m DiscardManager) Do(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}
