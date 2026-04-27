package event

import "context"

type Publisher interface {
	Publish(ctx context.Context, topic, key, value string) error
	Close() error
}
