# Base Framework

Base is a modern Go web framework designed for rapid development and maintainable code.

## Quick Links

- [Detailed Documentation](docs.md)
- [CLI Tool Repository](https://github.com/base-go/cmd)
- [CLI Documentation](https://github.com/base-go/cmd/blob/main/README.md)

## Features

### Core Features
- Built-in User Authentication & Authorization
- Module System with Auto-Registration
- Database Integration with GORM
- File Storage System
- Email Service Integration
- WebSocket Support
- Event-Driven Architecture with Emitter
- Structured Logging
- Environment-Based Configuration

### Development Tools
- Code Generation via `github.com/base-go/cmd`
- Development Server with Auto-Reload
- Module-Based Architecture
- Dependency Injection
- Custom Type System
- Helper Functions

### Security Features
- JWT Token Authentication
  - Extensible JWT Claims via `Extend` function in `app/init.go`
  - Default user context with `user_id` and structured role information
  - Customizable token expiration (24h by default)
  - Secure token validation and verification
- Role-Based Access Control
  - **First User Owner System**: First registered user automatically becomes Owner
  - Hierarchical role system: Owner → Administrator → Member → Viewer
  - Automatic role assignment with secure fallback protection
  - Role information embedded in JWT tokens for authorization checks
- API Key Authentication
- Rate Limiting Middleware
- Request Logging
- Security Headers
- CORS Support

### Storage & Files
- Local File Storage
- Active Storage Pattern
- File Type Validation
  - Image Attachments (5MB limit, image extensions)
  - File Attachments (50MB limit, document extensions)
  - Generic Attachments (10MB limit, mixed extensions)
- Image Processing

### Email Features
- Multiple Provider Support:
  - SMTP
  - SendGrid
  - Postmark
  - Custom Providers
- Template Support
- Attachment Handling
- HTML/Text Email Support

### Database Features
- GORM Integration
- Model Relationships:
  - belongs_to (One-to-one with foreign key in this model)
  - has_one (One-to-one with foreign key in other model)
  - has_many (One-to-many)
  - to_many (Many-to-many with join table)
  - **Automatic Relationship Detection**: Fields ending with `_id` automatically generate relationships
- Auto-Migration
- Transaction Support
- Connection Management

### API Features
- RESTful API Support
- Request/Response Handling
- Error Management
- Pagination
- Sorting & Filtering
- API Versioning
- Swagger Documentation

### Middleware System
- Built-in Middlewares:
  - Authentication
  - API Key Validation
  - Rate Limiting
  - Request Logging
  - Custom Middleware Support

### WebSocket Features
- Real-time Communication
- Channel Management
- Message Broadcasting
- Connection Handling
- Event Subscription

### Event System
- Thread-Safe Event Emitter
- Asynchronous Event Handling
- Panic Recovery in Listeners
- Event Subscription with `On`
- Event Broadcasting with `Emit`
- Support for Any Data Type
- Event Cleanup with `Clear`

For detailed examples and usage patterns, see [docs.md](docs.md).

Common Events:
- `{module}.created`: Emitted when a record is created
- `{module}.updated`: Emitted when a record is updated
- `{module}.deleted`: Emitted when a record is deleted
- `{module}.{field}.uploaded`: Emitted when a file is uploaded
- `{module}.{field}.deleted`: Emitted when a file is deleted

Example usage:
```go
// In your service
type PostService struct {
    DB      *gorm.DB
    Emitter *emitter.Emitter
    Logger  logger.Logger
}

// Register event listeners
func (s *PostService) Init() {
    // Listen for post creation events
    s.Emitter.On("post.created", func(data any) {
        if post, ok := data.(*models.Post); ok {
            s.Logger.Info("Post created", 
                logger.Int("id", int(post.Id)),
                logger.String("title", post.Title))
        }
    })
}

// Emit events in your methods
func (s *PostService) Create(post *models.Post) error {
    if err := s.DB.Create(post).Error; err != nil {
        s.Logger.Error("Failed to create post", 
            logger.String("error", err.Error()))
        return err
    }

    // Emit event after successful creation
    s.Emitter.Emit("post.created", post)
    return nil
}
```

## Installation & Usage

Install Base CLI with a single command:

```bash
curl -fsSL https://get.base.al | bash
```

### Available Commands

```bash
# Create a new project
base new myapp

# Start development server with hot reload
base start

# Generate modules
base g post title:string content:text published:bool

# Generate with relationships and attachments
base g post \
  title:string \
  content:text \
  featured_image:image \
  gallery:attachment \
  author:belongsTo:User \
  comments:hasMany:Comment

# Generate with automatic relationship detection
base g article \
  title:string \
  content:text \
  category_id:uint \      # Automatically creates Category relationship
  author_id:uint          # Automatically creates Author relationship

# Generate with specialized attachments
base g document \
  title:string \
  file:file          # Document attachment with validation
  author:belongsTo:User

# Remove modules
base d post

# Update framework
base update   # Update framework dependencies
base upgrade  # Upgrade to latest version

# Other commands
base version  # Show version information
base feed     # Show latest updates and news
```

### Create a New Project

```bash
# Create a new project
base new myapp
cd myapp

# Start the development server with hot reload
base start
```

Your API will be available at `http://localhost:8100`

### Configuration

Base uses environment variables for configuration. A `.env` file is automatically created with your new project:

```bash
SERVER_ADDRESS=:8100
JWT_SECRET=your_jwt_secret
API_KEY=your_api_key

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp
DB_USER=postgres
DB_PASSWORD=postgres

# Storage
STORAGE_DRIVER=local  # local, s3, r2
STORAGE_PATH=storage

# Email
MAIL_DRIVER=smtp     # smtp, sendgrid, postmark
MAIL_HOST=smtp.mailtrap.io
MAIL_PORT=2525
MAIL_USERNAME=username
MAIL_PASSWORD=password
```

### Project Structure

Base follows a modular architecture with a centralized models directory:

```
.
├── app/
│   ├── models/            # All models in one place
│   │   ├── post.go       # Post model with GORM tags
│   │   ├── user.go       # User model
│   │   └── comment.go    # Comment model
│   ├── posts/            # Post module
│   │   ├── controller.go # HTTP handlers & file upload
│   │   ├── service.go    # Business logic & storage
│   │   ├── module.go     # Module registration
│   │   └── validation.go # Validation rules
│   ├── users/            # User module
│   │   ├── controller.go
│   │   ├── service.go
│   │   ├── module.go
│   │   └── validation.go # Validation rules
│   └── init.go           # Module initialization
├── core/                 # Framework core
│   ├── storage/         # File storage system
│   ├── logger/          # Structured logging
│   └── emitter/         # Event system
├── storage/              # File storage directory
├── .env                  # Environment config
└── main.go              # Entry point
```

When you generate a new module:

```bash
# Generate a post module
base g post title:string content:text

# Creates:
app/
├── models/
│   └── post.go         # Post model with GORM tags
└── posts/              # Post module
    ├── controller.go   # HTTP handlers & validation
    ├── service.go      # Business logic
    ├── module.go       # Module registration
    └── validation.go   # Validation rules
```

### Model Organization

All models are kept in the `app/models` directory to:
1. Prevent circular dependencies between modules
2. Allow modules to reference each other's models
3. Maintain a single source of truth for data structures
4. Enable proper relationship definitions

Example:
```go
// app/models/post.go
package models

type Post struct {
    types.Model
    Title     string     `json:"title" gorm:"not null"`
    Content   string     `json:"content" gorm:"type:text"`
    AuthorId  uint      `json:"author_id"`
    Author    User      `json:"author" gorm:"foreignKey:AuthorId"`    // Can reference User model
    Comments  []Comment `json:"comments" gorm:"foreignKey:PostId"`    // Can reference Comment model
}

// app/posts/service.go
package posts

import "base/app/models"  // Clean import, no circular dependency

type PostService struct {
    db      *gorm.DB
    emitter *emitter.Emitter
}

func (s *PostService) Create(post *models.Post) error {
    if err := s.db.Create(post).Error; err != nil {
        return err
    }
    s.emitter.Emit("post.created", post)
    return nil
}
```

This structure ensures clean dependencies while maintaining modularity.

### Module Structure

Each module in Base is self-contained and follows HMVC principles:

1. **Controller Layer** (`controller.go`)
   - Handles HTTP requests and responses
   - Input validation
   - Route definitions
   - Response formatting

2. **Service Layer** (`service.go`)
   - Contains business logic
   - Database operations
   - External service integration
   - Data transformation

3. **Module Registration** (`module.go`)
   - Dependency injection
   - Route group configuration
   - Middleware setup
   - Module initialization

4. **Types** (`types.go`)
   - Request/Response structs
   - Module-specific types
   - Data Transfer Objects (DTOs)

### Module Generation

When you generate a new module using `base g`, it creates this HMVC structure:

```bash
# Generate a new post module
base g post title:string content:text

# Creates:
app/
├── models/
│   └── post.go        # Model with automatic relationships
└── posts/
    ├── controller.go  # RESTful endpoints
    ├── service.go     # Business logic
    ├── module.go      # Registration
    └── validator.go   # Input validation
```

#### Automatic Relationship Detection

Base automatically detects and creates relationships when field names end with `_id`:

```bash
# This command:
base g article title:string content:text category_id:uint author_id:uint

# Automatically generates:
type Article struct {
    Id         uint     `json:"id" gorm:"primarykey"`
    Title      string   `json:"title"`
    Content    string   `json:"content"`
    CategoryId uint     `json:"category_id"`
    Category   Category `json:"category,omitempty" gorm:"foreignKey:CategoryId"`
    AuthorId   uint     `json:"author_id"`  
    Author     Author   `json:"author,omitempty" gorm:"foreignKey:AuthorId"`
}
```

This eliminates the need to manually specify relationships - just use the `_id` suffix convention!

The module is automatically registered in `app/init.go` and integrated with the dependency injection system.

### Module Communication

Modules can communicate through:
1. Direct Service Calls
2. Event Emitter
3. WebSocket Channels
4. Shared Models

Example of module interaction:
```go
// Post service using user service
type PostService struct {
    userService *user.Service    // Direct service injection
    emitter     *emitter.Emitter // Event-based communication
}
```

### HMVC Example

Here's a complete example of a Post module following HMVC principles:

```go
// app/models/post.go
package models

type Post struct {
    types.Model
    Title     string     `json:"title" gorm:"not null"`
    Content   string     `json:"content" gorm:"type:text"`
    Published bool       `json:"published" gorm:"default:false"`
    AuthorId  uint      `json:"author_id"`
    Author    User      `json:"author" gorm:"foreignKey:AuthorId"`
    Tags      []Tag     `json:"tags" gorm:"many2many:post_tags;"`
    Comments  []Comment `json:"comments" gorm:"foreignKey:PostId"`
}

// app/posts/controller.go
package posts

type PostController struct {
    service *PostService
    logger  logger.Logger
}

func (c *PostController) Routes(router *router.RouterGroup) {
    router.GET("", c.List)
    router.GET("/:id", c.Get)
    router.POST("", c.Create)
    router.PUT("/:id", c.Update)
    router.DELETE("/:id", c.Delete)
}

// app/posts/service.go
package posts

type PostService struct {
    db          *gorm.DB
    userService *user.Service
    emitter     *emitter.Emitter
}

func (s *PostService) Create(post *models.Post) error {
    if err := s.db.Create(post).Error; err != nil {
        return err
    }
    s.emitter.Emit("post.created", post)
    return nil
}

// app/posts/module.go
package posts

type PostModule struct {
    controller *PostController
    service    *PostService
}

func NewPostModule(db *gorm.DB, router *router.RouterGroup, log logger.Logger, emitter *emitter.Emitter) module.Module {
    service := &PostService{
        db:      db,
        emitter: emitter,
    }
    
    controller := &PostController{
        service: service,
        logger:  log,
    }

    return &PostModule{
        controller: controller,
        service:    service,
    }
}
```

This structure provides:
1. Clear separation of concerns
2. Dependency injection
3. Event-driven capabilities
4. Clean routing
5. Type safety
6. Automatic model relationships

## Documentation

For detailed documentation, visit [docs.base-go.dev](https://docs.base-go.dev)

## License

MIT License - see LICENSE for more details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

If you encounter any issues or have questions, please file an issue on the GitHub repository.