package middleware

import (
	"os"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	if allowedOrigins := splitEnvList(os.Getenv("CORS_ALLOWED_ORIGINS")); len(allowedOrigins) > 0 {
		config.AllowOrigins = allowedOrigins
		config.AllowCredentials = true
	} else {
		config.AllowAllOrigins = true
		config.AllowCredentials = false
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Accept",
		"Authorization",
		"Cache-Control",
		"X-Requested-With",
		"X-Api-Key",
		"X-Goog-Api-Key",
		"Anthropic-Version",
		"Anthropic-Beta",
		"OpenAI-Beta",
		"OpenAI-Organization",
		"X-Request-Id",
		"New-Api-User",
	}
	return cors.New(config)
}

func splitEnvList(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}

func PoweredBy() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-New-Api-Version", common.Version)
		c.Next()
	}
}
