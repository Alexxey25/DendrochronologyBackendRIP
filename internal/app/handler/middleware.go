package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"metoda/internal/app/auth"
)

const (
	// AuthCookieName имя cookie с JWT
	AuthCookieName = "auth_token"
	ctxUserID      = "auth_user_id"
	ctxIsModerator = "auth_is_moderator"
)

func bearerPrefix() string { return "Bearer " }

// CORSMiddleware — для фронта (Vite), GitHub Pages и Tauri WebView.
// Нельзя одновременно ставить Origin * и Allow-Credentials: true (браузер это отклонит);
// при переданном Origin отражаем его и включаем credentials.
//
// WebKit/WebView в Tauri иногда не шлёт Origin на cross-origin fetch, но шлёт Referer —
// тогда восстанавливаем допустимый Allow-Origin из Referer (dev: Vite на 127.0.0.1:3000).
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if ref := c.Request.Header.Get("Referer"); ref != "" {
			if u, err := url.Parse(ref); err == nil && u.Scheme != "" && u.Host != "" {
				c.Writer.Header().Set("Access-Control-Allow-Origin", u.Scheme+"://"+u.Host)
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Vary", "Origin")
		c.Writer.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With",
		)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// extractJWT из cookie или заголовка Authorization.
func extractJWT(r *http.Request) string {
	if c, err := r.Cookie(AuthCookieName); err == nil && c.Value != "" {
		return c.Value
	}
	h := r.Header.Get("Authorization")
	p := bearerPrefix()
	if h != "" && strings.HasPrefix(h, p) {
		return strings.TrimPrefix(h, p)
	}
	return ""
}

// AuthMiddleware проверяет JWT и кладёт user_id / is_moderator в контекст Gin.
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractJWT(c.Request)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}
		blacklisted, err := h.Repository.IsTokenBlacklisted(context.Background(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "description": err.Error()})
			return
		}
		if blacklisted {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}
		claims, err := auth.ParseAndValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}
		uid, err := auth.UserIDFromClaims(claims)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}
		h.Repository.SetUserID(int(uid))
		c.Set(ctxUserID, uid)
		c.Set(ctxIsModerator, auth.IsModeratorFromClaims(claims))
		c.Next()
	}
}

// RequireModerator разрешает только пользователям с is_moderator (после AuthMiddleware).
func (h *Handler) RequireModerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get(ctxIsModerator)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "forbidden"})
			return
		}
		isMod, ok := v.(bool)
		if !ok || !isMod {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "forbidden"})
			return
		}
		c.Next()
	}
}

func authUserIDUint(ctx *gin.Context) (uint, error) {
	v, ok := ctx.Get(ctxUserID)
	if !ok {
		return 0, errors.New("нет user_id в контексте")
	}
	id, ok := v.(uint)
	if !ok || id == 0 {
		return 0, errors.New("неверный user_id")
	}
	return id, nil
}

func isModeratorFromCtx(ctx *gin.Context) bool {
	v, ok := ctx.Get(ctxIsModerator)
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}
