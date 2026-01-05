# 🔥 ReboloLang

A modern, fast, and elegant web framework for Go

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## ⚡ Why ReboloLang?

- **🚀 Blazing Fast** - Bun.js powered asset pipeline
- **🔥 Hot Reload** - Real-time development without manual restarts
- **📦 Standard Library** - Built on Go's standard library (`database/sql`, `html/template`)
- **✨ Type-Safe** - Full type safety with intelligent helpers
- **🧪 Testing First** - Comprehensive testing utilities included
- **🎯 Convention over Configuration** - Productive defaults, flexible when needed

## 🚀 Quick Start

### Install

```bash
go install github.com/Palaciodiego008/rebololang/cmd/rebolo@latest
```

### Create New App

```bash
# Basic app
rebolo new myblog

# With React frontend
rebolo new myblog --frontend react

# With Svelte frontend
rebolo new myblog --frontend svelte

# With Vue frontend
rebolo new myblog --frontend vue

cd myblog
```

### Generate Resource

```bash
rebolo generate resource Post title:string content:text published:bool
```

### Run

```bash
rebolo dev
```

Visit: `http://localhost:3000` 🎉

## ✨ Features

| Feature | Status |
|---------|--------|
| 🔥 Hot Reload | ✅ |
| 📧 Sessions & Flash Messages | ✅ |
| 🎯 Context Helpers | ✅ |
| ✅ Form Validation | ✅ |
| ❌ Error Handlers | ✅ |
| 🔧 Middleware Stack | ✅ |
| 🧪 Testing Helpers | ✅ |
| ⚡ Asset Pipeline (Bun.js) | ✅ |
| 🗄️ SQLite/PostgreSQL | ✅ |
| ⚛️ React/Svelte/Vue Support | ✅ |

## 📖 Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Commands Reference](docs/COMMANDS.md)
- [Frontend Frameworks](docs/FRONTEND.md) - React, Svelte, Vue support
- [Examples](examples/)

## 🎯 Example

### Simple Controller with Context

```go
func (c *PostsController) Create(ctx *rebolo.Context) error {
    var post Post
    
    // Bind and validate in one step
    if err := ctx.BindAndValidate(&post); err != nil {
        flash, _ := ctx.Flash()
        flash.Error("Validation failed")
        return ctx.Redirect("/posts/new", 303)
    }
    
    // Save post
    if err := c.repo.Save(&post); err != nil {
        return err
    }
    
    // Flash message + redirect
    flash, _ := ctx.Flash()
    flash.Success("Post created!")
    ctx.SaveSession()
    
    return ctx.Redirect("/posts", 303)
}
```

### Form Validation

```go
type CreatePostForm struct {
    Title   string `form:"title" validate:"required,min=3,max=100"`
    Content string `form:"content" validate:"required,min=10"`
}

// Automatic validation with Spanish error messages
func (c *PostsController) Create(ctx *rebolo.Context) error {
    var form CreatePostForm
    if err := ctx.BindAndValidate(&form); err != nil {
        // Handle validation errors
    }
}
```

### Testing

```go
func TestPostsController(t *testing.T) {
    app := rebolo.NewTestApp()
    
    resp := app.POST("/posts").
        WithForm(map[string]string{
            "title": "Test Post",
            "content": "Test content",
        }).
        Do()
    
    assert.True(t, resp.IsRedirect())
    assert.True(t, resp.Contains("created"))
}
```

## 🏗️ Project Structure

```
myapp/
├── app/
│   ├── controllers/      # HTTP handlers
│   ├── models/           # Database models
│   ├── middleware/       # Custom middleware
│   └── services/         # Business logic
├── config/
│   └── config.yml        # Configuration
├── views/
│   ├── layouts/          # Layout templates
│   ├── shared/           # Partials
│   └── errors/           # Error pages
├── assets/
│   ├── css/              # Stylesheets
│   ├── js/               # JavaScript
│   └── images/           # Images
├── public/               # Compiled assets
├── db/
│   ├── migrations/       # Database migrations
│   └── seeds/            # Seed data
└── main.go               # Entry point
```

## 🛠️ Commands

```bash
# Create new app
rebolo new myapp

# Create app with React frontend
rebolo new myapp --frontend react

# Create app with Svelte frontend
rebolo new myapp --frontend svelte

# Create app with Vue frontend
rebolo new myapp --frontend vue

# Generate resource (CRUD)
rebolo generate resource Post title:string content:text

# Run with hot reload
rebolo dev

# Run tests
go test ./...

# Build for production
go build -o myapp .
```

## 📦 Requirements

- Go 1.21+
- Bun.js (auto-installed)
- SQLite or PostgreSQL

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📝 License

MIT License - see [LICENSE](LICENSE) file

---

**Built with ❤️ in Barranquilla, Colombia 🇨🇴**
