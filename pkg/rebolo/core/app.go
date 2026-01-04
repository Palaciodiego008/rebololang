package core

import (
	"context"
	"net/http"
)

// App represents the core application
type App struct {
	config     Config
	router     Router
	database   Database
	renderer   Renderer
	middleware []Middleware
}

// Config interface for configuration
type Config interface {
	GetPort() string
	GetHost() string
	GetDatabaseDriver() string
	GetDatabaseURL() string
	GetDatabaseDebug() bool
	GetEnvironment() string
	IsHotReload() bool
}

// NamedRoute is a type alias for route naming support
// Implementations can return nil if route naming is not needed
type NamedRoute interface{}

// Router interface for HTTP routing
type Router interface {
	GET(path string, handler http.HandlerFunc) NamedRoute
	POST(path string, handler http.HandlerFunc) NamedRoute
	PUT(path string, handler http.HandlerFunc) NamedRoute
	DELETE(path string, handler http.HandlerFunc) NamedRoute
	Resource(path string, controller Controller)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Use(middleware Middleware)
}

// Database interface for data persistence
type Database interface {
	Connect(ctx context.Context) error
	Close() error
	Migrate(ctx context.Context) error
	Health() error
	DB() interface{} // Returns underlying database instance (*sql.DB)
}

// Renderer interface for template and JSON rendering
type Renderer interface {
	RenderHTML(w http.ResponseWriter, template string, data interface{}) error
	RenderJSON(w http.ResponseWriter, data interface{}) error
	RenderError(w http.ResponseWriter, message string, status int) error
}

// Controller interface for HTTP controllers
type Controller interface {
	Index(w http.ResponseWriter, r *http.Request)
	Show(w http.ResponseWriter, r *http.Request)
	New(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Edit(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

// Middleware type for HTTP middleware
type Middleware func(http.Handler) http.Handler

// NewApp creates a new application instance
func NewApp(config Config, router Router, database Database, renderer Renderer) *App {
	return &App{
		config:   config,
		router:   router,
		database: database,
		renderer: renderer,
	}
}

// Start starts the application server
func (a *App) Start() error {
	// Connect to database if configured
	if a.config.GetDatabaseURL() != "" {
		if err := a.database.Connect(context.Background()); err != nil {
			return err
		}
	}

	// Apply middleware - wrap the router with middleware in reverse order
	// (first middleware becomes outermost, last becomes innermost)
	var handler http.Handler = a.router
	for i := len(a.middleware) - 1; i >= 0; i-- {
		handler = a.middleware[i](handler)
	}

	port := a.config.GetPort()
	if port == "" {
		port = "3000"
	}

	return http.ListenAndServe(":"+port, handler)
}

// AddMiddleware adds middleware to the application
func (a *App) AddMiddleware(middleware Middleware) {
	a.middleware = append(a.middleware, middleware)
}

// Router returns the router instance
func (a *App) Router() Router {
	return a.router
}

// Database returns the database instance
func (a *App) Database() Database {
	return a.database
}

// Renderer returns the renderer instance
func (a *App) Renderer() Renderer {
	return a.renderer
}
