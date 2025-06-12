# BaseUI Framework

Modern JavaScript framework that seamlessly integrates with Go server-side rendering. Build powerful web applications with Vite, Tailwind CSS, and shadcn/ui components.

## Quick Start

### Prerequisites
- **Bun** (recommended) or **Node.js 18+**
- **Go 1.19+**

> 💡 **Bun is recommended** for faster package installation and better performance. Install from [bun.sh](https://bun.sh)

### Setup
```bash
# Clone and setup
git clone <your-repo>
cd base
./bin/setup

# Start development
./bin/dev
```

Visit http://localhost:8080 to see your application!

## Development Commands

### Main Commands
```bash
./bin/dev        # Start development servers (recommended)
./bin/build      # Build for production
./bin/setup      # Set up development environment
```

### Using Make (alternative)
```bash
make dev         # Start development servers
make build       # Build for production
make help        # Show all available commands
```

## Development Workflow

### 1. Frontend Development
The UI is built with modern tools:
- **Vite** - Fast build tool with HMR
- **Tailwind CSS** - Utility-first styling
- **shadcn/ui** - High-quality components

```bash
# Add new shadcn/ui components
make add-component COMPONENT=button

# Using Bun directly (faster)
cd ui && bunx shadcn-ui@latest add button

# Using npm
cd ui && npx shadcn-ui@latest add button
```

### 2. Component Integration
Components work seamlessly between Go templates and JavaScript:

**Go Template:**
```html
<{ Dropdown buttonText="Options" items=myItems }>
```

## Deployment

### Production Build
```bash
./bin/build
```

This creates optimized assets ready for deployment.