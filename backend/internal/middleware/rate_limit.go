package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oralhistory/oralhistory/internal/constants"
	"github.com/oralhistory/oralhistory/internal/util"
	"github.com/redis/go-redis/v9"
)

// RateLimit 基于 Redis 的固定窗口限流。
func RateLimit(rdb *redis.Client, limit int, window time.Duration, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil {
			c.Next()
			return
		}
		key := fmt.Sprintf("rl:%s:%s", c.ClientIP(), c.FullPath())
		ctx := context.Background()
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			logger.Warn("rate limit redis error", "error", err)
			c.Next()
			return
		}
		if count == 1 {
			rdb.Expire(ctx, key, window)
		}
		if count > int64(limit) {
			logger.Warn(fmt.Sprintf(constants.LogRateLimited, c.ClientIP(), key, limit))
			util.Fail(c, http.StatusTooManyRequests, constants.CodeRateLimited, constants.MsgTooManyRequests)
			c.Abort()
			return
		}
		c.Next()
	}
}
