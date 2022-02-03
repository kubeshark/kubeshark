package middlewares

import (
	"github.com/up9inc/mizu/agent/pkg/config"
	"github.com/up9inc/mizu/agent/pkg/providers/user"
	"github.com/up9inc/mizu/agent/pkg/providers/userRoles"

	"github.com/gin-gonic/gin"
	ory "github.com/ory/kratos-client-go"
	"github.com/up9inc/mizu/shared/logger"
)

func RequiresAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// auth is irrelevant for ephermeral mizu
		if !config.Config.StandaloneMode {
			c.Next()
			return
		}

		verifyKratosSessionForRequest(c)
		if !c.IsAborted() {
			c.Next()
		}
	}
}

func RequiresAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// auth is irrelevant for ephermeral mizu
		if !config.Config.StandaloneMode {
			c.Next()
			return
		}

		session := verifyKratosSessionForRequest(c)
		if c.IsAborted() {
			return
		}

		traits := session.Identity.Traits.(map[string]interface{})
		username := traits["username"].(string)

		userRole, err := userRoles.GetUserSystemRole(username)
		if err != nil {
			logger.Log.Errorf("error checking user role %v", err)
			c.AbortWithStatusJSON(403, gin.H{"error": "unknown auth error occured"})
		} else if userRole != userRoles.AdminRole {
			logger.Log.Warningf("user %s attempted to call an admin only endpoint with insufficient privileges", username)
			c.AbortWithStatusJSON(403, gin.H{"error": "unauthorized"})
		} else {
			c.Next()
		}
	}
}

func verifyKratosSessionForRequest(c *gin.Context) *ory.Session {
	token := c.GetHeader(user.SessionTokenHeader)
	if token == "" {
		token = c.Query("sessionToken")
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "this request has no token"})
			return nil
		}
	}

	if session, err := user.VerifyToken(token, c.Request.Context()); err != nil {
		logger.Log.Errorf("error verifying token %v", err)
		c.AbortWithStatusJSON(401, gin.H{"error": "unknown auth error occured"})
		return nil
	} else if session == nil {
		c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
		return nil
	} else {
		return session
	}
}
