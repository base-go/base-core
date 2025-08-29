package oauth

import (
	"base/core/app/profile"
	"base/core/storage"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

type OAuthService struct {
	DB            *gorm.DB
	Config        *OAuthConfig
	ActiveStorage *storage.ActiveStorage
}

func NewOAuthService(db *gorm.DB, config *OAuthConfig, activeStorage *storage.ActiveStorage) *OAuthService {
	return &OAuthService{
		DB:            db,
		Config:        config,
		ActiveStorage: activeStorage,
	}
}

func (s *OAuthService) ProcessAppleOAuth(idToken string) (*OAuthUser, error) {
	email, name, username, picture, providerId, err := s.handleAppleOAuth(idToken)
	if err != nil {
		return nil, err
	}

	return s.processUser(email, name, username, picture, "apple", providerId, idToken)
}

func (s *OAuthService) ProcessGoogleOAuth(idToken string) (*OAuthUser, error) {
	email, name, username, picture, providerId, err := s.handleGoogleOAuth(idToken)
	if err != nil {
		return nil, err
	}

	return s.processUser(email, name, username, picture, "google", providerId, idToken)
}

func (s *OAuthService) ProcessFacebookOAuth(accessToken string) (*OAuthUser, error) {
	email, name, username, picture, providerId, err := s.handleFacebookOAuth(accessToken)
	if err != nil {
		return nil, err
	}

	return s.processUser(email, name, username, picture, "facebook", providerId, accessToken)
}

func (s *OAuthService) handleAppleOAuth(idToken string) (email, name, username, picture, providerId string, err error) {
	payload, err := idtoken.Validate(context.Background(), idToken, s.Config.Apple.ClientId)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("invalid Id token: %w", err)
	}

	email, _ = payload.Claims["email"].(string)
	name, _ = payload.Claims["name"].(string)
	username = strings.ToLower(strings.ReplaceAll(name, " ", ""))
	picture, _ = payload.Claims["picture"].(string)
	providerId, _ = payload.Claims["sub"].(string)

	return email, name, username, picture, providerId, nil
}

func (s *OAuthService) handleGoogleOAuth(idToken string) (email, name, username, picture, providerId string, err error) {
	payload, err := idtoken.Validate(context.Background(), idToken, s.Config.Google.ClientId)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("invalid Id token: %w", err)
	}

	email, _ = payload.Claims["email"].(string)
	name, _ = payload.Claims["name"].(string)
	username = strings.ToLower(strings.ReplaceAll(name, " ", ""))
	picture, _ = payload.Claims["picture"].(string)
	providerId, _ = payload.Claims["sub"].(string)

	return email, name, username, picture, providerId, nil
}

func (s *OAuthService) handleFacebookOAuth(accessToken string) (email, name, username, picture, providerId string, err error) {
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

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to parse Facebook response: %w", err)
	}

	providerId, _ = result["id"].(string)
	name, _ = result["name"].(string)
	email, _ = result["email"].(string)
	username = strings.ToLower(strings.ReplaceAll(name, " ", ""))

	if pictureData, ok := result["picture"].(map[string]any); ok {
		if data, ok := pictureData["data"].(map[string]any); ok {
			picture, _ = data["url"].(string)
		}
	}

	return email, name, username, picture, providerId, nil
}

func (s *OAuthService) processUser(email, name, username, pictureURL, provider, providerId, token string) (*OAuthUser, error) {
	var user OAuthUser
	err := s.DB.Where("email = ?", email).First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new user
			user = OAuthUser{
				User: profile.User{
					Email:     email,
					FirstName: name[:strings.Index(name, " ")],
					LastName:  name[strings.Index(name, " ")+1:],
					Username:  s.generateUniqueUsername(username),
				},
				Provider:       provider,
				ProviderId:     providerId,
				AccessToken:    token,
				OAuthLastLogin: time.Now(),
			}

			// Fetch and attach avatar if URL is provided
			if pictureURL != "" {
				attachment, err := s.fetchAndAttachAvatar(&user, pictureURL)
				if err == nil {
					user.User.Avatar = attachment
				} else {
					// Log this failure but proceed with user creation
					fmt.Printf("failed to fetch and attach avatar: %v\n", err)
				}
			}

			// Create the user in the database
			if err := s.DB.Create(&user).Error; err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to query user: %w", err)
		}
	} else {
		// Update existing user
		user.User.FirstName = name[:strings.Index(name, " ")]
		user.User.LastName = name[strings.Index(name, " ")+1:]
		user.Provider = provider
		user.ProviderId = providerId
		user.AccessToken = token
		user.OAuthLastLogin = time.Now()

		// Update avatar if a new URL is provided
		if pictureURL != "" {
			attachment, err := s.fetchAndAttachAvatar(&user, pictureURL)
			if err == nil {
				user.User.Avatar = attachment
			} else {
				fmt.Printf("failed to fetch and attach avatar: %v\n", err)
			}
		}

		if err := s.DB.Save(&user).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Update or create AuthProvider
	authProvider := AuthProvider{
		UserId:      user.Id,
		Provider:    provider,
		ProviderId:  providerId,
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

// fetchAndAttachAvatar downloads the avatar from the URL and attaches it to the user using ActiveStorage.

func (s *OAuthService) fetchAndAttachAvatar(user *OAuthUser, avatarURL string) (*storage.Attachment, error) {
	// Download the avatar from the URL
	resp, err := http.Get(avatarURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download avatar: %w", err)
	}
	defer resp.Body.Close()

	// Read the avatar data
	avatarData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read avatar data: %w", err)
	}

	// Create a multipart.FileHeader from the avatar data
	fileName := filepath.Base(avatarURL)
	fileType := http.DetectContentType(avatarData)
	fileBuffer := bytes.NewBuffer(avatarData)

	// Create a new multipart writer
	form := multipart.NewWriter(fileBuffer)
	part, err := form.CreateFormFile("avatar", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Write avatar data into the form file
	if _, err = part.Write(avatarData); err != nil {
		return nil, fmt.Errorf("failed to write avatar data to form: %w", err)
	}

	// Close the form to finalize
	if err := form.Close(); err != nil {
		return nil, fmt.Errorf("failed to close form: %w", err)
	}

	// Emulate a *multipart.FileHeader for ActiveStorage
	header := make(map[string][]string)
	header["Content-Disposition"] = []string{`form-data; name="avatar"; filename="` + fileName + `"`}
	header["Content-Type"] = []string{fileType}
	fileHeader := &multipart.FileHeader{
		Filename: fileName,
		Header:   header,
		Size:     int64(fileBuffer.Len()),
	}

	// Attach the avatar to ActiveStorage
	attachment, err := s.ActiveStorage.Attach(&user.User, "avatar", fileHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to attach avatar: %w", err)
	}

	return attachment, nil
}

func (s *OAuthService) generateUniqueUsername(baseUsername string) string {
	username := baseUsername
	counter := 1
	for {
		var existingUser profile.User
		if s.DB.Where("username = ?", username).First(&existingUser).Error == gorm.ErrRecordNotFound {
			break
		}
		username = fmt.Sprintf("%s%d", baseUsername, counter)
		counter++
	}
	return username
}
