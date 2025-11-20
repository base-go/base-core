# Base Framework Documentation

This document provides detailed examples and usage patterns for the Base framework's features.

## Table of Contents
- [Event System](#event-system)
- [File Storage](#file-storage)
- [Logging](#logging)
- [Database](#database)
- [Authentication](#authentication)
- [Email System](#email-system)

## Event System

The event system in Base is built around a thread-safe emitter that supports asynchronous event handling.

### Basic Usage

```go
type UserService struct {
    DB      *gorm.DB
    Emitter *emitter.Emitter
    Logger  logger.Logger
}

// Initialize event listeners
func (s *UserService) Init() {
    // User creation events
    s.Emitter.On("user.created", func(data any) {
        if user, ok := data.(*models.User); ok {
            s.Logger.Info("New user registered",
                logger.Int("id", int(user.Id)),
                logger.String("email", user.Email))
        }
    })

    // User update events
    s.Emitter.On("user.updated", func(data any) {
        if user, ok := data.(*models.User); ok {
            s.Logger.Info("User profile updated",
                logger.Int("id", int(user.Id)))
        }
    })
}

// Emit events in methods
func (s *UserService) Create(user *models.User) error {
    if err := s.DB.Create(user).Error; err != nil {
        s.Logger.Error("Failed to create user",
            logger.String("error", err.Error()))
        return err
    }
    s.Emitter.Emit("user.created", user)
    return nil
}
```

### File Upload Events

```go
type PostService struct {
    DB      *gorm.DB
    Emitter *emitter.Emitter
    Logger  logger.Logger
    Storage storage.Storage
}

func (s *PostService) Init() {
    // File upload events
    s.Emitter.On("post.featured_image.uploaded", func(data any) {
        if post, ok := data.(*models.Post); ok {
            s.Logger.Info("Featured image uploaded",
                logger.Int("post_id", int(post.Id)),
                logger.String("filename", post.FeaturedImage))
        }
    })

    // File deletion events
    s.Emitter.On("post.featured_image.deleted", func(data any) {
        if post, ok := data.(*models.Post); ok {
            s.Logger.Info("Featured image deleted",
                logger.Int("post_id", int(post.Id)))
        }
    })
}

func (s *PostService) UploadFeaturedImage(post *models.Post, file *multipart.FileHeader) error {
    filename, err := s.Storage.Store(file)
    if err != nil {
        return err
    }
    
    post.FeaturedImage = filename
    if err := s.DB.Save(post).Error; err != nil {
        s.Storage.Delete(filename) // cleanup on error
        return err
    }
    
    s.Emitter.Emit("post.featured_image.uploaded", post)
    return nil
}
```

### Multiple Listeners

```go
type NotificationService struct {
    Emitter *emitter.Emitter
    Logger  logger.Logger
    Email   email.Sender
}

func (s *NotificationService) Init() {
    // Email notification on post creation
    s.Emitter.On("post.created", func(data any) {
        if post, ok := data.(*models.Post); ok {
            msg := email.Message{
                To:      []string{post.Author.Email},
                Subject: "Post Created",
                Body:    "Your post has been published",
                IsHTML:  false,
            }
            if err := s.Email.Send(msg); err != nil {
                s.Logger.Error("Failed to send email notification",
                    logger.String("error", err.Error()))
            }
        }
    })

    // Log post creation
    s.Emitter.On("post.created", func(data any) {
        if post, ok := data.(*models.Post); ok {
            s.Logger.Info("Post published",
                logger.Int("id", int(post.Id)),
                logger.String("title", post.Title),
                logger.Int("author_id", int(post.AuthorId)))
        }
    })
}
```

### Error Handling

The emitter includes built-in panic recovery:

```go
func (s *PostService) Init() {
    s.Emitter.On("post.created", func(data any) {
        // Even if this panics, other listeners will still execute
        panic("something went wrong")
    })

    s.Emitter.On("post.created", func(data any) {
        // This will still run
        if post, ok := data.(*models.Post); ok {
            s.Logger.Info("Post created", logger.Int("id", int(post.Id)))
        }
    })
}
```

### Event Cleanup

```go
func (s *PostService) Shutdown() {
    // Clear all event listeners
    s.Emitter.Clear()
}
```

## File Storage

Examples for file storage coming soon...

## Logging

Examples for logging coming soon...

## Database

Examples for database operations coming soon...

## Authentication

Examples for authentication coming soon...

## Email System

Base provides a flexible email system that supports multiple providers through a unified interface.

### Supported Providers

- **SMTP**: Standard email sending using SMTP servers
- **SendGrid**: Email service by Twilio
- **Postmark**: Transactional email service

### Configuration

Configure your email provider in `.env`:

```env
# Common Settings
EMAIL_PROVIDER=smtp  # smtp, sendgrid, postmark
EMAIL_FROM_ADDRESS=noreply@example.com

# SMTP Settings
SMTP_HOST=smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USERNAME=your_username
SMTP_PASSWORD=your_password

# SendGrid Settings
SENDGRID_API_KEY=your_sendgrid_api_key

# Postmark Settings
POSTMARK_SERVER_TOKEN=your_server_token
POSTMARK_ACCOUNT_TOKEN=your_account_token
```

### Basic Usage

```go
// Initialize email system
if err := email.Initialize(cfg); err != nil {
    log.Fatal("Failed to initialize email system:", err)
}

// Send a simple text email
msg := email.Message{
    To:      []string{"user@example.com"},
    Subject: "Welcome to Base",
    Body:    "Thank you for joining us!",
    IsHTML:  false,
}

if err := email.Send(msg); err != nil {
    log.Error("Failed to send email:", err)
}
```

### HTML Emails

```go
// Send an HTML email
msg := email.Message{
    To:      []string{"user@example.com"},
    Subject: "Welcome to Base",
    Body:    `
        <h1>Welcome to Base!</h1>
        <p>Thank you for joining us. Here's what you can do next:</p>
        <ul>
            <li>Complete your profile</li>
            <li>Explore our features</li>
            <li>Read the documentation</li>
        </ul>
    `,
    IsHTML:  true,
}

if err := email.Send(msg); err != nil {
    log.Error("Failed to send HTML email:", err)
}
```

### Service Integration

```go
type UserService struct {
    DB      *gorm.DB
    Email   email.Sender
    Logger  logger.Logger
}

func (s *UserService) SendWelcomeEmail(user *models.User) error {
    msg := email.Message{
        To:      []string{user.Email},
        Subject: "Welcome to " + config.AppName,
        Body:    fmt.Sprintf("Welcome %s! Thank you for joining us.", user.Name),
        IsHTML:  false,
    }

    if err := s.Email.Send(msg); err != nil {
        s.Logger.Error("Failed to send welcome email",
            logger.String("user_email", user.Email),
            logger.String("error", err.Error()))
        return err
    }

    s.Logger.Info("Welcome email sent",
        logger.String("user_email", user.Email))
    return nil
}
```

### Password Reset Example

```go
func (s *UserService) SendPasswordResetEmail(user *models.User, token string) error {
    resetURL := fmt.Sprintf("%s/reset-password?token=%s", config.AppURL, token)
    
    msg := email.Message{
        To:      []string{user.Email},
        Subject: "Password Reset Request",
        Body:    fmt.Sprintf(`
            <h2>Password Reset Request</h2>
            <p>Hello %s,</p>
            <p>We received a request to reset your password. Click the link below to proceed:</p>
            <p><a href="%s">Reset Password</a></p>
            <p>If you didn't request this, please ignore this email.</p>
            <p>The link will expire in 1 hour.</p>
        `, user.Name, resetURL),
        IsHTML:  true,
    }

    if err := s.Email.Send(msg); err != nil {
        s.Logger.Error("Failed to send password reset email",
            logger.String("user_email", user.Email),
            logger.String("error", err.Error()))
        return err
    }

    s.Logger.Info("Password reset email sent",
        logger.String("user_email", user.Email))
    return nil
}
```

### Provider-Specific Features

Each email provider has its own strengths:

- **SMTP**:
  - Standard protocol support
  - Works with any SMTP server
  - Full control over email headers

- **SendGrid**:
  - High deliverability
  - Detailed analytics
  - Template support
  - Webhook integration

- **Postmark**:
  - Specialized for transactional email
  - High delivery speed
  - Detailed bounce handling
  - Template support

Choose the provider that best fits your needs. You can easily switch providers by updating your configuration without changing your code.
