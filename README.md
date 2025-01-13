# Base Framework

Base is a modern Go web framework designed for rapid development and maintainable code.

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
- API Key Authentication
- Rate Limiting Middleware
- Request Logging
- Security Headers
- CORS Support

### Storage & Files
- Local File Storage
- S3 Compatible Storage
- Cloudflare R2 Support
- Active Storage Pattern
- File Type Validation
- Image Processing
- Custom Storage Providers

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
- Model Relationships
- Auto-Migration
- Query Building
- Transaction Support
- Connection Management
- Type Safe Queries

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

## Installation & Usage

First, install the Base CLI:

```bash
go install github.com/base-go/cmd@latest
```

### Available Commands

```bash
# Create a new project
base new myapp

# Start development server with hot reload
base start

# Generate modules
base g post title:string content:text        # Basic module
base g post title:string author:belongsTo:User  # With relationships

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

Your API will be available at `http://localhost:8080`

### Generate Modules

```bash
# Generate a module with fields
base g post title:string content:text published:bool

# Generate with relationships
base g post title:string content:text author:belongsTo:User category:belongsTo:Category tags:hasMany:Tag

# Remove a module
base d post
```

### Configuration

Base uses environment variables for configuration. A `.env` file is automatically created with your new project:

```bash
SERVER_ADDRESS=:8080
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

Base follows the HMVC (Hierarchical Model View Controller) pattern with a centralized models directory to prevent circular imports:

```
.
├── app/
│   ├── models/            # All models in one place to prevent circular imports
│   │   ├── post.go
│   │   ├── user.go
│   │   └── comment.go
│   ├── posts/            # Module implementation
│   │   ├── controller.go # HTTP request handling
│   │   ├── service.go    # Business logic
│   │   └── module.go     # Module registration
│   └── init.go           # Module registration
├── core/                 # Framework core
├── storage/              # File storage
├── .env                  # Environment config
└── main.go              # Entry point
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
    AuthorID  uint      `json:"author_id"`
    Author    User      `json:"author" gorm:"foreignKey:AuthorID"`    // Can reference User model
    Comments  []Comment `json:"comments" gorm:"foreignKey:PostID"`    // Can reference Comment model
}

// app/posts/service.go
package posts

import "base/app/models"  // Clean import, no circular dependency

type PostService struct {
    db      *gorm.DB
    emitter *emitter.Emitter
}

func (s *PostService) Create(post *models.Post) error {
    return s.db.Create(post).Error
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
└── post/
    ├── controller.go  # RESTful endpoints
    ├── service.go     # Business logic
    ├── module.go      # Registration
    └── types.go       # Types and DTOs
```

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
    AuthorID  uint      `json:"author_id"`
    Author    User      `json:"author" gorm:"foreignKey:AuthorID"`
    Tags      []Tag     `json:"tags" gorm:"many2many:post_tags;"`
    Comments  []Comment `json:"comments" gorm:"foreignKey:PostID"`
}

// app/posts/controller.go
package posts

type PostController struct {
    service *PostService
    logger  logger.Logger
}

func (c *PostController) Routes(router *gin.RouterGroup) {
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

func NewPostModule(db *gorm.DB, router *gin.RouterGroup, log logger.Logger, emitter *emitter.Emitter) module.Module {
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