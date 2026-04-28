package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/pkg/jwtutil"
	"github.com/redis/go-redis/v9"
)

// SlidingWindow returns a Gin middleware that limits requests per user (or IP)
// to `limit` requests within `window` duration using Redis sorted sets.
//
// Algorithm: store each request as a member (timestamp+random) in a sorted set
// keyed by user. On each request:
//  1. Remove members older than now-window
//  2. Count remaining members
//  3. If count >= limit → 429
//  4. Add current timestamp as new member
func SlidingWindow(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := resolveIdentifier(c)
		key := fmt.Sprintf("rate:%s", identifier)
		now := time.Now()
		windowStart := now.Add(-window)

		ctx := context.Background()

		pipe := rdb.Pipeline()
		// Remove old entries outside the window
		pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))
		// Count current entries
		countCmd := pipe.ZCard(ctx, key)
		// Add current request
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(now.UnixNano()), Member: now.UnixNano()})
		// Expire key after window so Redis doesn't accumulate stale keys
		pipe.Expire(ctx, key, window+time.Second)

		if _, err := pipe.Exec(ctx); err != nil {
			// On Redis error, allow through (fail open)
			c.Next()
			return
		}

		count := countCmd.Val()
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(limit)-count-1)))

		if count >= int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			return
		}

		c.Next()
	}
}

// resolveIdentifier uses the authenticated userID if available, otherwise falls back to client IP.
func resolveIdentifier(c *gin.Context) string {
	if uid := jwtutil.GetUserID(c); uid != "" {
		return "user:" + uid
	}
	return "ip:" + c.ClientIP()
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
