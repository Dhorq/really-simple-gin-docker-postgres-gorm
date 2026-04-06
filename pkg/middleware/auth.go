package middleware

import (
	"net/http"

	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/pkg/auth"
	"github.com/Dhorq/really-simple-gin-docker-postgres-gorm/pkg/response"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// tokenString := c.GetHeader("Authorization")

		// if tokenString == "" {
		// 	response.Error(c, http.StatusUnauthorized, "token is required")
		// 	c.Abort()
		// 	return
		// }

		// if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		// 	tokenString = tokenString[7:]
		// }

		tokenString, err := c.Cookie("access_token") // pake cookie httpCookie
		if err != nil || tokenString == "" {
			response.Error(c, http.StatusUnauthorized, "token is required")
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}
