# ReboloLang Clean Architecture 🏗️

## Architecture Overview

ReboloLang follows **Hexagonal Architecture** (Ports & Adapters) principles for clean, maintainable, and testable code.

```
pkg/rebolo/
├── core/           # Business Logic (Domain)
│   └── app.go      # Core application logic
├── ports/          # Interfaces (Contracts)
│   └── ports.go    # Port definitions
├── adapters/       # Infrastructure (External Dependencies)
│   ├── config.go   # YAML configuration
│   ├── router.go   # HTTP routing (Mux)
│   ├── database.go # Database factory (standard database/sql)
│   └── renderer.go # Template/JSON rendering
└── rebolo.go       # Application Facade
```

## Layers Explained

### 🎯 Core (Domain Layer)
- **Pure business logic**
- **No external dependencies**
- **Defines interfaces for external services**
- **Contains the main App struct and business rules**

```go
// Core defines what the app needs, not how it's implemented
type Database interface {
    Connect(ctx context.Context) error
    Close() error
    Migrate(ctx context.Context) error
}
```

### 🔌 Ports (Interface Layer)
- **Contracts between core and adapters**
- **Define data structures**
- **No implementation details**

```go
type ConfigPort interface {
    Load() (ConfigData, error)
    GetEnv(key, defaultValue string) string
}
```

### 🔧 Adapters (Infrastructure Layer)
- **Implement port interfaces**
- **Handle external dependencies**
- **Can be easily swapped**

```go
// YAML adapter implements ConfigPort
type YAMLConfig struct{}
func (c *YAMLConfig) Load() (ConfigData, error) { ... }

// Could easily add JSON adapter
type JSONConfig struct{}
func (c *JSONConfig) Load() (ConfigData, error) { ... }
```

### 🎭 Facade (Application Layer)
- **Simple API for users**
- **Wires everything together**
- **Provides convenience methods**

```go
func New() *Application {
    // Wire dependencies
    config := adapters.NewYAMLConfig()
    router := adapters.NewMuxRouter()
    database := adapters.NewBunDatabase()
    
    return &Application{...}
}
```

## Benefits

### ✅ **Testability**
```go
// Easy to mock interfaces for testing
type MockDatabase struct{}
func (m *MockDatabase) Connect(ctx context.Context) error { return nil }

// Test with mock
app := core.NewApp(config, router, mockDB, renderer)
```

### ✅ **Flexibility**
```go
// Swap Mux for Gin easily
ginRouter := adapters.NewGinRouter()  // New adapter
app := core.NewApp(config, ginRouter, database, renderer)

// Swap PostgreSQL for SQLite
factory := adapters.NewDatabaseFactory()
sqliteDB, _ := factory.CreateDatabase("sqlite")
app := core.NewApp(config, router, sqliteDB, renderer)
```

### ✅ **Maintainability**
- Clear separation of concerns
- Dependencies point inward
- Easy to understand and modify

### ✅ **Extensibility**
```go
// Add new features without changing core
type CacheAdapter struct{}
func (c *CacheAdapter) Get(key string) interface{} { ... }

// Core doesn't need to change
```

## Usage Examples

### Basic App
```go
package main

import "github.com/Palaciodiego008/rebololang/pkg/rebolo"

func main() {
    app := rebolo.New()  // Facade handles wiring
    
    app.GET("/", HomeHandler)
    app.Start()
}
```

### Custom Configuration
```go
// Could easily add custom config adapter
type EnvConfig struct{}
func (e *EnvConfig) Load() (ports.ConfigData, error) {
    // Load from environment only
}

// Wire custom adapter
app := rebolo.NewWithConfig(customConfig)
```

### Testing
```go
func TestApp(t *testing.T) {
    mockDB := &MockDatabase{}
    mockRenderer := &MockRenderer{}
    
    app := core.NewApp(config, router, mockDB, mockRenderer)
    
    // Test business logic without external dependencies
}
```

## Directory Structure

```
your-app/
├── main.go                 # Application entry point
├── controllers/            # HTTP controllers
├── models/                 # Domain models
├── views/                  # Templates
└── config.yml             # Configuration

# Framework structure
pkg/rebolo/
├── core/                   # 🎯 Business Logic
│   └── app.go
├── ports/                  # 🔌 Interfaces
│   └── ports.go
├── adapters/               # 🔧 Infrastructure
│   ├── config.go
│   ├── router.go
│   ├── database.go
│   └── renderer.go
└── rebolo.go              # 🎭 Facade
```

## Design Principles

1. **Dependency Inversion** - Core depends on interfaces, not implementations
2. **Single Responsibility** - Each adapter has one job
3. **Open/Closed** - Open for extension, closed for modification
4. **Interface Segregation** - Small, focused interfaces
5. **Dependency Injection** - Dependencies injected, not created

## Why This Architecture?

### 🚀 **Performance**
- No unnecessary abstractions in hot paths
- Direct access to underlying implementations when needed

### 🧪 **Testing**
- Mock any external dependency
- Test business logic in isolation
- Fast unit tests

### 🔄 **Evolution**
- Swap databases (PostgreSQL → SQLite → MySQL)
- Change routers (Mux → Gin → Chi)
- Use any ORM or none at all
- Add caching, logging, monitoring
- All without changing core business logic

### 🎯 **Focus**
- Developers focus on business logic
- Infrastructure concerns are isolated
- Clear boundaries between layers

## Multi-Database Support

ReboloLang includes adapters for PostgreSQL, SQLite, and MySQL:

```go
// Database factory creates the right adapter
factory := adapters.NewDatabaseFactory()
db, _ := factory.CreateDatabase("sqlite")  // or "postgres", "mysql"
db.ConnectWithDSN("file:./app.db", true)

// All adapters return *sql.DB - use any ORM you want!
sqlDB := db.DB()  // Standard database/sql
```

This architecture makes ReboloLang both powerful for beginners and flexible for advanced use cases! 🇨🇴🚀
