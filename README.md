# ReboloLang Framework 🚀🇨🇴

A modern Go web framework inspired by **Rebolo**, Barranquilla, Colombia. Built with convention over configuration, hot reload, and Bun.js for lightning-fast asset compilation.

## Features

- 🔥 **Hot Reload** - Both Go server and frontend assets
- ⚡ **Bun.js Integration** - Ultra-fast asset compilation with Bun.js toolkit
- 🛠️ **Code Generators** - Rails-like resource generation
- 🗃️ **Multi-Database Support** - PostgreSQL, SQLite, and MySQL (standard database/sql)
- 🎨 **HTML Templates** - Server-side rendering with layouts
- 📱 **API Support** - JSON APIs out of the box
- 🔧 **CLI Tools** - Complete development workflow
- 🛡️ **Middleware** - Logging, recovery, and more
- 🏗️ **Clean Architecture** - Hexagonal architecture (Ports & Adapters)
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
  driver: "sqlite"  # or "postgres", "mysql"
  url: "file:./blog.db?cache=shared&mode=rwc"
  debug: true
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
  driver: postgres  # or sqlite, mysql
  url: postgres://localhost/myapp_development
  debug: false  # Enable SQL query logging

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

Supports **PostgreSQL**, **SQLite**, and **MySQL** using Go's standard `database/sql`. No ORM imposed - use what you prefer!

```yaml
database:
  driver: "sqlite"  # or postgres, mysql
  url: "file:./app.db?cache=shared&mode=rwc"
  debug: true
```

```go
func HomeHandler(w http.ResponseWriter, r *http.Request) {
    db := app.DB()  // Returns *sql.DB
    
    rows, err := db.QueryContext(r.Context(), 
        "SELECT id, title FROM posts")
    // ... use standard database/sql or wrap with your favorite ORM
}
```

## Frontend Assets

ReboloLang uses **Bun.js** (the JavaScript toolkit) for ultra-fast asset compilation:

```javascript
// src/index.js
console.log('🚀 Blog loaded with ReboloLang!');

// Hot reload is automatic in development
if (process.env.NODE_ENV === 'development') {
  const eventSource = new EventSource('/dev/reload');
  eventSource.onmessage = () => location.reload();
}
```

> **What is Bun.js?** Bun is a fast, all-in-one JavaScript toolkit that includes a bundler, runtime, and package manager. It's significantly faster than traditional tools like Webpack or Vite.

Assets are:
- Compiled with **Bun.js** in development
- Watched for changes and hot-reloaded automatically
- Minified and optimized for production
- Can be embedded in Go binary

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
- **⚡ Fast** - Go backend + Bun.js for lightning-fast asset compilation
- **🎯 Simple** - Convention over configuration
- **📦 Complete** - CLI, database support, templates, hot reload included
- **🔥 Modern** - Hot reload with Bun.js, clean architecture, single binary
- **🔄 Flexible** - Swap databases easily, use any ORM or none at all
- **🇨🇴 Colombian** - Proudly named after Rebolo, Barranquilla

## Example

Check `examples/sqlite-todo/` for a complete REST API example.

## Roadmap

- [x] CLI tool with app generation
- [x] Resource generators (models, controllers, views, migrations)
- [x] Hot reload for Go and assets
- [x] Multi-database support (PostgreSQL, SQLite, MySQL)
- [x] Hexagonal architecture (Ports & Adapters)
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
