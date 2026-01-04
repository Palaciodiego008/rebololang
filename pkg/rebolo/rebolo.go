package rebolo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Palaciodiego008/rebololang/pkg/rebolo/adapters"
	rebolocontext "github.com/Palaciodiego008/rebololang/pkg/rebolo/context"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/core"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/errors"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/logging"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/middleware"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/ports"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/resource"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/routing"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/session"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/validation"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/watcher"
	"github.com/Palaciodiego008/rebololang/pkg/rebolo/worker"
)

// Application represents the main application facade
type Application struct {
	*core.App
	config          *ConfigAdapter
	router          *adapters.MuxRouter
	database        adapters.DatabaseAdapter
	renderer        *adapters.HTMLRenderer
	watcher         *watcher.FileWatcher
	sessionStore    *session.SessionStore       // Session management
	errorHandlers   errors.ErrorHandlers        // Custom error handlers
	middlewareStack *middleware.MiddlewareStack // Middleware stack with skip patterns
	worker          worker.Worker               // Background worker for jobs
	mu              sync.RWMutex                // For thread-safe template reloading
	ctx             context.Context
	cancelFunc      context.CancelFunc
	lastChangeTime  time.Time // Track last file change for polling
}

// ConfigAdapter adapts ports.ConfigData to core.Config
type ConfigAdapter struct {
	data ports.ConfigData
}

func (c *ConfigAdapter) GetPort() string           { return c.data.Server.Port }
func (c *ConfigAdapter) GetHost() string           { return c.data.Server.Host }
func (c *ConfigAdapter) GetDatabaseDriver() string { return c.data.Database.Driver }
func (c *ConfigAdapter) GetDatabaseURL() string    { return c.data.Database.URL }
func (c *ConfigAdapter) GetDatabaseDebug() bool    { return c.data.Database.Debug }
func (c *ConfigAdapter) GetEnvironment() string    { return c.data.App.Env }
func (c *ConfigAdapter) IsHotReload() bool         { return c.data.Assets.HotReload }

// New creates a new ReboloLang application
func New() *Application {
	// Load configuration
	configPort := adapters.NewYAMLConfig()
	configData, err := configPort.Load()
	if err != nil {
		log.Printf("Failed to load config: %v", err)
	}

	config := &ConfigAdapter{data: configData}
	router := adapters.NewMuxRouter()
	renderer := adapters.NewHTMLRenderer()

	// Create database adapter based on driver from config
	var database adapters.DatabaseAdapter
	if config.GetDatabaseURL() != "" {
		driver := config.GetDatabaseDriver()
		if driver == "" {
			driver = "postgres" // Default to postgres for backward compatibility
			log.Printf("⚠️  No database driver specified, defaulting to 'postgres'")
		}

		factory := adapters.NewDatabaseFactory()
		database, err = factory.CreateDatabase(driver)
		if err != nil {
			log.Printf("❌ Failed to create database adapter: %v", err)
			database = adapters.NewBunDatabase() // Fallback to postgres
		} else {
			// Connect to database
			debug := config.GetDatabaseDebug() || config.GetEnvironment() == "development"
			if err := database.ConnectWithDSN(config.GetDatabaseURL(), debug); err != nil {
				log.Printf("❌ Database connection failed: %v", err)
			} else {
				log.Printf("✅ Database connected successfully (driver: %s)", driver)
			}
		}
	} else {
		// No database configured, use a default instance
		database = adapters.NewBunDatabase()
	}

	// Create core app
	coreApp := core.NewApp(config, router, database, renderer)

	// Add default middleware
	coreApp.AddMiddleware(middleware.MethodOverride)
	coreApp.AddMiddleware(LoggingMiddleware)
	coreApp.AddMiddleware(RecoveryMiddleware)

	ctx, cancel := context.WithCancel(context.Background())

	// Generate a random secret key for sessions in development
	// In production, this should come from environment variable
	secretKey := []byte("rebolo-secret-key-change-in-production")
	sessionStore := session.NewCookieSessionStore("rebolo_session", secretKey)

	// Create background worker
	bgWorker := worker.NewSimpleWithContext(ctx)

	app := &Application{
		App:             coreApp,
		config:          config,
		router:          router,
		database:        database,
		renderer:        renderer,
		sessionStore:    sessionStore,
		errorHandlers:   errors.NewErrorHandlers(),
		middlewareStack: middleware.NewMiddlewareStack(),
		worker:          bgWorker,
		ctx:             ctx,
		cancelFunc:      cancel,
	}

	// Set custom error handlers on router
	router.Router.NotFoundHandler = app.NotFoundHandler()
	router.Router.MethodNotAllowedHandler = app.MethodNotAllowedHandler()

	return app
}

// Start starts the application
func (a *Application) Start() error {
	port := a.config.GetPort()
	if port == "" {
		port = "3000"
	}

	// Start background worker
	if a.worker != nil {
		if err := a.worker.Start(a.ctx); err != nil {
			log.Printf("⚠️  Failed to start worker: %v", err)
		} else {
			log.Println("✅ Background worker started")
		}
	}

	fmt.Printf("🚀 ReboloLang server starting on port %s\n", port)
	return a.App.Start()
}

// Convenience methods for routing
func (a *Application) GET(path string, handler http.HandlerFunc) *routing.NamedRoute {
	nr := a.router.GET(path, handler)
	if nr == nil {
		return nil
	}
	return nr.(*routing.NamedRoute)
}

func (a *Application) POST(path string, handler http.HandlerFunc) *routing.NamedRoute {
	nr := a.router.POST(path, handler)
	if nr == nil {
		return nil
	}
	return nr.(*routing.NamedRoute)
}

func (a *Application) PUT(path string, handler http.HandlerFunc) *routing.NamedRoute {
	nr := a.router.PUT(path, handler)
	if nr == nil {
		return nil
	}
	return nr.(*routing.NamedRoute)
}

func (a *Application) DELETE(path string, handler http.HandlerFunc) *routing.NamedRoute {
	nr := a.router.DELETE(path, handler)
	if nr == nil {
		return nil
	}
	return nr.(*routing.NamedRoute)
}

// ServeStatic serves static files from a directory
func (a *Application) ServeStatic(prefix, dir string) {
	fs := http.FileServer(http.Dir(dir))
	a.router.PathPrefix(prefix).Handler(http.StripPrefix(prefix, fs))
}

// Resource registers a RESTful resource using the old Controller interface
func (a *Application) Resource(path string, controller core.Controller) {
	a.router.Resource(path, controller)
}

// ResourceWithContext registers a RESTful resource using the new Resource interface with Context
func (a *Application) ResourceWithContext(path string, res resource.Resource) {
	base := path

	// Convert Resource methods to http.HandlerFunc using ContextMiddleware
	a.GET(base, a.ContextMiddleware(func(ctx *rebolocontext.Context) error {
		return res.List(ctx)
	}))

	a.GET(base+"/{id}", a.ContextMiddleware(func(ctx *rebolocontext.Context) error {
		return res.Show(ctx)
	}))

	a.POST(base, a.ContextMiddleware(func(ctx *rebolocontext.Context) error {
		return res.Create(ctx)
	}))

	a.PUT(base+"/{id}", a.ContextMiddleware(func(ctx *rebolocontext.Context) error {
		return res.Update(ctx)
	}))

	a.DELETE(base+"/{id}", a.ContextMiddleware(func(ctx *rebolocontext.Context) error {
		return res.Destroy(ctx)
	}))
}

// createRenderer creates a new HTML renderer (used for hot reload)
func (a *Application) createRenderer() *adapters.HTMLRenderer {
	return adapters.NewHTMLRenderer()
}

// EnableHotReload enables file watching and hot reload for development
func (a *Application) EnableHotReload() error {
	// Create file watcher
	fw := watcher.NewFileWatcher(a, []string{"views", "src", "public", "controllers"})

	// Start watching
	if err := fw.Start(); err != nil {
		return fmt.Errorf("failed to start file watcher: %v", err)
	}

	a.watcher = fw

	// Add hot reload middleware FIRST to inject script into HTML
	a.AddMiddleware(middleware.HotReloadMiddleware(true, "/__rebolo__/changes"))

	// Register polling endpoint for checking changes
	a.GET("/__rebolo__/changes", a.hotReloadChangesHandler)

	log.Printf("🔥 Hot reload enabled - watching files for changes")
	return nil
}

// hotReloadChangesHandler handles polling requests to check for file changes
func (a *Application) hotReloadChangesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Check if there are any pending changes
	// For simplicity, we'll use a timestamp-based approach
	a.mu.RLock()
	lastChange := a.lastChangeTime
	a.mu.RUnlock()

	response := map[string]interface{}{
		"changed": false,
		"time":    time.Now().Unix(),
	}

	// Check if there was a change in the last 2 seconds
	if time.Since(lastChange) < 2*time.Second {
		response["changed"] = true
		response["lastChange"] = lastChange.Unix()
	}

	// Use RenderJSON instead of global JSON() to avoid creating new renderer
	a.RenderJSON(w, response)
}

// GetSession retrieves the session for the current request
func (a *Application) GetSession(r *http.Request, w http.ResponseWriter) (*session.Session, error) {
	return a.sessionStore.Get(r, w)
}

// SetSessionStore allows custom session store configuration
func (a *Application) SetSessionStore(store *session.SessionStore) {
	a.sessionStore = store
}

// Shutdown gracefully shuts down the application
func (a *Application) Shutdown() {
	if a.watcher != nil {
		a.watcher.Close()
	}
	if a.worker != nil {
		a.worker.Stop()
	}
	if a.cancelFunc != nil {
		a.cancelFunc()
	}
}

// Convenience methods for rendering
func (a *Application) RenderHTML(w http.ResponseWriter, template string, data interface{}) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.renderer.RenderHTML(w, template, data)
}

func (a *Application) RenderJSON(w http.ResponseWriter, data interface{}) error {
	return a.renderer.RenderJSON(w, data)
}

func (a *Application) RenderError(w http.ResponseWriter, message string, status int) error {
	return a.renderer.RenderError(w, message, status)
}

// DB returns the underlying database/sql instance for convenience
func (a *Application) DB() *sql.DB {
	if a.database != nil {
		if db, ok := a.database.DB().(*sql.DB); ok {
			return db
		}
	}
	return nil
}

// LogQuery logs a SQL query in yellow (helper for controllers)
func (a *Application) LogQuery(query string, args ...interface{}) {
	if a.config.GetDatabaseDebug() || a.config.GetEnvironment() == "development" {
		logging.LogQuery(query, args...)
	}
}

// LogQueryError logs a SQL query error (helper for controllers)
func (a *Application) LogQueryError(query string, err error, args ...interface{}) {
	logging.LogQueryError(query, err, args...)
}

// responseWriter wraps http.ResponseWriter to capture status code and size
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK, 0}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(b)
	lrw.size += size
	return size, err
}

// Middleware
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip logging for hot reload polling endpoint to avoid spam
		if r.URL.Path == "/__rebolo__/changes" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		lrw := newLoggingResponseWriter(w)

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf("[%s] %s %s %d %d %v %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			lrw.statusCode,
			lrw.size,
			duration,
			r.UserAgent(),
		)
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Global convenience functions for backward compatibility
func Render(w http.ResponseWriter, template string, data interface{}) error {
	renderer := adapters.NewHTMLRenderer()
	return renderer.RenderHTML(w, template, data)
}

func JSON(w http.ResponseWriter, data interface{}) error {
	renderer := adapters.NewHTMLRenderer()
	return renderer.RenderJSON(w, data)
}

func JSONError(w http.ResponseWriter, message string, status int) error {
	renderer := adapters.NewHTMLRenderer()
	return renderer.RenderError(w, message, status)
}

// Use adds a middleware to the global stack
// Returns the MiddlewareConfig to allow chaining with Skip()
func (a *Application) Use(mw middleware.MiddlewareFunc) *middleware.MiddlewareConfig {
	return a.middlewareStack.Use(mw)
}

// Group creates a middleware group for specific routes
func (a *Application) Group(middlewares ...middleware.MiddlewareFunc) *middleware.MiddlewareGroup {
	group := middleware.NewMiddlewareGroup(a.middlewareStack)
	for _, mw := range middlewares {
		group.Use(mw)
	}
	return group
}

// UpdateLastChangeTime updates the last change time for hot reload
func (a *Application) UpdateLastChangeTime(t time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastChangeTime = t
}

// ReloadTemplates reloads HTML templates
func (a *Application) ReloadTemplates() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.renderer = a.createRenderer()
}

// Bind binds request data to a struct
func (a *Application) Bind(r *http.Request, v interface{}) error {
	return validation.Bind(r, v)
}

// BindAndValidate binds and validates in one step
func (a *Application) BindAndValidate(r *http.Request, v interface{}) error {
	return validation.BindAndValidate(r, v)
}

// SetErrorHandler sets a custom error handler for a status code
func (a *Application) SetErrorHandler(code int, handler errors.ErrorHandler) {
	if a.errorHandlers == nil {
		a.errorHandlers = errors.NewErrorHandlers()
	}
	a.errorHandlers[code] = handler
}

// HandleError handles an error with the appropriate error handler
func (a *Application) HandleError(w http.ResponseWriter, r *http.Request, err error, code int) {
	if a.errorHandlers == nil {
		a.errorHandlers = errors.NewErrorHandlers()
	}

	// Try to render custom error page from views/errors/{code}.html
	templatePath := fmt.Sprintf("errors/%d.html", code)
	a.mu.RLock()
	renderer := a.renderer
	a.mu.RUnlock()

	if renderer != nil {
		renderErr := renderer.RenderHTML(w, templatePath, map[string]interface{}{
			"Code":  code,
			"Error": err,
			"Path":  r.URL.Path,
		})
		if renderErr == nil {
			return
		}
	}

	// Use custom handler if available
	if handler, ok := a.errorHandlers[code]; ok {
		handler(w, r, err, code)
		return
	}

	// Fallback to standard error
	http.Error(w, fmt.Sprintf("Error %d", code), code)
}

// NotFoundHandler is a custom 404 handler
func (a *Application) NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.HandleError(w, r, fmt.Errorf("page not found: %s", r.URL.Path), 404)
	}
}

// MethodNotAllowedHandler is a custom 405 handler
func (a *Application) MethodNotAllowedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s[ERROR 405]%s Method Not Allowed: %s %s (if using PUT/DELETE, ensure form has _method field)",
			"\033[31m", "\033[0m", r.Method, r.URL.Path)
		a.HandleError(w, r, fmt.Errorf("method not allowed: %s %s (if using PUT/DELETE, ensure form has _method field)", r.Method, r.URL.Path), 405)
	}
}

// InternalErrorHandler handles 500 errors
func (a *Application) InternalErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("❌ Internal Server Error: %v", err)
	a.HandleError(w, r, err, 500)
}

// Worker methods

// RegisterWorker registers a handler for background jobs
func (a *Application) RegisterWorker(name string, handler worker.Handler) error {
	if a.worker == nil {
		return fmt.Errorf("worker not initialized")
	}
	return a.worker.Register(name, handler)
}

// Perform enqueues a job to be performed as soon as possible
func (a *Application) Perform(job worker.Job) error {
	if a.worker == nil {
		return fmt.Errorf("worker not initialized")
	}
	return a.worker.Perform(job)
}

// PerformAt enqueues a job to be performed at a specific time
func (a *Application) PerformAt(job worker.Job, t time.Time) error {
	if a.worker == nil {
		return fmt.Errorf("worker not initialized")
	}
	return a.worker.PerformAt(job, t)
}

// PerformIn enqueues a job to be performed after a duration
func (a *Application) PerformIn(job worker.Job, d time.Duration) error {
	if a.worker == nil {
		return fmt.Errorf("worker not initialized")
	}
	return a.worker.PerformIn(job, d)
}

// URLFor generates a URL for a named route with the given parameters
func (a *Application) URLFor(name string, params map[string]string) (string, error) {
	return routing.URLFor(a.router.Router, name, params)
}

// URLForString is a convenience function that returns the URL as a string
// or returns an empty string if there's an error
func (a *Application) URLForString(name string, params map[string]string) string {
	return routing.URLForString(a.router.Router, name, params)
}
