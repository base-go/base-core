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
	if config.Google.ClientID == "" || config.Google.ClientSecret == "" {
		log.Fatal("Missing Google Client ID or Secret")
	}
	if config.Facebook.ClientID == "" || config.Facebook.ClientSecret == "" {
		log.Fatal("Missing Facebook Client ID or Secret")
	}
	if config.Apple.ClientID == "" || config.Apple.ClientSecret == "" {
		log.Fatal("Missing Apple Client ID or Secret")
	}
	if config.JWTSecret == "" {
		log.Fatal("Missing JWT Secret")
	}
	log.Println("OAuth configuration validated successfully")
}
