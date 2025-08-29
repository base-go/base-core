package middleware

import (
	"base/core/router"
)

func CORSMiddleware(allowedOrigins []string) router.MiddlewareFunc {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			origin := c.GetHeader("Origin")

			// Allow all origins if "*" is present, otherwise match against allowedOrigins
			allowOrigin := ""
			if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
				allowOrigin = "*"
			} else {
				for _, o := range allowedOrigins {
					if o == origin {
						allowOrigin = origin
						break
					}
				}
			}

			if allowOrigin != "" {
				c.SetHeader("Access-Control-Allow-Origin", allowOrigin)
				c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				c.SetHeader("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Api-Key, Base-Orgid")
				c.SetHeader("Access-Control-Expose-Headers", "Content-Length, Content-Type")
				c.SetHeader("Access-Control-Allow-Credentials", "true")
				c.SetHeader("Access-Control-Max-Age", "43200") // 12 hours
			}

			// Handle preflight OPTIONS requests
			if c.Request.Method == "OPTIONS" {
				return c.NoContent()
			}

			return next(c)
		}
	}
}
