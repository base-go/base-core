# BaseUI - Modern JavaScript Framework for Base Applications

BaseUI is a Vite-powered JavaScript framework that provides modern tooling and components for Base applications while maintaining full compatibility with Go server-side rendering.

## Features

- 🚀 **Vite Build System** - Fast development with HMR
- 🎨 **Tailwind CSS** - Utility-first CSS framework  
- 🧩 **shadcn/ui Ready** - Copy-paste high-quality components
- 📦 **Modular Components** - Object-oriented JavaScript components
- 🔄 **SSR Compatible** - Works seamlessly with Go templates

## Quick Start

### Installation
```bash
cd ui/
npm install
```

### Development
```bash
# Terminal 1: Start Vite dev server
npm run dev

# Terminal 2: Start Go server  
cd ..
go run main.go
```

### Production Build
```bash
npm run build
```

## Adding shadcn/ui Components
```bash
npx shadcn-ui@latest add button
npx shadcn-ui@latest add dropdown-menu
```