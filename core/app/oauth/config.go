package oauth

import (
	"log"
	"os"
)

type OAuthConfig struct {
	Google    ProviderConfig
	Facebook  ProviderConfig
	Apple     ProviderConfig
	JWTSecret string
}

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func LoadConfig() *OAuthConfig {
	log.Println("Loading OAuth configuration")
	config := &OAuthConfig{
		Google: ProviderConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		},
		Facebook: ProviderConfig{
			ClientID:     os.Getenv("FACEBOOK_CLIENT_ID"),
			ClientSecret: os.Getenv("FACEBOOK_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("FACEBOOK_REDIRECT_URL"),
		},
		Apple: ProviderConfig{
			ClientID:     os.Getenv("APPLE_CLIENT_ID"),
			ClientSecret: os.Getenv("APPLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("APPLE_REDIRECT_URL"),
		},
		JWTSecret: os.Getenv("JWT_SECRET"),
	}
	log.Println("OAuth configuration loaded successfully")
	return config
}

func ValidateConfig(config *OAuthConfig) {
	log.Println("Validating OAuth configuration")

	// Check if at least one provider is configured
	hasProvider := false
	if config.Google.ClientID != "" && config.Google.ClientSecret != "" {
		hasProvider = true
		log.Println("Google OAuth provider configured")
	}
	if config.Facebook.ClientID != "" && config.Facebook.ClientSecret != "" {
		hasProvider = true
		log.Println("Facebook OAuth provider configured")
	}
	if config.Apple.ClientID != "" && config.Apple.ClientSecret != "" {
		hasProvider = true
		log.Println("Apple OAuth provider configured")
	}

	if !hasProvider {
		log.Println("Warning: No OAuth providers configured. OAuth functionality will be disabled.")
	}

	if config.JWTSecret == "" {
		log.Println("Warning: JWT Secret not configured. Using default for development.")
		config.JWTSecret = "default-jwt-secret-for-development"
	}

	log.Println("OAuth configuration validated successfully")
}
