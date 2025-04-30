package feedrepo

import (
	"context"
	"time"
)

type IFeedRepository interface {
	Cache(ctx context.Context, userID int, tag string, page int, limit int, feed interface{}) error
	GetCached(ctx context.Context, userID int, tag string, page int, limit int, result interface{}) (bool, error)
	AddPollToTag(ctx context.Context, tag string, pollID int, timestamp time.Time) error
	GetPollIDsByTag(ctx context.Context, tag string, page int, limit int) ([]int, error)
	GetFilteredFeedPollIDs(ctx context.Context, userID int, tag string, page int, limit int) ([]int, error)
	InvalidateUserFeeds(ctx context.Context, userID uint) error
}
