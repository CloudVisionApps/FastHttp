package admin

import (
	"crypto/subtle"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled  bool
	Username string
	Password string
	Token    string // Optional: JWT token for API access
}

// NewAuthConfig creates auth config from provided values
func NewAuthConfig(enabled bool, username, password, token string) AuthConfig {
	return AuthConfig{
		Enabled:  enabled,
		Username: username,
		Password: password,
		Token:    token,
	}
}

// BasicAuthMiddleware provides HTTP Basic Authentication
func BasicAuthMiddleware(config AuthConfig) fiber.Handler {
	if !config.Enabled {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		// Check for token in header (for API access)
		if config.Token != "" {
			authHeader := c.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if subtle.ConstantTimeCompare([]byte(token), []byte(config.Token)) == 1 {
					return c.Next()
				}
			}
		}

		// Check for Basic Auth
		username, password, ok := c.BasicAuth()
		if !ok {
			c.Set("WWW-Authenticate", `Basic realm="FastHTTP Admin"`)
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Use constant-time comparison to prevent timing attacks
		usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(config.Username)) == 1
		passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(config.Password)) == 1

		if !usernameMatch || !passwordMatch {
			c.Set("WWW-Authenticate", `Basic realm="FastHTTP Admin"`)
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		return c.Next()
	}
}

// IPWhitelistMiddleware restricts access to specific IPs
func IPWhitelistMiddleware(allowedIPs []string) fiber.Handler {
	if len(allowedIPs) == 0 {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	allowedMap := make(map[string]bool)
	for _, ip := range allowedIPs {
		allowedMap[ip] = true
	}

	return func(c *fiber.Ctx) error {
		clientIP := c.IP()
		
		// Check if IP is in whitelist
		if !allowedMap[clientIP] {
			return c.Status(403).JSON(fiber.Map{
				"error": "Forbidden: IP not whitelisted",
			})
		}

		return c.Next()
	}
}
