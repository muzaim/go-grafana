package middleware

import (
	"strings"

	pkgauth "go-up-monitoring/pkg/auth"
	"go-up-monitoring/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID = "user_id"
	ContextEmail  = "email"
	ContextRole   = "role"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Unauthorized(c, "Invalid Authorization header format. Must be Bearer <token>")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims, err := pkgauth.ValidateToken(tokenStr, secret)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		if claims.TokenType != "access" {
			response.Unauthorized(c, "Invalid token type: access token required")
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextEmail, claims.Email)
		c.Set(ContextRole, claims.Role)

		c.Next()
	}
}
