package oauth

import (
	"base/core/app/users"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v80"
	"github.com/stripe/stripe-go/v80/customer"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

type OAuthService struct {
	DB     *gorm.DB
	Config *OAuthConfig
}

func NewOAuthService(db *gorm.DB, config *OAuthConfig) *OAuthService {
	return &OAuthService{
		DB:     db,
		Config: config,
	}
}

func (s *OAuthService) ProcessGoogleOAuth(idToken string) (*OAuthUser, error) {
	email, name, username, picture, providerID, err := s.handleGoogleOAuth(idToken)
	if err != nil {
		return nil, err
	}

	return s.processUser(email, name, username, picture, "google", providerID, idToken)
}

func (s *OAuthService) ProcessFacebookOAuth(accessToken string) (*OAuthUser, error) {
	email, name, username, picture, providerID, err := s.handleFacebookOAuth(accessToken)
	if err != nil {
		return nil, err
	}

	return s.processUser(email, name, username, picture, "facebook", providerID, accessToken)
}

func (s *OAuthService) handleGoogleOAuth(idToken string) (email, name, username, picture, providerID string, err error) {
	payload, err := idtoken.Validate(context.Background(), idToken, s.Config.Google.ClientID)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("invalid ID token: %w", err)
	}

	email, _ = payload.Claims["email"].(string)
	name, _ = payload.Claims["name"].(string)
	username = strings.ToLower(strings.ReplaceAll(name, " ", ""))
	picture, _ = payload.Claims["picture"].(string)
	providerID, _ = payload.Claims["sub"].(string)

	return email, name, username, picture, providerID, nil
}

func (s *OAuthService) handleFacebookOAuth(accessToken string) (email, name, username, picture, providerID string, err error) {
	url := fmt.Sprintf("https://graph.facebook.com/me?fields=id,name,email,picture.type(large)&access_token=%s", accessToken)

	resp, err := http.Get(url)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to fetch user data from Facebook: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to read Facebook response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to parse Facebook response: %w", err)
	}

	providerID, _ = result["id"].(string)
	name, _ = result["name"].(string)
	email, _ = result["email"].(string)
	username = strings.ToLower(strings.ReplaceAll(name, " ", ""))

	if pictureData, ok := result["picture"].(map[string]interface{}); ok {
		if data, ok := pictureData["data"].(map[string]interface{}); ok {
			picture, _ = data["url"].(string)
		}
	}

	return email, name, username, picture, providerID, nil
}

func (s *OAuthService) processUser(email, name, username, picture, provider, providerID, token string) (*OAuthUser, error) {
	var user OAuthUser
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new user
			user = OAuthUser{
				User: users.User{
					Email:    email,
					Name:     name,
					Username: s.generateUniqueUsername(username),
					Avatar:   picture,
				},
				Provider:       provider,
				ProviderID:     providerID,
				AccessToken:    token,
				OAuthLastLogin: time.Now(),
			}

			// Create Stripe customer
			customerParams := &stripe.CustomerParams{
				Email: stripe.String(user.Email),
				Name:  stripe.String(user.Name),
			}
			cust, err := customer.New(customerParams)
			if err != nil {
				return nil, fmt.Errorf("failed to create Stripe customer: %w", err)
			}
			user.StripeID = cust.ID

			if err := s.DB.Create(&user).Error; err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to query user: %w", err)
		}
	} else {
		// Update existing user
		user.Name = name
		user.Avatar = picture
		user.Provider = provider
		user.ProviderID = providerID
		user.AccessToken = token
		user.OAuthLastLogin = time.Now()
		if err := s.DB.Save(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Update or create AuthProvider
	authProvider := AuthProvider{
		UserID:      user.Id,
		Provider:    provider,
		ProviderID:  providerID,
		AccessToken: token,
		LastLogin:   time.Now(),
	}
	if err := s.DB.Where("user_id = ? AND provider = ?", user.Id, provider).
		Assign(authProvider).
		FirstOrCreate(&authProvider).Error; err != nil {
		return nil, fmt.Errorf("failed to update or create auth provider: %w", err)
	}

	return &user, nil
}

func (s *OAuthService) generateUniqueUsername(baseUsername string) string {
	username := baseUsername
	counter := 1
	for {
		var existingUser users.User
		if s.DB.Where("username = ?", username).First(&existingUser).Error == gorm.ErrRecordNotFound {
			break
		}
		username = fmt.Sprintf("%s%d", baseUsername, counter)
		counter++
	}
	return username
}
