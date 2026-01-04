# 📁 ReboloLang - Package Structure

## Overview

ReboloLang follows a modular package structure, similar to other Go frameworks but with clear separation of concerns.

## Package Organization

```
pkg/rebolo/
├── adapters/          # External adapters (DB, Router, Renderer)
│   ├── database.go
│   ├── renderer.go
│   └── router.go
├── context/           # Request context helpers
│   └── context.go
├── core/              # Core business logic
│   ├── app.go
│   └── controller.go
├── errors/            # Error handling
│   └── errors.go
├── middleware/        # Middleware system
│   ├── middleware_stack.go
│   ├── middleware_helpers.go
│   └── hotreload_middleware.go
├── ports/             # Interfaces (Hexagonal Architecture)
│   ├── config.go
│   ├── database.go
│   ├── renderer.go
│   └── router.go
├── session/           # Session management
│   ├── session.go
│   ├── flash.go
│   └── helpers.go
├── testing/           # Testing utilities
│   └── testing.go
├── validation/        # Form validation & binding
│   ├── validation.go
│   └── binding.go
├── watcher/           # Hot reload file watcher
│   └── watcher.go
└── rebolo.go          # Main facade (Application)
```

## Package Responsibilities

### `adapters/`
External implementations of ports (interfaces). Adapters can be swapped without changing core logic.

- **database.go** - Database adapters (SQLite, PostgreSQL)
- **renderer.go** - HTML template renderer
- **router.go** - HTTP router (Gorilla Mux)

### `context/`
Request context with convenient helpers for controllers.

- **context.go** - Context struct with Session(), Flash(), Param(), JSON(), Render(), etc.

### `core/`
Pure business logic, independent of external dependencies.

- **app.go** - Core application logic
- **controller.go** - Controller interface

### `errors/`
Error handling and custom error pages.

- **errors.go** - Error handlers, 404/500 pages

### `middleware/`
HTTP middleware system with skip patterns.

- **middleware_stack.go** - Middleware stack with ordering
- **middleware_helpers.go** - Common middleware (CORS, Auth, etc.)
- **hotreload_middleware.go** - Hot reload script injection

### `ports/`
Interfaces (contracts) for hexagonal architecture.

- **config.go** - Configuration interface
- **database.go** - Database interface
- **renderer.go** - Renderer interface
- **router.go** - Router interface

### `session/`
Session management and flash messages.

- **session.go** - Session store and operations
- **flash.go** - Flash message helpers
- **helpers.go** - Convenience functions

### `testing/`
Testing utilities for easy test writing.

- **testing.go** - TestApp, fluent API for HTTP testing

### `validation/`
Form binding and validation.

- **validation.go** - Struct validation with go-playground/validator
- **binding.go** - Form/JSON binding to structs

### `watcher/`
File system watcher for hot reload.

- **watcher.go** - File watcher with fsnotify

### `rebolo.go`
Main application facade that ties everything together.

## Design Principles

### 1. **Hexagonal Architecture**
- Core business logic is independent
- External dependencies are adapters
- Easy to test and swap implementations

### 2. **Single Responsibility**
- Each package has one clear purpose
- No circular dependencies
- Clean imports

### 3. **Explicit over Implicit**
- Clear package names
- Obvious responsibilities
- Easy to navigate

### 4. **Modular**
- Packages can be used independently
- Minimal coupling
- Maximum cohesion

## Import Patterns

### From Application Code

```go
import (
    "github.com/Palaciodiego008/rebololang/pkg/rebolo"
    "github.com/Palaciodiego008/rebololang/pkg/rebolo/context"
    "github.com/Palaciodiego008/rebololang/pkg/rebolo/session"
)
```

### Within Framework

```go
// In pkg/rebolo/context/context.go
import (
    "github.com/Palaciodiego008/rebololang/pkg/rebolo/session"
    "github.com/Palaciodiego008/rebololang/pkg/rebolo/validation"
)
```

## Comparison with Other Frameworks

### Buffalo
```
buffalo/
├── actions/       # Controllers (like our app/controllers)
├── render/        # Rendering (like our adapters/renderer)
├── worker/        # Background jobs
└── middleware/    # Middleware
```

### ReboloLang
```
rebolo/
├── core/          # Pure business logic
├── ports/         # Interfaces
├── adapters/      # Implementations
├── context/       # Request helpers
├── session/       # Session management
├── validation/    # Form validation
├── middleware/    # Middleware system
└── testing/       # Testing utilities
```

## Benefits

1. **Clear Organization** - Easy to find code
2. **Testable** - Each package can be tested independently
3. **Maintainable** - Changes are localized
4. **Scalable** - Easy to add new packages
5. **Documented** - Package purpose is obvious

---

**This structure balances simplicity with scalability, making ReboloLang easy to learn but powerful enough for production.**


