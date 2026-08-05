package middleware

import (
	"net/http"
	"strings"

	"am-keramika-backend/auth"
	"am-keramika-backend/repositories"

	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Nedostaje Authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Neispravan Authorization header"})
			return
		}

		claims, err := auth.ParseToken(strings.TrimSpace(parts[1]))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Nevalidan ili istekao token"})
			return
		}

		user, err := repositories.GetUserByID(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije pronađen"})
			return
		}
		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Korisnički nalog je deaktiviran"})
			return
		}

		// Uloga uvijek iz baze, ne iz (moguće zastarjelog) JWT-a.
		auth.SetAuthContext(c, user.ID, user.Username, user.Role)
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		role, err := auth.GetRole(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Korisnik nije autentifikovan"})
			return
		}
		if _, ok := allowed[role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Nemate dozvolu za ovu akciju"})
			return
		}
		c.Next()
	}
}
