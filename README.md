# ReboloLang Framework 🚀🇨🇴

A modern Go web framework inspired by **Rebolo**, Barranquilla, Colombia. Built with convention over configuration, hot reload, and Bun.js for lightning-fast asset compilation.

## Features

- 🔥 **Hot Reload** - Both Go server and frontend assets
- ⚡ **Bun.js Integration** - Ultra-fast asset compilation
- 🛠️ **Code Generators** - Rails-like resource generation
- 🗃️ **Built-in ORM** - Powered by Bun ORM with PostgreSQL
- 🎨 **HTML Templates** - Server-side rendering with layouts
- 📱 **API Support** - JSON APIs out of the box
- 🔧 **CLI Tools** - Complete development workflow
- 🛡️ **Middleware** - Logging, recovery, and more
- 🇨🇴 **Colombian Pride** - Named after Rebolo neighborhood

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/Palaciodiego008/rebololang/main/install.sh | bash
```

Or manually:
```bash
go install github.com/Palaciodiego008/rebololang/cmd/rebolo@latest
```

## Quick Start

### 1. Create a new app
```bash
rebolo new blog
cd blog
```

### 2. Generate a resource
```bash
rebolo generate resource posts title:string content:text published:bool author:string
```

### 3. Configure database (optional)
Edit `config.yml`:
```yaml
database:
  url: postgres://localhost/blog_development
```

### 4. Start development server
```bash
rebolo dev
```

Your app runs at `http://localhost:3000` with hot reload! 🎉

## CLI Commands

### App Management
```bash
rebolo new myapp              # Create new application
rebolo dev                    # Start development server with hot reload
```

### Code Generation
```bash
rebolo generate resource users name:string email:string age:int
# or shorthand:
rebolo g resource posts title:string content:text published:bool
```

### Database Operations
```bash
rebolo db migrate             # Run database migrations
```

## Generated Structure

```
blog/
├── main.go                 # Application entry point
├── config.yml              # Configuration
├── package.json            # Bun.js dependencies
├── controllers/            # HTTP controllers
│   └── posts_controller.go
├── models/                 # Database models
│   └── posts.go
├── views/                  # HTML templates
│   ├── layouts/
│   │   └── application.html
│   ├── home/
│   │   └── index.html
│   └── posts/
│       ├── index.html
│       ├── show.html
│       ├── new.html
│       └── edit.html
├── public/                 # Compiled assets
├── src/                    # Frontend source
│   └── index.js
└── db/
    └── migrations/         # Database migrations
```

## Resource Generation

When you run:
```bash
rebolo g resource posts title:string content:text published:bool
```

ReboloLang generates:

### 1. Model (`models/posts.go`)
```go
type Posts struct {
    bun.BaseModel `bun:"table:posts"`
    
    ID        int64     `bun:",pk,autoincrement"`
    Title     string    `bun:"title"`
    Content   string    `bun:"content"`
    Published bool      `bun:"published"`
    CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
    UpdatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
}
```

### 2. Controller (`controllers/posts_controller.go`)
Complete CRUD controller with:
- Index, Show, New, Create, Edit, Update, Delete actions
- Form parsing and validation
- Database operations (ready to uncomment)
- Proper error handling

### 3. Views (`views/posts/`)
- **index.html** - List all posts with edit/delete buttons
- **show.html** - Display single post
- **new.html** - Create form with proper styling
- **edit.html** - Edit form with pre-filled values

### 4. Migration (`db/migrations/xxx_create_posts.sql`)
```sql
CREATE TABLE posts (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255),
    content TEXT,
    published BOOLEAN,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Configuration

Edit `config.yml`:

```yaml
app:
  name: MyApp
  env: development

server:
  port: 3000
  host: localhost

database:
  url: postgres://localhost/myapp_development

assets:
  hot_reload: true
```

Override with environment variables:
- `PORT` - Server port
- `HOST` - Server host
- `REBOLOLANG_ENV` - Environment (development/production)

## Controllers & Routes

```go
func main() {
    app := rebololang.New()
    
    // Simple routes
    app.GET("/", HomeHandler)
    app.POST("/api/posts", CreatePostAPI)
    
    // RESTful resources (generates 7 routes)
    app.Resource("/posts", &controllers.PostsController{})
    
    // Static files
    app.Router.PathPrefix("/public/").Handler(
        http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))
    
    app.Start()
}
```

## Database Integration

ReboloLang uses Bun ORM with PostgreSQL:

```go
// In your controller
func (c *PostsController) Index(w http.ResponseWriter, r *http.Request) {
    var posts []models.Posts
    
    err := app.DB.NewSelect().Model(&posts).Scan(r.Context())
    if err != nil {
        rebololang.JSONError(w, "Failed to fetch posts", 500)
        return
    }
    
    rebololang.Render(w, "posts/index.html", map[string]interface{}{
        "Posts": posts,
    })
}
```

## Frontend Assets

ReboloLang uses Bun.js for ultra-fast asset compilation:

```javascript
// src/index.js
console.log('🚀 Blog loaded with ReboloLang!');

// Hot reload is automatic in development
if (process.env.NODE_ENV === 'development') {
  const eventSource = new EventSource('/dev/reload');
  eventSource.onmessage = () => location.reload();
}
```

Assets are:
- Compiled with Bun.js in development
- Watched for changes and hot-reloaded
- Embedded in Go binary for production

## Development Features

### Hot Reload
- **Go files** - Server automatically restarts
- **Frontend assets** - Bun.js rebuilds and browser refreshes
- **Templates** - Automatically reloaded

### Middleware
Built-in middleware:
- **Logging** - Request logging
- **Recovery** - Panic recovery
- **Static files** - Serve public assets

### Error Handling
```go
// JSON errors
rebololang.JSONError(w, "Not found", 404)

// Template rendering
rebololang.Render(w, "posts/show.html", data)

// JSON responses
rebololang.JSON(w, map[string]interface{}{
    "posts": posts,
    "total": len(posts),
})
```

## Field Types

When generating resources, use these field types:

| Type | Go Type | SQL Type | HTML Input |
|------|---------|----------|------------|
| `string` | `string` | `VARCHAR(255)` | `text` |
| `text` | `string` | `TEXT` | `textarea` |
| `int` | `int64` | `BIGINT` | `number` |
| `bool` | `bool` | `BOOLEAN` | `checkbox` |
| `float` | `float64` | `DECIMAL` | `number` |
| `time` | `time.Time` | `TIMESTAMP` | `datetime-local` |

## Production Deployment

```bash
# Build for production
rebolo build  # (coming soon)

# Single binary with embedded assets
./myapp
```

## Why ReboloLang?

- **🏠 Familiar** - Rails/Buffalo conventions in Go
- **⚡ Fast** - Bun.js for assets, Go for backend  
- **🎯 Simple** - Convention over configuration
- **📦 Complete** - CLI, ORM, templates, assets included
- **🔥 Modern** - Hot reload, embedded assets, single binary
- **🇨🇴 Colombian** - Proudly named after Rebolo, Barranquilla

## Roadmap

- [x] CLI tool with app generation
- [x] Resource generators (models, controllers, views, migrations)
- [x] Hot reload for Go and assets
- [x] Database integration with Bun ORM
- [x] HTML template rendering
- [x] JSON API support
- [x] Middleware system
- [ ] Authentication middleware
- [ ] WebSocket support
- [ ] Background jobs
- [ ] Production build command
- [ ] Docker integration

## Contributing

Built with ❤️ in honor of **Rebolo**, Barranquilla, Colombia 🇨🇴

Created by [@Palaciodiego008](https://github.com/Palaciodiego008)

## License

MIT License

---

**¡Vamos Rebolo!** 🚀🇨🇴
# Rebolo
