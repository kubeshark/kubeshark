package middlewares

import (
	"mizuserver/pkg/config"
	"mizuserver/pkg/providers"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/up9inc/mizu/shared/logger"
)

const cachedValidTokensRetainmentTime = time.Minute * 1

var cachedValidTokens = cache.New(cachedValidTokensRetainmentTime, cachedValidTokensRetainmentTime)

func RequiresAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// auth is irrelevant for ephermeral mizu
		if !config.Config.StandaloneMode {
			c.Next()
			return
		}

		token := c.GetHeader("x-session-token")
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "token header is empty"})
			return
		}

		if _, isTokenCached := cachedValidTokens.Get(token); isTokenCached {
			c.Next()
			return
		}

		if isTokenValid, err := providers.VerifyToken(token, c.Request.Context()); err != nil {
			logger.Log.Errorf("error verifying token %s", err)
			c.AbortWithStatusJSON(401, gin.H{"error": "unknown auth error occured"})
			return
		} else if !isTokenValid {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		cachedValidTokens.Set(token, true, cachedValidTokensRetainmentTime)

		c.Next()
	}
}
