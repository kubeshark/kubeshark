package middlewares

import (
	"errors"
	"mizuserver/pkg/config"
	"mizuserver/pkg/providers"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/shared/logger"
)

const errorMessage = "unknown authentication error occured"

func RequiresAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// authentication is irrelevant for ephermeral mizu
		if !config.Config.StandaloneMode {
			c.Next()
			return
		}

		token, err := c.Cookie("x-session-token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				c.AbortWithStatusJSON(401, gin.H{"error": "could not find session cookie"})
			} else {
				logger.Log.Errorf("error reading cookie %s", err)
				c.AbortWithStatusJSON(500, gin.H{"error": errorMessage})
			}

			return
		}

		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "token cookie is empty"})
			return
		}

		if isTokenValid, err := providers.VerifyToken(token, c.Request.Context()); err != nil {
			logger.Log.Errorf("error verifying token %s", err)
			c.AbortWithStatusJSON(500, gin.H{"error": errorMessage})
			return
		} else if !isTokenValid {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		c.Next()
	}
}
