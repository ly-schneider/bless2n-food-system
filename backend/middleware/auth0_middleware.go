package middleware

import (
	"backend/config"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/labstack/echo/v4"
)

func Auth0(cfg *config.Config) echo.MiddlewareFunc {
	log.Print("Initializing Auth0 middleware...")

	jwksURL := cfg.AUTH0.Domain + "/.well-known/jwks.json"
	u, err := url.Parse(jwksURL)
	if err != nil {
		log.Printf("Failed to parse JWKS URL: %v", err)
		panic(err)
	}

	provider := jwks.NewCachingProvider(u, cfg.AUTH0.CacheTTL)

	issuer := cfg.AUTH0.Domain
	if !strings.HasSuffix(issuer, "/") {
		issuer += "/"
	}

	v, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		cfg.AUTH0.Audience,
		[]string{issuer},
	)
	if err != nil {
		log.Printf("Failed to create JWT validator: %v", err)
		panic(err)
	}

	stdMW := jwtmiddleware.New(v.ValidateToken)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log.Printf("Auth0 middleware called for path: %s", c.Request().URL.Path)

			// Log if Authorization header exists
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" {
				log.Printf("Authorization header found: %s", strings.Replace(authHeader, "Bearer ", "Bearer [REDACTED]", 1))
			} else {
				log.Printf("No Authorization header found")
			}

			h := stdMW.CheckJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Printf("JWT validation successful for path: %s", r.URL.Path)

				// propagate validated claims
				if claims, ok := r.Context().
					Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims); ok {
					c.Set("claims", claims)

					// Log JWT payload
					payloadBytes, err := json.MarshalIndent(claims, "", "  ")
					if err != nil {
						log.Printf("Error marshaling JWT claims: %v", err)
					} else {
						log.Printf("JWT Payload: %s", string(payloadBytes))
					}
				} else {
					log.Printf("No claims found in context")
				}

				c.SetRequest(r) // update Echo's request with the new context
				_ = next(c)
			}))

			err := func() error {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Recovered from panic in Auth0 middleware: %v", r)
					}
				}()
				h.ServeHTTP(c.Response().Writer, c.Request())
				return nil
			}()

			log.Printf("Auth0 middleware completed for path: %s", c.Request().URL.Path)
			return err
		}
	}
}
